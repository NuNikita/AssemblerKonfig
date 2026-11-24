package vm

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// Dump writes an inclusive memory interval as CSV with decimal address and value columns.
func (machine *Machine) Dump(output io.Writer, start, end int) error {
	if start < 0 || end < start || end >= len(machine.memory) {
		return fmt.Errorf("диапазон должен удовлетворять 0 <= start <= end < %d", len(machine.memory))
	}
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"address", "value"}); err != nil {
		return err
	}
	for address := start; address <= end; address++ {
		row := []string{strconv.Itoa(address), strconv.FormatUint(uint64(machine.memory[address]), 10)}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
