package main

import (
	"fmt"
	"math"
	"os/exec"
	"testing"
	"time"
)

// Ensure that --jitter requires an argument
func TestJitterRequiresArg(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--jitter").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
}

// Ensure that the argument must be a time delta
func TestJitterArgIsDelta(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--jitter", "bogus", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--jitter", "1r", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--jitter", "1", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}

	out, err = exec.Command("go", "run", "cronwrap.go", "--jitter", "0", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--jitter", "1s", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	// out, err = exec.Command("go", "run", "cronwrap.go", "--jitter", "1m", "true").CombinedOutput()
	// if err != nil {
	//   t.Error(string(out))
	// }
	// out, err = exec.Command("go", "run", "cronwrap.go", "--jitter", "1h", "true").CombinedOutput()
	// if err != nil {
	//   t.Error(string(out))
	// }
}

// I actually don't know how to test that the job was delayed, as some machines
// are going to have a jitter value very close to zero.
//
// However, we do expect the jitter to be consistent on any given machine. That
// we can test.  One minute (--jitter 1m) is good enough for testing as cronwrap
// actually sleeps for a random number of seconds, so even with one minute of
// jitter cronwrap will sleep anywhere from 0-59 seconds.
func TestJitterIsConsistent(t *testing.T) {
	// FIXME: It would speed up the test suite considerably to run all three tests
	// simultaneously in goroutines and compare the results

	// Time one run
	start := time.Now()
	err := exec.Command("go", "run", "cronwrap.go", "--jitter", "1m", "true").Run()
	end := time.Now()
	elapsed := end.Sub(start)
	if err != nil {
		t.FailNow()
	}

	// Now verify that a few more runs are within one second of the same delay
	allowed_diff := 2
	for i := 1; i <= 2; i++ {
		start := time.Now()
		err := exec.Command("go", "run", "cronwrap.go", "--jitter", "1m", "true").Run()
		end := time.Now()
		test_elapsed := end.Sub(start)
		if err != nil {
			t.FailNow()
		}
		diff := math.Abs(test_elapsed.Seconds() - elapsed.Seconds())
		if diff > float64(allowed_diff) {
			t.Error(fmt.Sprintf("Expected diff <= %d, was %f", allowed_diff, diff))
		}
	}
}
