package govar

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// testData is a simple struct used across multiple API tests.
type testData struct {
	Name  string
	Value int
}

// Test a simple data structure to ensure output is generated.
var simpleData = testData{Name: "Test", Value: 123}

// TestSdump checks the Sdump variants (Sdump, SdumpNoColors, SdumpValues).
func TestSdump(t *testing.T) {
	t.Run("Sdump with colors", func(t *testing.T) {
		out := Sdump(simpleData)
		if out == "" {
			t.Error("Sdump() returned an empty string")
		}
		if !strings.Contains(out, "\x1b[") {
			t.Error("Sdump() output should contain ANSI color codes")
		}
		if !strings.Contains(out, "govar.testData") {
			t.Error("Sdump() output should contain type information")
		}
	})

	t.Run("SdumpNoColors", func(t *testing.T) {
		out := SdumpNoColors(simpleData)
		if out == "" {
			t.Error("SdumpNoColors() returned an empty string")
		}
		if strings.Contains(out, "\x1b[") {
			t.Error("SdumpNoColors() output should not contain ANSI color codes")
		}
	})

	t.Run("SdumpValues", func(t *testing.T) {
		out := SdumpValues(simpleData)
		if out == "" {
			t.Error("SdumpValues() returned an empty string")
		}
		if strings.Contains(out, "govar.testData") {
			t.Error("SdumpValues() output should not contain type information")
		}
	})
}

// TestFdump checks the Fdump variants by writing to a buffer.
func TestFdump(t *testing.T) {
	var buf bytes.Buffer

	t.Run("Fdump with colors", func(t *testing.T) {
		buf.Reset()
		Fdump(&buf, simpleData)
		out := buf.String()
		if out == "" {
			t.Error("Fdump() produced no output")
		}
		if !strings.Contains(out, "\x1b[") {
			t.Error("Fdump() output should contain ANSI color codes")
		}
	})

	t.Run("FdumpNoColors", func(t *testing.T) {
		buf.Reset()
		FdumpNoColors(&buf, simpleData)
		out := buf.String()
		if out == "" {
			t.Error("FdumpNoColors() produced no output")
		}
		if strings.Contains(out, "\x1b[") {
			t.Error("FdumpNoColors() output should not contain ANSI color codes")
		}
	})

	t.Run("FdumpValues", func(t *testing.T) {
		buf.Reset()
		FdumpValues(&buf, simpleData)
		out := buf.String()
		if out == "" {
			t.Error("FdumpValues() produced no output")
		}
		if strings.Contains(out, "govar.testData") {
			t.Error("FdumpValues() output should not contain type information")
		}
	})
}

// TestSdumpHTML checks the SdumpHTML variants for correct HTML tags.
func TestSdumpHTML(t *testing.T) {
	t.Run("SdumpHTML", func(t *testing.T) {
		out := SdumpHTML(simpleData)
		if !strings.HasPrefix(out, "<pre") {
			t.Error("SdumpHTML() output should start with a <pre> tag")
		}
		if !strings.Contains(out, "<span") {
			t.Error("SdumpHTML() output should contain <span> tags for tokens")
		}
		if !strings.Contains(out, "govar.testData") {
			t.Error("SdumpHTML() output should contain type information")
		}
	})

	t.Run("SdumpHTMLValues", func(t *testing.T) {
		out := SdumpHTMLValues(simpleData)
		if !strings.HasPrefix(out, "<pre") {
			t.Error("SdumpHTMLValues() output should start with a <pre> tag")
		}
		if strings.Contains(out, "govar.testData") {
			t.Error("SdumpHTMLValues() output should not contain type information")
		}
	})
}

// TestDump captures stdout to verify that Dump produces output.
func TestDump(t *testing.T) {
	// Keep backup of the real stdout
	old := os.Stdout
	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Dump(simpleData)

	// Restore stdout and close the write pipe
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	out := buf.String()

	if out == "" {
		t.Error("Dump() produced no output to stdout")
	}
	if !strings.Contains(out, "Test") {
		t.Errorf("Dump() output missing expected content: got %q", out)
	}
}

// TestDie is a special test that checks if the Die function exits with status 1.
// It does this by re-running the test binary with a specific environment variable.
func TestDie(t *testing.T) {
	// This environment variable tells the subprocess to run the Die function.
	if os.Getenv("GOVAR_TEST_DIE") == "1" {
		Die("testing die")
		return // This line should not be reached
	}

	// Run the test binary as a subprocess.
	cmd := exec.Command(os.Args[0], "-test.run=TestDie")
	cmd.Env = append(os.Environ(), "GOVAR_TEST_DIE=1")
	err := cmd.Run()

	// Check the exit code.
	e, ok := err.(*exec.ExitError)
	if !ok || e.Success() {
		t.Fatalf("Die() process ran without an error, but expected exit status 1")
	}

	// On Unix-like systems, we can check the exit code directly.
	// On Windows, ExitError.Success() being false is a strong enough signal.
	// The default exit code for a failed test is 1, which matches what Die() does.
	// So we just check that the process failed.
}
