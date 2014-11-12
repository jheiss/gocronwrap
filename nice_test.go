package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Ensure that --nice requires an argument
func TestNiceRequiresArg(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--nice").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
}

// Ensure that the argument must be an integer
func TestNiceArgIsInt(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--nice", "bogus", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--nice", "1.0", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--nice", "0", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--nice", "1", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
}

// Running as a regular user requesting a negative priority should fail
func TestNiceNegativePriority(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--nice", "-1", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
}

// Priority 0 should work
func TestNicePriorityZero(t *testing.T) {
	testprio := 0
	priority, err := VerifyPriority(testprio)
	if err != nil {
		t.Fail()
	}
	if priority != testprio {
		t.Error(fmt.Sprintf("priority %d expected, was %d", testprio, priority))
	}
}

// As should any positive integer up to around 19 or 20 depending on the
// operating system
func TestNicePriorityOne(t *testing.T) {
	testprio := 1
	priority, err := VerifyPriority(testprio)
	if err != nil {
		t.Fail()
	}
	if priority != testprio {
		t.Error(fmt.Sprintf("priority %d expected, was %d", testprio, priority))
	}
}
func TestNicePriorityTen(t *testing.T) {
	testprio := 10
	priority, err := VerifyPriority(testprio)
	if err != nil {
		t.Fail()
	}
	if priority != testprio {
		t.Error(fmt.Sprintf("priority %d expected, was %d", testprio, priority))
	}
}

func VerifyPriority(testprio int) (prioseen int, err error) {
	cmd := exec.Command("go", "run", "cronwrap.go", "--nice", strconv.Itoa(testprio), "sleep", "30")
	cmd.Start()
	time.Sleep(time.Duration(1) * time.Second) // Give the process time to start

	// go run wraps the command it executes, so the PID we have is just the go run
	// wrapper.  We need its child PID to check priority.  This is horrible but the
	// best I'm coming up with at the moment.  go test is compiling our
	// executable in a directory $TMPDIR that we could run directly, but I haven't
	// figured out how to find that temp directory from within test code.
	var childpid int
	out, _ := exec.Command("ps", "-ef").CombinedOutput()
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		columns := strings.Fields(line)
		if len(columns) >= 3 && columns[2] == strconv.Itoa(cmd.Process.Pid) {
			_, _ = fmt.Sscanf(columns[1], "%d", &childpid)
		}
	}
	if childpid == 0 {
		return 0, errors.New("did not find child pid")
	}

	// prioseen, err = syscall.Getpriority(syscall.PRIO_PROCESS, cmd.Process.Pid)
	// fmt.Printf("Checking priority of %d, expecting %d, got %d\n", cmd.Process.Pid, testprio, prioseen)
	prioseen, err = syscall.Getpriority(syscall.PRIO_PROCESS, childpid)
	if err != nil {
		return 0, err
	}
	err = cmd.Wait()
	return prioseen, err
}
