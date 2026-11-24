package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AssemblerKonfig/src/cli"
)

// TestCLIEmptyBinary verifies that both modes create an empty file and report zero bytes.
func TestCLIEmptyBinary(t *testing.T) {
	for _, mode := range []string{"-test=true", "-test=false"} {
		directory := t.TempDir()
		input, output := filepath.Join(directory, "empty.asm"), filepath.Join(directory, "empty.bin")
		writeFixture(t, input, "// empty program\n")
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{"-input", input, "-output", output, mode}, &stdout, &stderr)
		if code != cli.Success || stderr.Len() != 0 || stdout.String() != "Размер двоичного файла: 0 байт.\n" {
			t.Fatalf("exit=%d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
		}
		content, err := os.ReadFile(output)
		if err != nil || len(content) != 0 {
			t.Fatalf("empty binary=% X, error=%v", content, err)
		}
	}
}

// TestCLIInvalidSourcePreservesBinary ensures no partial output or truncation on syntax or range errors.
func TestCLIInvalidSourcePreservesBinary(t *testing.T) {
	for _, invalid := range []string{"invalid", "stack = 536870912", "memory[pop() + 256] = pop()"} {
		directory := t.TempDir()
		input, output := filepath.Join(directory, "invalid.asm"), filepath.Join(directory, "existing.bin")
		writeFixture(t, input, "stack = 452\n"+invalid)
		writeFixture(t, output, "original binary")
		assertCLIFailure(t, input, output)
		assertFileText(t, output, "original binary")
	}
}

// TestCLIBinaryWriteFailures verifies nonexistent directories and failed rename cleanup.
func TestCLIBinaryWriteFailures(t *testing.T) {
	directory := t.TempDir()
	outputs := []string{filepath.Join(directory, "missing", "program.bin"), directory}
	for _, output := range outputs {
		assertCLIFailure(t, "../examples/specification.asm", output)
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary files left behind: %v, error=%v", entries, err)
	}
}

// TestCLISourceAliases rejects identical paths, symbolic links, and hard links to the source.
func TestCLISourceAliases(t *testing.T) {
	directory := t.TempDir()
	input := filepath.Join(directory, "source.asm")
	const source = "stack = 452\n"
	writeFixture(t, input, source)
	for _, link := range []struct {
		name   string
		create func(string, string) error
	}{{"symbolic", os.Symlink}, {"hard", os.Link}} {
		alias := filepath.Join(directory, link.name)
		if err := link.create(input, alias); err != nil {
			t.Fatal(err)
		}
		assertCLIFailure(t, input, alias)
		assertFileText(t, input, source)
	}
	assertCLIFailure(t, input, input)
	assertFileText(t, input, source)
}

// TestCLIByteOutputFailure checks a stream failure after the size line was successfully written.
func TestCLIByteOutputFailure(t *testing.T) {
	output := filepath.Join(t.TempDir(), "program.bin")
	writer := &failAfterSizeWriter{}
	var stderr bytes.Buffer
	arguments := []string{"-input", "../examples/specification.asm", "-output", output, "-test"}
	if code := cli.Run(arguments, writer, &stderr); code != cli.Failure {
		t.Fatalf("exit=%d, stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "вывод машинного кода:") {
		t.Fatalf("stderr=%q", stderr.String())
	}
	assertSpecificationBinary(t, output)
}

// failAfterSizeWriter accepts the first write, then simulates a broken diagnostic stream.
type failAfterSizeWriter struct{ wroteSize bool }

// Write succeeds for the size line and fails on subsequent byte output.
func (writer *failAfterSizeWriter) Write(data []byte) (int, error) {
	if writer.wroteSize {
		return failingWriter{}.Write(data)
	}
	writer.wroteSize = true
	return len(data), nil
}

// assertCLIFailure requires a failed invocation with no success output.
func assertCLIFailure(t *testing.T, input, output string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"-input", input, "-output", output, "-test"}, &stdout, &stderr)
	if code != cli.Failure || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("exit=%d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
}

// writeFixture stores test input in a caller-owned temporary directory.
func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// assertFileText checks that a file was preserved after an unsuccessful invocation.
func assertFileText(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("content=%q, error=%v; want %q", got, err, want)
	}
}
