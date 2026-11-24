package assembler

import "fmt"

const opcodeMask uint32 = 7

// Decode converts a machine word to the shared IR, rejecting unknown opcodes and reserved bits.
func Decode(word uint32) (Instruction, error) {
	instruction := Instruction{A: word & opcodeMask, B: word >> operandShift}
	if err := validateInstruction(instruction); err != nil {
		return Instruction{}, fmt.Errorf("некорректное машинное слово 0x%08X: %w", word, err)
	}
	return instruction, nil
}
