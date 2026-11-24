// Package cli provides the command-line interface for stage one of the assembler.
package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"AssemblerKonfig/src/assembler"
)

const (
	// Success is the exit code for a successful invocation or requested help.
	Success = 0
	// Failure is the exit code for invalid arguments, input, or output errors.
	Failure = 1
)

// options holds the input path, future binary path, and diagnostic output mode.
type options struct {
	input    string
	output   string
	testMode bool
}

// Run translates a source file and returns a process exit code without terminating the caller.
func Run(arguments []string, stdout, stderr io.Writer) int {
	settings, err := parseOptions(arguments, stdout)
	if errors.Is(err, flag.ErrHelp) {
		return Success
	}
	if err == nil {
		err = assemble(settings, stdout)
	}
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return Failure
	}
	return Success
}

// parseOptions validates all CLI arguments and prints usage for -h or -help.
func parseOptions(arguments []string, help io.Writer) (options, error) {
	var settings options
	flags := flag.NewFlagSet("assembler", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&settings.input, "input", "", "путь к исходному файлу")
	flags.StringVar(&settings.output, "output", "", "путь к будущему двоичному файлу (этап 2)")
	flags.BoolVar(&settings.testMode, "test", false, "вывести поля промежуточного представления")
	flags.Usage = func() {}
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			if helpErr := writeHelp(flags, help); helpErr != nil {
				return settings, helpErr
			}
		}
		return settings, err
	}
	if flags.NArg() != 0 {
		return settings, fmt.Errorf("лишние позиционные аргументы: %s", strings.Join(flags.Args(), " "))
	}
	if strings.TrimSpace(settings.input) == "" || strings.TrimSpace(settings.output) == "" {
		return settings, errors.New("обязательны аргументы -input и -output; справка: -help")
	}
	return settings, nil
}

// writeHelp renders usage only for an explicit help request and propagates output errors.
func writeHelp(flags *flag.FlagSet, output io.Writer) error {
	var usage bytes.Buffer
	fmt.Fprintln(&usage, "Использование: assembler -input program.asm -output program.bin [-test]")
	flags.SetOutput(&usage)
	flags.PrintDefaults()
	_, err := io.Copy(output, &usage)
	return err
}

// assemble reads the source and emits stage-one output without opening the future binary path.
func assemble(settings options, output io.Writer) error {
	source, err := os.Open(settings.input)
	if err != nil {
		return fmt.Errorf("открытие исходного файла: %w", err)
	}
	defer source.Close()
	program, err := assembler.Parse(source)
	if err != nil {
		return err
	}
	return writeProgram(program, settings.testMode, output)
}

// writeProgram prints either exact instruction fields or a stage-one completion message.
func writeProgram(program []assembler.Instruction, testMode bool, output io.Writer) error {
	if !testMode {
		_, err := fmt.Fprintf(output, "Разобрано команд: %d. Промежуточное представление построено.\n", len(program))
		return err
	}
	for _, instruction := range program {
		if _, err := fmt.Fprintln(output, instruction); err != nil {
			return fmt.Errorf("вывод промежуточного представления: %w", err)
		}
	}
	return nil
}
