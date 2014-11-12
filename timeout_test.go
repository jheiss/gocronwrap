package main

import (
	"os/exec"
	"strconv"
	"testing"
	"time"
)

// Ensure that --timeout requires an argument
func TestTimeoutRequiresArg(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--timeout").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
}

// Ensure that the argument must be a time delta
func TestTimeoutArgIsDelta(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--timeout", "bogus", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--timeout", "1r", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--timeout", "1", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}

	out, err = exec.Command("go", "run", "cronwrap.go", "--timeout", "0", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--timeout", "1s", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--timeout", "1m", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--timeout", "1h", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
}

// A job that runs less than the timeout should run to completion
func TestTimeoutCompletion(t *testing.T) {
	start := time.Now()
	out, err := exec.Command("go", "run", "cronwrap.go", "--timeout", "30s", "sleep", "5").CombinedOutput()
	end := time.Now()
	elapsed := end.Sub(start)
	if err != nil {
		t.Error(string(out))
	}
	// This builds in a few seconds of fudge factor for system load, etc.
	if elapsed.Seconds() < 5 || elapsed.Seconds() > 10 {
		t.Error("Expected elapsed 5<>10, was: " + strconv.Itoa(int(elapsed.Seconds())))
	}
}

// A job that runs longer than the timeout should be terminated
func TestTimeout(t *testing.T) {
	start := time.Now()
	out, err := exec.Command("go", "run", "cronwrap.go", "--timeout", "5s", "sleep", "30").CombinedOutput()
	end := time.Now()
	elapsed := end.Sub(start)
	if err == nil {
		t.Error(string(out))
	}
	// This builds in a few seconds of fudge factor for system load, etc.
	if elapsed.Seconds() < 5 || elapsed.Seconds() > 10 {
		t.Error("Expected elapsed 5<>10, was: " + strconv.Itoa(int(elapsed.Seconds())))
	}
}

// A job that ignores SIGTERM should be killed by SIGKILL
func TestTimeoutSigkill(t *testing.T) {
	start := time.Now()
	out, err := exec.Command("go", "run", "cronwrap.go", "--timeout", "5s", "./sigtermignore").CombinedOutput()
	end := time.Now()
	elapsed := end.Sub(start)
	if err == nil {
		t.Error(string(out))
	}
	// 5 seconds of timeout, plus cronwrap's 5 second wait before using SIGKILL
	if elapsed.Seconds() < 10 || elapsed.Seconds() > 15 {
		t.Error("Expected elapsed 10<>15, was: " + strconv.Itoa(int(elapsed.Seconds())))
	}
}
