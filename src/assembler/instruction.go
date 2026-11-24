// Package assembler translates the algebraic language of variant 17 into instruction fields.
package assembler

import "fmt"

const (
	// LoadConstant is the opcode for pushing an immediate constant.
	LoadConstant uint32 = 1
	// ReadMemory is the opcode for reading memory using a popped address.
	ReadMemory uint32 = 7
	// WriteMemory is the opcode for writing a popped value at a popped address plus an offset.
	WriteMemory uint32 = 5
	// Maximum is the opcode for pushing the maximum of two memory operands.
	Maximum uint32 = 2
	// MaxConstant is the largest unsigned value in bits 3 through 31.
	MaxConstant uint32 = (1 << 29) - 1
	// MaxOffset is the largest unsigned value in bits 3 through 10.
	MaxOffset uint32 = (1 << 8) - 1
)

// Instruction holds the fields of one parsed instruction; ReadMemory has no B field.
type Instruction struct {
	A uint32
	B uint32
}

// String formats fields in the same order and notation as the specification tests.
func (instruction Instruction) String() string {
	if instruction.A == ReadMemory {
		return fmt.Sprintf("(A=%d)", instruction.A)
	}
	return fmt.Sprintf("(A=%d, B=%d)", instruction.A, instruction.B)
}
