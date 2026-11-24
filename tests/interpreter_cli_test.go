package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"AssemblerKonfig/src/cli"
)

// TestCopyArrayCLI assembles and executes the required copy program and checks both arrays independently.
func TestCopyArrayCLI(t *testing.T) {
	directory := t.TempDir()
	binaryPath, dumpPath := filepath.Join(directory, "copy.bin"), filepath.Join(directory, "copy.csv")
	var stdout, stderr bytes.Buffer
	if cli.Run([]string{"-input", "../examples/copy_array.asm", "-output", binaryPath}, &stdout, &stderr) != 0 {
		t.Fatal(stderr.String())
	}
	for _, interval := range [][2]string{{"2000", "2004"}, {"1000", "1004"}} {
		stdout.Reset()
		args := []string{"-input", binaryPath, "-output", dumpPath, "-start", interval[0], "-end", interval[1]}
		if cli.RunInterpreter(args, &stdout, &stderr) != 0 || stderr.Len() != 0 || stdout.Len() != 0 {
			t.Fatalf("stdout=%q, stderr=%q", stdout.String(), stderr.String())
		}
		want := "address,value\n" + interval[0] + ",0\n"
		if interval[0] == "2000" {
			want += "2001,17\n2002,452\n2003,65535\n2004,536870911\n"
		} else {
			want += "1001,17\n1002,452\n1003,65535\n1004,536870911\n"
		}
		assertFileText(t, dumpPath, want)
	}
}

// TestInterpreterArguments checks help, required flags, strict parsing, and invalid intervals.
func TestInterpreterArguments(t *testing.T) {
	cases := [][]string{nil, {"-unknown"}, {"extra"}, {"-input", "x", "-output", "y"},
		{"-input", "x", "-output", "y", "-start", "1", "-end", "0"},
		{"-input", "x", "-output", "y", "-start", "0", "-end", "65536"},
		{"-input", "x", "-output", "y", "-start", "bad", "-end", "0"}}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		if cli.RunInterpreter(args, &stdout, &stderr) != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatalf("args=%v, stdout=%q, stderr=%q", args, stdout.String(), stderr.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if cli.RunInterpreter([]string{"-help"}, &stdout, &stderr) != 0 || stdout.Len() == 0 || stderr.Len() != 0 {
		t.Fatal("invalid help behavior")
	}
	if cli.RunInterpreter([]string{"-h"}, failingWriter{}, &stderr) != 1 {
		t.Fatal("help writer error ignored")
	}
}

// TestInterpreterPreservesFiles verifies runtime failures and output aliases cannot destroy the program or dump.
func TestInterpreterPreservesFiles(t *testing.T) {
	directory := t.TempDir()
	input, output := filepath.Join(directory, "program.bin"), filepath.Join(directory, "dump.csv")
	writeFixture(t, input, "\x07\x00\x00\x00")
	writeFixture(t, output, "old dump")
	assertInterpreterFailure(t, input, output)
	assertFileText(t, output, "old dump")
	writeFixture(t, input, "\x01\x00\x00\x00")
	assertInterpreterFailure(t, input, input)
	for _, link := range []struct {
		name   string
		create func(string, string) error
	}{
		{"symbolic", os.Symlink}, {"hard", os.Link},
	} {
		alias := filepath.Join(directory, link.name)
		if err := link.create(input, alias); err != nil {
			t.Fatal(err)
		}
		assertInterpreterFailure(t, input, alias)
	}
	assertFileText(t, input, "\x01\x00\x00\x00")
	assertInterpreterFailure(t, input, filepath.Join(directory, "missing", "dump.csv"))
	assertInterpreterFailure(t, input, directory)
	assertInterpreterFailure(t, filepath.Join(directory, "missing.bin"), output)
}

// assertInterpreterFailure requires a nonzero exit and a diagnostic without success output.
func assertInterpreterFailure(t *testing.T, input, output string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	args := []string{"-input", input, "-output", output, "-start", "0", "-end", "0"}
	if cli.RunInterpreter(args, &stdout, &stderr) != 1 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("stdout=%q, stderr=%q", stdout.String(), stderr.String())
	}
}
