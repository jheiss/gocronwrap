package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

var jitter time.Duration
var overlap bool
var nice int
var timeout time.Duration
var suppress int
var debug bool
var version bool

func main() {
	const ver = "0.0.1"

	//
	// Parse Flags
	//

	flag.DurationVar(&jitter, "jitter", 0, "Random delay before executing job")
	flag.BoolVar(&overlap, "overlap", false, "Prevent multiple simultaneous copies of job")
	flag.IntVar(&nice, "nice", 0, "Set process priority, a la the utility nice")
	flag.DurationVar(&timeout, "timeout", 0, "Terminate job if it runs longer than given time")
	flag.IntVar(&suppress, "suppress", 0, "Suppress errors unless job has N consecutive failures")
	flag.BoolVar(&debug, "debug", false, "Print lots of messages about what cronwrap is doing")
	flag.BoolVar(&version, "version", false, "Print cronwrap version and exit")
	flag.Parse()

	// flag.Args() has all of the remaining command line arguments, which will be
	// the job command and its arguments
	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Error, must specify a command\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if suppress < 0 {
		fmt.Fprintf(os.Stderr, "Error: suppress should be a positive integer\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if version {
		fmt.Printf("cronwrap version %s\n", ver)
		os.Exit(0)
	}

	//
	// Prep work
	//

	// This value for workdir is open to debate.  Using a system directory like
	// /var/lib/cronwrap would restrict cronwrap to use by root, which doesn't seem
	// desirable.  Using $TMPDIR or other world writable, sticky bit enabled
	// directories makes it hard for us to come up with a way to identify what
	// directory cronwrap should use, given that we can't guarantee any specific
	// filename will be available. I.e. we can't assume /tmp/cronwrap will be
	// available for us to use.  If one instance uses mktemp and creates
	// /tmp/cronwrap.45e2f7 how is any other instance to know that's valid?  And we
	// don't want to lose state to cleanup from tmpwatch or system reboots.  Using
	// $HOME for variable/temporary state data isn't ideal, but it's the best I'm
	// coming up with at the moment.
	workdir := path.Join(os.Getenv("HOME"), ".cronwrap")
	err := os.MkdirAll(workdir, 0755)
	check(err)

	cmdAsString := fmt.Sprintf("%q", flag.Args())
	cmdsha1bytes := sha1.Sum([]byte(cmdAsString))
	cmdsha1 := fmt.Sprintf("%x", cmdsha1bytes)
	if debug {
		fmt.Printf("Command SHA1: %s\n", cmdsha1)
	}

	jobdir := path.Join(workdir, cmdsha1)
	err = os.MkdirAll(jobdir, 0755)
	check(err)

	// Write out a file with the command line to make it easier for users to figure
	// out which job is associated with a directory in our working space.  A
	// directory full of SHA1 sums isn't very user friendly.
	cmdfile := path.Join(jobdir, "command")
	_, err = os.Stat(cmdfile)
	if err != nil {
		file, err := os.Create(cmdfile)
		check(err)
		_, err = file.WriteString(cmdAsString)
		check(err)
		err = file.Close()
		check(err)
	}

	//
	// Jitter
	//

	if jitter.Seconds() != 0 {
		// Seed the random number generator with the hostname of this box so that
		// we get a consistent random number.  We want to run the job at a
		// consistent time on each individual box, but randomize the runs across
		// the environment.
		hostname, err := os.Hostname()
		check(err)
		hostsha1bytes := sha1.Sum([]byte(hostname))
		seed := new(big.Int)
		seed.SetBytes(hostsha1bytes[0:8])
		rand.Seed(seed.Int64())

		delay := time.Duration(rand.Int63n(int64(jitter.Seconds())) * int64(time.Second))
		if debug {
			fmt.Printf("Jitter delay of %d seconds\n", int64(delay.Seconds()))
		}
		time.Sleep(delay)
	}

	//
	// Overlap
	//

	pidfilename := path.Join(jobdir, "pid")
	var pidfile *os.File
	if overlap {
		if debug {
			fmt.Printf("Overlap protection enabled, checking for existing process\n")
		}
		pidfile, err = os.OpenFile(pidfilename, os.O_WRONLY|os.O_CREATE, 0644)
		check(err)
		if debug {
			fmt.Printf("Attempting to lock PID file: %s\n", pidfilename)
		}
		err = syscall.Flock(int(pidfile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Job is already running\n")
			os.Exit(1)
		}
		if debug {
			fmt.Printf("Locked PID file: %s\n", pidfilename)
		}
		_, err = pidfile.WriteString(fmt.Sprintf("%d", os.Getpid()))
		check(err)
	}

	//
	// Priority
	//
	// Note this is specifically placed after overlap protection.  If running on a
	// system with a higher priority process consuming 100% of CPU this process
	// will not make any forward progress once it drops its priority.  As such, if
	// we dropped priority before performing overlap protection we'd build up
	// instances of this process all waiting for a chance to run.
	//

	if nice != 0 {
		if debug {
			fmt.Printf("Setting priority to %d\n", nice)
		}
		err = syscall.Setpriority(syscall.PRIO_PROCESS, 0, nice)
		check(err)
	}

	//
	// Spawn the job
	//

	// Run the job in a goroutine
	type CombinedOutput struct {
		output    []byte
		exitvalue int
	}
	cmdch := make(chan *exec.Cmd, 1)
	resch := make(chan CombinedOutput, 1)
	go func() {
		if debug {
			fmt.Printf("Spawning job\n")
		}
		cmd := exec.Command(flag.Args()[0], flag.Args()[1:]...)
		// Stash the command in a channel in case we need in later to handle
		// a timeout
		cmdch <- cmd
		output, err := cmd.CombinedOutput()
		exitvalue := 0
		if err != nil {
			// This involves some Go magic I don't yet understand
			// http://stackoverflow.com/questions/10385551/get-exit-code-go
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					exitvalue = status.ExitStatus()
				}
			}
		}
		if debug {
			fmt.Printf("Job exited with status %d\n", exitvalue)
			fmt.Printf("Captured %d characters of output from job\n", utf8.RuneCountInString(string(output)))
		}
		resch <- CombinedOutput{output, exitvalue}
	}()
	// Start another goroutine to signal a timeout
	timech := make(chan bool, 1)
	if timeout.Seconds() != 0 {
		go func() {
			time.Sleep(timeout)
			timech <- true
		}()
	}

	// Read from the channels, select will give us whichever one returns first
	var output []byte
	var exitvalue int
	select {
	case combinedOutput := <-resch:
		output = combinedOutput.output
		exitvalue = combinedOutput.exitvalue
	case <-timech:
		if debug {
			fmt.Printf("Process timed out, sending SIGTERM\n")
		}
		cmd := <-cmdch
		// cmd.Process.Kill() sends SIGKILL
		// We want to try to be more graceful so we start with SIGTERM
		needskill := true
		_ = cmd.Process.Signal(syscall.SIGTERM)
		for i := 0; i < 5; i++ {
			var waitstat syscall.WaitStatus
			waitpid, _ := syscall.Wait4(cmd.Process.Pid, &waitstat, syscall.WNOHANG, nil)
			if waitpid != 0 {
				needskill = false
				break
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
		if needskill {
			if debug {
				fmt.Printf("Process did not die, sending SIGKILL\n")
			}
			_ = cmd.Process.Kill()
		}
		if debug {
			fmt.Printf("Process timed out, terminated\n")
		}
		exitvalue = 1
	}

	if overlap {
		if debug {
			fmt.Printf("Removing PID file\n")
		}
		err = pidfile.Close()
		check(err)
		err = os.Remove(pidfilename)
		check(err)
	}

	//
	// Failure suppression
	//

	failcountfilename := path.Join(jobdir, "failcount")
	suppress_failure := false
	var failcount int
	if exitvalue == 0 {
		failcount = 0
		if suppress != 0 {
			if debug {
				fmt.Printf("Suppressing output\n")
			}
			suppress_failure = true
		}
	} else {
		// Get existing failcount for this job and increment by one
		oldcount := 0
		if debug {
			fmt.Printf("Reading old failure count\n")
		}
		oldcountbytes, err := ioutil.ReadFile(failcountfilename)
		if err == nil {
			oldcountstring := string(oldcountbytes)
			_, _ = fmt.Sscanf(strings.TrimSpace(oldcountstring), "%d", &oldcount)
			if debug {
				fmt.Printf("Old failure count is %d\n", oldcount)
			}
		}
		failcount = oldcount + 1
		if debug {
			fmt.Printf("Failure count for this job is %d\n", failcount)
		}
		if suppress != 0 && failcount < suppress {
			if debug {
				fmt.Printf("Suppressing output\n")
			}
			suppress_failure = true
		}
	}

	if debug {
		fmt.Printf("Saving failure count for this job\n")
	}
	file, err := os.Create(failcountfilename)
	check(err)
	_, err = file.WriteString(fmt.Sprintf("%d", failcount))
	check(err)
	err = file.Close()
	check(err)

	if suppress_failure {
		os.Exit(0)
	} else {
		_, _ = os.Stdout.Write(output)
		os.Exit(exitvalue)
	}
}

func check(e error) {
	if e != nil {
		if debug {
			panic(e)
		} else {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
	}
}
