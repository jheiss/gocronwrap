package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Ensure that --suppress requires an argument
func TestSuppressRequiresArg(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--suppress").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
}

// Ensure that the argument must be a non-negative integer
func TestSuppressArgIsInt(t *testing.T) {
	out, err := exec.Command("go", "run", "cronwrap.go", "--suppress", "bogus", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "1.0", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "-1", "true").CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "0", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "1", "true").CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
}

func TestSuppress(t *testing.T) {
	file, err := ioutil.TempFile("", "cronwrap")
	if err != nil {
		t.Error("file.Name()")
	}

	// Success not suppressed without --suppress
	file.Seek(0, os.SEEK_SET)
	file.WriteString("succeed\n")
	file.Sync()
	out, err := exec.Command("go", "run", "cronwrap.go", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if !strings.Contains(string(out), "tester is succeeding") {
		t.Error(string(out))
	}

	// Failure not suppressed without --suppress
	file.Seek(0, os.SEEK_SET)
	file.WriteString("fail\n")
	file.Sync()
	out, err = exec.Command("go", "run", "cronwrap.go", "./tester", file.Name()).CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	if !strings.Contains(string(out), "tester is failing") {
		t.Error(fmt.Sprintf("'%s'", string(out)))
	}

	// Have a successful run to reset the count
	file.Seek(0, os.SEEK_SET)
	file.WriteString("succeed\n")
	file.Sync()
	err = exec.Command("go", "run", "cronwrap.go", "./tester", file.Name()).Run()

	// Ensure that the appropriate number of failures are suppressed
	file.Seek(0, os.SEEK_SET)
	file.WriteString("fail\n")
	file.Sync()
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	if !strings.Contains(string(out), "tester is failing") {
		t.Error(string(out))
	}

	// Fail another time to ensure it stays unsuppressed
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	if !strings.Contains(string(out), "tester is failing") {
		t.Error(string(out))
	}

	// Have the command succeed to ensure that we reset to suppression on success
	file.Seek(0, os.SEEK_SET)
	file.WriteString("succeed\n")
	file.Sync()
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}

	// Switch back to failing and build up a couple of consecutive failures
	file.Seek(0, os.SEEK_SET)
	file.WriteString("fail\n")
	file.Sync()
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}

	// Have a successful run to reset the count
	file.Seek(0, os.SEEK_SET)
	file.WriteString("succeed\n")
	file.Sync()
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}

	// And now fail enough to trigger unsuppression
	file.Seek(0, os.SEEK_SET)
	file.WriteString("fail\n")
	file.Sync()
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}
	if string(out) != "" {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	if !strings.Contains(string(out), "tester is failing") {
		t.Error(string(out))
	}
	out, err = exec.Command("go", "run", "cronwrap.go", "--suppress", "3", "./tester", file.Name()).CombinedOutput()
	if err == nil {
		t.Error(string(out))
	}
	if !strings.Contains(string(out), "tester is failing") {
		t.Error(string(out))
	}
}
