package tests

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"AssemblerKonfig/src/assembler"
	"AssemblerKonfig/src/vm"
)

// machineFromSource exercises parsing, encoding, loading, and execution with the production pipeline.
func machineFromSource(t *testing.T, source string) (*vm.Machine, error) {
	t.Helper()
	instructions, err := assembler.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	code, err := assembler.Encode(instructions)
	if err != nil {
		t.Fatal(err)
	}
	machine, err := vm.New(code)
	if err != nil {
		t.Fatal(err)
	}
	return machine, machine.Run()
}

// TestUnifiedMemory verifies reading code as data and fetching a modified future instruction.
func TestUnifiedMemory(t *testing.T) {
	cases := []struct{ source, want string }{
		{"stack = 100\nstack = 0\nstack = memory[pop()]\nmemory[pop()+0] = pop()",
			"address,value\n100,801\n"},
		{"stack = 4\nstack = 57\nmemory[pop()+0] = pop()\nstack = 100\n" +
			"stack = 1\nmemory[pop()+0] = pop()",
			"address,value\n100,7\n"},
	}
	for _, test := range cases {
		machine, err := machineFromSource(t, test.source)
		if err != nil {
			t.Fatal(err)
		}
		var dump bytes.Buffer
		if err := machine.Dump(&dump, 100, 100); err != nil || dump.String() != test.want {
			t.Fatalf("dump=%q, error=%v; want %q", dump.String(), err, test.want)
		}
	}
}

// TestVMRuntimeErrors checks stack underflow, address bounds, offset addition, and stage-four diagnostics.
func TestVMRuntimeErrors(t *testing.T) {
	cases := []struct{ source, message string }{
		{"stack = memory[pop()]", "недостаточно"},
		{"memory[pop()+0] = pop()", "недостаточно"},
		{"stack = 0\nmemory[pop()+0] = pop()", "недостаточно"},
		{"stack = 65536\nstack = memory[pop()]", "адрес чтения"},
		{"stack = 65535\nstack = 1\nmemory[pop()+1] = pop()", "адрес записи"},
		{"stack = max(memory[pop()], memory[pop()+0])", "этапу 4"},
	}
	for _, test := range cases {
		_, err := machineFromSource(t, test.source)
		if err == nil || !strings.Contains(err.Error(), test.message) || !strings.Contains(err.Error(), "команда") {
			t.Fatalf("source=%q, error=%v", test.source, err)
		}
	}
}

// TestVMFullWordData verifies full-width memory values and non-wrapping address addition.
func TestVMFullWordData(t *testing.T) {
	code := []byte{249, 255, 255, 255, 9, 0, 0, 0, 1, 0, 0, 0, 7, 0, 0, 0, 1, 0, 0, 0, 253, 7, 0, 0}
	machine, err := vm.New(code)
	if err != nil {
		t.Fatal(err)
	}
	if err := machine.Run(); err == nil || !strings.Contains(err.Error(), "4294967544") {
		t.Fatalf("expected checked uint64 address, got %v", err)
	}
	var dump bytes.Buffer
	if err := machine.Dump(&dump, 0, 0); err != nil || dump.String() != "address,value\n0,4294967289\n" {
		t.Fatalf("dump=%q, error=%v", dump.String(), err)
	}
}

// TestVMInvalidBinary checks truncation, excessive size, invalid opcodes, and reserved bits.
func TestVMInvalidBinary(t *testing.T) {
	for _, code := range [][]byte{{1}, make([]byte, (vm.MemorySize+1)*assembler.InstructionSize)} {
		if _, err := vm.New(code); err == nil {
			t.Fatal("invalid binary accepted")
		}
	}
	for _, word := range []uint32{0, 3, 4, 6, 15, 2053, 2050} {
		code := binary.LittleEndian.AppendUint32(nil, word)
		machine, err := vm.New(code)
		if err != nil {
			t.Fatal(err)
		}
		if err := machine.Run(); err == nil {
			t.Fatalf("invalid word %X accepted", word)
		}
	}
}

// TestVMDumpBounds checks zeroed memory, inclusive endpoints, invalid intervals, and writer failures.
func TestVMDumpBounds(t *testing.T) {
	machine, err := vm.New(nil)
	if err != nil || machine.Run() != nil {
		t.Fatalf("empty program: %v", err)
	}
	var dump bytes.Buffer
	if err := machine.Dump(&dump, vm.MemorySize-1, vm.MemorySize-1); err != nil {
		t.Fatal(err)
	}
	if dump.String() != "address,value\n65535,0\n" {
		t.Fatal(dump.String())
	}
	for _, interval := range [][2]int{{-1, 0}, {1, 0}, {0, vm.MemorySize}} {
		if err := machine.Dump(&dump, interval[0], interval[1]); err == nil {
			t.Fatal("invalid range accepted")
		}
	}
	if err := machine.Dump(failingWriter{}, 0, 0); err == nil {
		t.Fatal("writer failure ignored")
	}
}

// TestDecodeSpecification checks independent machine words and all operand boundaries against the shared IR.
func TestDecodeSpecification(t *testing.T) {
	for index, want := range []assembler.Instruction{{A: 1, B: 452}, {A: 7}, {A: 5, B: 100}, {A: 2, B: 170}} {
		word := binary.LittleEndian.Uint32(specificationBinary[index*assembler.InstructionSize:])
		if got, err := assembler.Decode(word); err != nil || got != want {
			t.Fatalf("decoded=%v, error=%v; want %v", got, err, want)
		}
	}
}

// TestVMMemoryBoundary exercises the maximum offset and last memory cell with a full-width round trip.
func TestVMMemoryBoundary(t *testing.T) {
	source := "stack = 65280\nstack = 536870911\nmemory[pop()+255] = pop()\n" +
		"stack = 65534\nstack = 65535\nstack = memory[pop()]\nmemory[pop()+0] = pop()"
	machine, err := machineFromSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	var dump bytes.Buffer
	if err := machine.Dump(&dump, 65534, 65535); err != nil {
		t.Fatal(err)
	}
	if dump.String() != "address,value\n65534,536870911\n65535,536870911\n" {
		t.Fatalf("dump=%q", dump.String())
	}
}

// TestVMSelfModificationError ensures an invalid replacement is decoded at fetch time.
func TestVMSelfModificationError(t *testing.T) {
	_, err := machineFromSource(t, "stack = 3\nstack = 0\nmemory[pop()+0] = pop()\nstack = 1")
	if err == nil || !strings.Contains(err.Error(), "команда 4 (адрес 3)") {
		t.Fatalf("error=%v", err)
	}
}
