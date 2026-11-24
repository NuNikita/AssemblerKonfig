package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// writeBinary protects source aliases and replaces the output only after a complete write and close.
func writeBinary(input, output string, code []byte) error {
	source, err := os.Stat(input)
	if err != nil {
		return fmt.Errorf("проверка исходного файла: %w", err)
	}
	target, err := os.Stat(output)
	if err == nil && os.SameFile(source, target) {
		return fmt.Errorf("исходный и выходной файлы должны различаться")
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("проверка выходного файла: %w", err)
	}
	if err := saveBinary(output, code); err != nil {
		return fmt.Errorf("запись двоичного файла: %w", err)
	}
	return nil
}

// saveBinary writes a temporary sibling and atomically renames it, removing temporary files on errors.
func saveBinary(path string, code []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".assembler-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(code); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
