// Package vm implements the unified-memory stack machine of variant 17, stage three.
package vm

import (
	"encoding/binary"
	"fmt"

	"AssemblerKonfig/src/assembler"
)

const (
	// MemorySize is the number of addressable 32-bit words shared by instructions and data.
	MemorySize    = 1 << 16
	writeOperands = 2
)

// Machine owns unified word-addressed memory, a separate stack, and the next instruction address.
type Machine struct {
	memory      []uint32
	stack       []uint32
	pc          int
	programSize int
}

// New loads little-endian words into zero-initialized unified memory beginning at address zero.
func New(code []byte) (*Machine, error) {
	if len(code)%assembler.InstructionSize != 0 {
		return nil, fmt.Errorf("размер программы должен быть кратен %d байтам", assembler.InstructionSize)
	}
	count := len(code) / assembler.InstructionSize
	if count > MemorySize {
		return nil, fmt.Errorf("программа превышает память: %d слов", MemorySize)
	}
	machine := &Machine{memory: make([]uint32, MemorySize), programSize: count}
	for index := 0; index < count; index++ {
		machine.memory[index] = binary.LittleEndian.Uint32(code[index*assembler.InstructionSize:])
	}
	return machine, nil
}

// Run fetches and decodes each instruction from current memory until the original program boundary.
func (machine *Machine) Run() error {
	for machine.pc < machine.programSize {
		instruction, err := assembler.Decode(machine.memory[machine.pc])
		if err == nil {
			err = machine.execute(instruction)
		}
		if err != nil {
			return fmt.Errorf("команда %d (адрес %d): %w", machine.pc+1, machine.pc, err)
		}
		machine.pc++
	}
	return nil
}

// execute dispatches the three basic operations without implementing stage-four arithmetic.
func (machine *Machine) execute(instruction assembler.Instruction) error {
	switch instruction.A {
	case assembler.LoadConstant:
		machine.stack = append(machine.stack, instruction.B)
		return nil
	case assembler.ReadMemory:
		return machine.read()
	case assembler.WriteMemory:
		return machine.write(instruction.B)
	case assembler.Maximum:
		return fmt.Errorf("операция max() относится к этапу 4 и пока не поддерживается")
	default:
		return fmt.Errorf("неизвестный код операции: %d", instruction.A)
	}
}

// read replaces the top address with its memory value, preserving state on an invalid address.
func (machine *Machine) read() error {
	if len(machine.stack) == 0 {
		return fmt.Errorf("недостаточно элементов стека для чтения")
	}
	top := len(machine.stack) - 1
	address := uint64(machine.stack[top])
	if address >= uint64(len(machine.memory)) {
		return fmt.Errorf("адрес чтения вне памяти: %d", address)
	}
	machine.stack[top] = machine.memory[address]
	return nil
}

// write pops the value first and then the base address, adding the offset without uint32 overflow.
func (machine *Machine) write(offset uint32) error {
	if len(machine.stack) < writeOperands {
		return fmt.Errorf("недостаточно элементов стека для записи")
	}
	top := len(machine.stack) - 1
	address := uint64(machine.stack[top-1]) + uint64(offset)
	if address >= uint64(len(machine.memory)) {
		return fmt.Errorf("адрес записи вне памяти: %d", address)
	}
	machine.memory[address] = machine.stack[top]
	machine.stack = machine.stack[:top-1]
	return nil
}
