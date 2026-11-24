package tests

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"AssemblerKonfig/src/assembler"
)

// TestSpecification compares all fields against the four variant 17 specification examples.
func TestSpecification(t *testing.T) {
	source, err := os.Open("../examples/specification.asm")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	got, err := assembler.Parse(source)
	if err != nil {
		t.Fatal(err)
	}
	want := []assembler.Instruction{
		{A: assembler.LoadConstant, B: 452},
		{A: assembler.ReadMemory},
		{A: assembler.WriteMemory, B: 100},
		{A: assembler.Maximum, B: 170},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields = %#v, want %#v", got, want)
	}
}

// TestOperandBoundaries verifies unsigned field widths and zero operands for every parameterized command.
func TestOperandBoundaries(t *testing.T) {
	cases := []struct {
		source string
		want   assembler.Instruction
	}{
		{"stack = 0", assembler.Instruction{A: assembler.LoadConstant}},
		{"stack = 536870911", assembler.Instruction{A: assembler.LoadConstant, B: assembler.MaxConstant}},
		{"memory[pop() + 0] = pop()", assembler.Instruction{A: assembler.WriteMemory}},
		{"memory[pop() + 255] = pop()", assembler.Instruction{A: assembler.WriteMemory, B: assembler.MaxOffset}},
		{"stack = max(memory[pop()], memory[pop() + 0])", assembler.Instruction{A: assembler.Maximum}},
		{"stack = max(memory[pop()], memory[pop() + 255])", assembler.Instruction{A: assembler.Maximum, B: 255}},
	}
	for _, test := range cases {
		t.Run(test.source, func(t *testing.T) {
			got, err := assembler.Parse(strings.NewReader(test.source))
			if err != nil || len(got) != 1 {
				t.Fatalf("parse = %v, %v", got, err)
			}
			if got[0] != test.want {
				t.Fatalf("fields = %v, want %v", got[0], test.want)
			}
		})
	}
}

// TestInvalidPrograms rejects malformed expressions, unsupported syntax, and overflowing fields.
func TestInvalidPrograms(t *testing.T) {
	cases := []string{
		"stack = 536870912", "stack = 4294967296", "stack = 999999999999999999999999",
		"stack = -1", "stack = +1", "stack = 1.5", "stack = 0x10", "stack = 1 2",
		"memory[pop() + 256] = pop()", "memory[pop() + -1] = pop()",
		"stack = max(memory[pop()], memory[pop() + 256])",
		"stack = max(memory[pop()], memory[pop() + -1])",
		"stack = memory[pop(1)]", "stack = memory[0]", "stack = memory[p o p()]",
		"stack = 1; stack = 2", "stack = 1;;", "stack = 1 trailing", "stack =",
		"push 452", "STACK = 452", "stack = min(memory[pop()], memory[pop() + 1])",
		"stack = max(memory[pop() + 1], memory[pop()])", "unknown", ";",
	}
	for _, source := range cases {
		t.Run(source, func(t *testing.T) {
			got, err := assembler.Parse(strings.NewReader("// comment\nstack = 1\n" + source))
			if err == nil || !strings.Contains(err.Error(), "строка 3:") {
				t.Fatalf("expected line 3 error, got %v", err)
			}
			if got != nil {
				t.Fatalf("partial program returned on failure: %v", got)
			}
		})
	}
}

// TestWhitespace accepts spaces, tabs, CRLF, comments, and a final line without a newline.
func TestWhitespace(t *testing.T) {
	source := "\t// comment\r\n\r\n stack\t= 00452 ; // load\r\n" +
		"stack = memory [ pop ( ) ]\r\n" +
		"memory [ pop ( ) + 100 ] = pop ( ) ;\r\n" +
		"stack=max ( memory [pop()], memory[ pop() + 170 ] )"
	got, err := assembler.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	want := []assembler.Instruction{{A: 1, B: 452}, {A: 7}, {A: 5, B: 100}, {A: 2, B: 170}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fields = %v, want %v", got, want)
	}
}

// TestEmptyProgram accepts an empty source or one containing only whitespace and comments.
func TestEmptyProgram(t *testing.T) {
	for _, source := range []string{"", "\n\t\r\n// comment"} {
		got, err := assembler.Parse(strings.NewReader(source))
		if err != nil || len(got) != 0 {
			t.Fatalf("empty program = %v, %v", got, err)
		}
	}
}

// failingReader simulates an input device failure.
type failingReader struct{}

// Read always returns an input failure.
func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read failure")
}

// TestReadFailure checks that input failures are reported with their source line.
func TestReadFailure(t *testing.T) {
	got, err := assembler.Parse(failingReader{})
	if got != nil || err == nil || !strings.Contains(err.Error(), "строка 1: чтение:") {
		t.Fatalf("expected read failure, got %v, %v", got, err)
	}
}
