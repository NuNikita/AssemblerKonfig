package assembler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	statementStart = `^\s*`
	statementEnd   = `\s*;?\s*$`
	stackTarget    = `stack\s*=\s*`
	popExpression  = `pop\s*\(\s*\)`
	memoryStart    = `memory\s*\[\s*`
	memoryEnd      = `\s*\]`
	numberPattern  = `([0-9]+)`
	decimalBase    = 10
	fieldBits      = 32
)

// rule binds a complete algebraic statement to an opcode and operand limit.
type rule struct {
	pattern *regexp.Regexp
	opcode  uint32
	limit   uint32
}

var rules = []rule{
	newRule(stackTarget+numberPattern, LoadConstant, MaxConstant),
	newRule(stackTarget+memoryStart+popExpression+memoryEnd, ReadMemory, 0),
	newRule(memoryStart+popExpression+`\s*\+\s*`+numberPattern+memoryEnd+`\s*=\s*`+popExpression,
		WriteMemory, MaxOffset),
	newRule(stackTarget+`max\s*\(\s*`+memoryStart+popExpression+memoryEnd+`\s*,\s*`+
		memoryStart+popExpression+`\s*\+\s*`+numberPattern+memoryEnd+`\s*\)`, Maximum, MaxOffset),
}

// newRule compiles a statement pattern with common whitespace and terminator rules.
func newRule(expression string, opcode, limit uint32) rule {
	return rule{regexp.MustCompile(statementStart + expression + statementEnd), opcode, limit}
}

// Parse reads a complete program, preserving command order and reporting source line numbers.
// An invalid program yields no partial intermediate representation.
func Parse(source io.Reader) ([]Instruction, error) {
	reader := bufio.NewReader(source)
	program := make([]Instruction, 0)
	for lineNumber := 1; ; lineNumber++ {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return nil, fmt.Errorf("строка %d: чтение: %w", lineNumber, readErr)
		}
		statement := strings.TrimSpace(strings.SplitN(line, "//", 2)[0])
		if statement != "" {
			instruction, err := parseStatement(statement)
			if err != nil {
				return nil, fmt.Errorf("строка %d: %w", lineNumber, err)
			}
			program = append(program, instruction)
		}
		if errors.Is(readErr, io.EOF) {
			return program, nil
		}
	}
}

// parseStatement recognizes exactly one supported command and validates its operand.
func parseStatement(statement string) (Instruction, error) {
	for _, candidate := range rules {
		fields := candidate.pattern.FindStringSubmatch(statement)
		if fields == nil {
			continue
		}
		instruction := Instruction{A: candidate.opcode}
		if candidate.opcode == ReadMemory {
			return instruction, nil
		}
		value, err := strconv.ParseUint(fields[1], decimalBase, fieldBits)
		if err != nil || value > uint64(candidate.limit) {
			return Instruction{}, fmt.Errorf("операнд B должен быть целым числом от 0 до %d", candidate.limit)
		}
		instruction.B = uint32(value)
		return instruction, nil
	}
	return Instruction{}, fmt.Errorf("неизвестная команда или неверный синтаксис: %q", statement)
}
