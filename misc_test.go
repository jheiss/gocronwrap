package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"unicode/utf8"
)

// Test --help
func TestHelp(t *testing.T) {
	// --help exits with an error but we don't really care
	out, _ := exec.Command("go", "run", "cronwrap.go", "--help").CombinedOutput()
	lines := strings.Split(string(out), "\n")

	// Make sure at least something resembling help output is there
	if !strings.HasPrefix(lines[0], "Usage") {
		t.Error("Help message doesn't contain Usage")
	}

	// Make sure each line other than the Usage line fits in 80 characters.  The
	// Usage line contains the full path to the executable, which we don't control
	// and may cause the line to exceed 80 characters.
	for _, line := range lines {
		// utf8.RuneCountInString() is only somewhat correct.  A rune (or code point)
		// does not necessarily directly correspond to an on-screen character.
		// Combining characters can result in multiple runes mapping to a single
		// character.  The unicode/norm package might allow us to make this completely
		// correct but I'm punting for now.
		// http://godoc.org/code.google.com/p/go.text/unicode/norm
		// http://blog.golang.org/normalization
		if !strings.HasPrefix(line, "Usage") && utf8.RuneCountInString(line) > 80 {
			t.Error(fmt.Sprintf("Help line too long: %s", line))
		}
	}
	if len(lines) > 23 {
		t.Error("Too many help lines")
	}
}
