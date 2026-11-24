package tests

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AssemblerKonfig/src/cli"
)

// TestCLISpecification verifies the exact test-mode output against a separately stored reference.
func TestCLISpecification(t *testing.T) {
	var stdout, stderr bytes.Buffer
	output := filepath.Join(t.TempDir(), "program.bin")
	arguments := []string{"-input", "../examples/specification.asm", "-output", output, "-test"}
	code := cli.Run(arguments, &stdout, &stderr)
	want, err := os.ReadFile("specification.bytes.expected")
	if err != nil {
		t.Fatal(err)
	}
	if code != cli.Success || stderr.Len() != 0 || stdout.String() != "Размер двоичного файла: 16 байт.\n"+string(want) {
		t.Fatalf("exit=%d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	assertSpecificationBinary(t, output)
}

// TestCLIReplacesOutput checks normal-mode size reporting and replacement of an existing binary.
func TestCLIReplacesOutput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "existing.bin")
	const sentinel = "existing binary content"
	if err := os.WriteFile(output, []byte(sentinel), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"-input", "../examples/specification.asm", "-output", output}, &stdout, &stderr)
	if code != cli.Success || stderr.Len() != 0 || stdout.String() != "Размер двоичного файла: 16 байт.\n" {
		t.Fatalf("exit=%d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	assertSpecificationBinary(t, output)
}

// TestCLIInvalidArguments rejects missing paths, extra arguments, and invalid flags.
func TestCLIInvalidArguments(t *testing.T) {
	cases := [][]string{
		{}, {"-input", "source.asm"}, {"-output", "program.bin"}, {"-unknown"},
		{"-input"}, {"-test=invalid"}, {"-input", " ", "-output", "program.bin"},
		{"-input", "source.asm", "-output", "program.bin", "extra"},
		{"-input", "source.asm", "-output", ""},
	}
	for _, arguments := range cases {
		t.Run(strings.Join(arguments, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(arguments, &stdout, &stderr)
			if code != cli.Failure || stdout.Len() != 0 || stderr.Len() == 0 {
				t.Fatalf("exit=%d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

// TestCLIHelp checks successful help without requiring source and output paths.
func TestCLIHelp(t *testing.T) {
	for _, argument := range []string{"-h", "-help"} {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{argument}, &stdout, &stderr)
		if code != cli.Success || stderr.Len() != 0 || !strings.Contains(stdout.String(), "-input") {
			t.Fatalf("exit=%d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
		}
	}
}

// TestCLIInputFailures checks missing files and syntax errors without partial test output.
func TestCLIInputFailures(t *testing.T) {
	input := filepath.Join(t.TempDir(), "invalid program.asm")
	checkInputFailure(t, input)
	if err := os.WriteFile(input, []byte("stack = 1\ninvalid"), 0600); err != nil {
		t.Fatal(err)
	}
	checkInputFailure(t, input)
}

// checkInputFailure asserts a failed invocation with diagnostics only on stderr.
func checkInputFailure(t *testing.T, input string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"-input", input, "-output", "unused.bin", "-test"}, &stdout, &stderr)
	if code != cli.Failure || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("exit=%d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

// failingWriter simulates an unavailable output stream.
type failingWriter struct{}

// Write always returns an output failure.
func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failure")
}

// TestCLIOutputFailure requires a nonzero exit when the diagnostic output cannot be written.
func TestCLIOutputFailure(t *testing.T) {
	for _, mode := range []string{"-test=true", "-test=false"} {
		var stderr bytes.Buffer
		output := filepath.Join(t.TempDir(), "program.bin")
		arguments := []string{"-input", "../examples/specification.asm", "-output", output, mode}
		code := cli.Run(arguments, failingWriter{}, &stderr)
		if code != cli.Failure || !strings.Contains(stderr.String(), "write failure") {
			t.Fatalf("exit=%d, stderr=%q", code, stderr.String())
		}
	}
}

// TestCLIHelpOutputFailure verifies that an unwritable help stream produces a failure exit.
func TestCLIHelpOutputFailure(t *testing.T) {
	var stderr bytes.Buffer
	code := cli.Run([]string{"-help"}, failingWriter{}, &stderr)
	if code != cli.Failure || !strings.Contains(stderr.String(), "write failure") {
		t.Fatalf("exit=%d, stderr=%q", code, stderr.String())
	}
}
