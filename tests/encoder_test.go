package tests

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"AssemblerKonfig/src/assembler"
)

// specificationBinary contains all four independent byte sequences from pages 72 and 73 of the assignment.
var specificationBinary = []byte{
	0x21, 0x0E, 0x00, 0x00,
	0x07, 0x00, 0x00, 0x00,
	0x25, 0x03, 0x00, 0x00,
	0x52, 0x05, 0x00, 0x00,
}

// TestEncodeSpecification verifies the entire source-to-binary pipeline against the assignment bytes.
func TestEncodeSpecification(t *testing.T) {
	source, err := os.ReadFile("../examples/specification.asm")
	if err != nil {
		t.Fatal(err)
	}
	program, err := assembler.Parse(strings.NewReader(string(source)))
	if err != nil {
		t.Fatal(err)
	}
	got, err := assembler.Encode(program)
	if err != nil || !bytes.Equal(got, specificationBinary) {
		t.Fatalf("binary = % X, error = %v; want % X", got, err, specificationBinary)
	}
}

// TestEncodeBoundaries checks field packing, reserved zero bits, and byte order at both field boundaries.
func TestEncodeBoundaries(t *testing.T) {
	cases := []struct {
		instruction assembler.Instruction
		want        []byte
	}{
		{assembler.Instruction{A: assembler.LoadConstant}, []byte{0x01, 0x00, 0x00, 0x00}},
		{assembler.Instruction{A: assembler.LoadConstant, B: assembler.MaxConstant}, []byte{0xF9, 0xFF, 0xFF, 0xFF}},
		{assembler.Instruction{A: assembler.ReadMemory}, []byte{0x07, 0x00, 0x00, 0x00}},
		{assembler.Instruction{A: assembler.WriteMemory}, []byte{0x05, 0x00, 0x00, 0x00}},
		{assembler.Instruction{A: assembler.WriteMemory, B: assembler.MaxOffset}, []byte{0xFD, 0x07, 0x00, 0x00}},
		{assembler.Instruction{A: assembler.Maximum}, []byte{0x02, 0x00, 0x00, 0x00}},
		{assembler.Instruction{A: assembler.Maximum, B: assembler.MaxOffset}, []byte{0xFA, 0x07, 0x00, 0x00}},
	}
	for _, test := range cases {
		t.Run(test.instruction.String(), func(t *testing.T) {
			got, err := assembler.Encode([]assembler.Instruction{test.instruction})
			if err != nil || !bytes.Equal(got, test.want) {
				t.Fatalf("binary = % X, error = %v; want % X", got, err, test.want)
			}
		})
	}
}

// TestEncodeInvalidIR rejects invalid manually constructed instructions without returning partial output.
func TestEncodeInvalidIR(t *testing.T) {
	cases := []assembler.Instruction{
		{A: 0}, {A: 8}, {A: assembler.LoadConstant, B: assembler.MaxConstant + 1},
		{A: assembler.ReadMemory, B: 1}, {A: assembler.WriteMemory, B: assembler.MaxOffset + 1},
		{A: assembler.Maximum, B: assembler.MaxOffset + 1},
	}
	for _, instruction := range cases {
		program := []assembler.Instruction{{A: assembler.LoadConstant}, instruction}
		got, err := assembler.Encode(program)
		if got != nil || err == nil || !strings.Contains(err.Error(), "команда 2:") {
			t.Fatalf("instruction=%v, binary=% X, error=%v", instruction, got, err)
		}
	}
}

// TestEncodeEmpty accepts empty intermediate programs.
func TestEncodeEmpty(t *testing.T) {
	got, err := assembler.Encode(nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty binary=% X, error=%v", got, err)
	}
}

// assertSpecificationBinary verifies actual file contents, not just the CLI diagnostics.
func assertSpecificationBinary(t *testing.T, path string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, specificationBinary) {
		t.Fatalf("file=% X, error=%v; want % X", got, err, specificationBinary)
	}
}
