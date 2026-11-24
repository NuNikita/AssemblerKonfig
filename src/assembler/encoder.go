package assembler

import (
	"encoding/binary"
	"fmt"
)

const (
	// InstructionSize is the fixed machine instruction width in bytes for variant 17.
	InstructionSize = 4
	operandShift    = 3
)

// Encode converts validated instruction fields into little-endian machine code.
// Reserved bits are zero; an invalid instruction yields no partial binary.
func Encode(program []Instruction) ([]byte, error) {
	code := make([]byte, 0)
	for index, instruction := range program {
		if err := validateInstruction(instruction); err != nil {
			return nil, fmt.Errorf("команда %d: %w", index+1, err)
		}
		word := instruction.A | instruction.B<<operandShift
		code = binary.LittleEndian.AppendUint32(code, word)
	}
	return code, nil
}

// validateInstruction checks opcodes and field widths even for manually constructed IR.
func validateInstruction(instruction Instruction) error {
	var limit uint32
	switch instruction.A {
	case LoadConstant:
		limit = MaxConstant
	case WriteMemory, Maximum:
		limit = MaxOffset
	case ReadMemory:
		limit = 0
	default:
		return fmt.Errorf("неизвестный код операции A=%d", instruction.A)
	}
	if instruction.B > limit {
		return fmt.Errorf("для A=%d поле B должно быть от 0 до %d", instruction.A, limit)
	}
	return nil
}
