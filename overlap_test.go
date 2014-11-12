package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Ensure job runs when no overlap condition exists
func TestNoOverlap(t *testing.T) {
	file, err := ioutil.TempFile("", "cronwrap")
	if err != nil {
		t.Error("tempfile")
	}

	out, err := exec.Command("go", "run", "cronwrap.go", "--overlap", "--", "sh", "-c", "printf test > "+file.Name()).CombinedOutput()
	if err != nil {
		t.Error(string(out))
	}

	bytes, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Error("readfile " + file.Name())
	}
	contents := string(bytes)
	if contents != "test" {
		t.Error("expected 'test', was '" + contents + "'")
	}

	err = file.Close()
	if err != nil {
		t.Error("close " + file.Name())
	}
	err = os.Remove(file.Name())
	if err != nil {
		t.Error("remove " + file.Name())
	}
}

// Ensure job does not run when an overlap condition exists
func TestOverlap(t *testing.T) {
	cmd := exec.Command("go", "run", "cronwrap.go", "--overlap", "sleep", "3")
	cmd.Start()
	time.Sleep(time.Duration(1) * time.Second) // Give the process time to start

	out, err := exec.Command("go", "run", "cronwrap.go", "--overlap", "sleep", "3").CombinedOutput()
	if err == nil {
		// t.Error("Overlap exit: " + string(err))
		t.Error("Overlap exit: ")
	}
	if !strings.Contains(string(out), "Job is already running") {
		t.Error("Overlap output: " + string(out))
	}
}
