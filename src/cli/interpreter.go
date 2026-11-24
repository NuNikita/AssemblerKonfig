package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"AssemblerKonfig/src/vm"
)

// interpreterOptions holds the binary path, CSV path, and inclusive dump interval.
type interpreterOptions struct {
	input  string
	output string
	start  int
	end    int
}

// RunInterpreter executes a binary program and returns a process exit code.
func RunInterpreter(arguments []string, stdout, stderr io.Writer) int {
	settings, err := parseInterpreterOptions(arguments, stdout)
	if errors.Is(err, flag.ErrHelp) {
		return Success
	}
	if err == nil {
		err = interpret(settings)
	}
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return Failure
	}
	return Success
}

// parseInterpreterOptions requires all four arguments and validates the interval before execution.
func parseInterpreterOptions(arguments []string, help io.Writer) (interpreterOptions, error) {
	settings := interpreterOptions{start: -1, end: -1}
	flags := flag.NewFlagSet("interpreter", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&settings.input, "input", "", "путь к двоичной программе")
	flags.StringVar(&settings.output, "output", "", "путь к CSV-дампу памяти")
	flags.IntVar(&settings.start, "start", -1, "начальный адрес дампа (включительно)")
	flags.IntVar(&settings.end, "end", -1, "конечный адрес дампа (включительно)")
	flags.Usage = func() {}
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			if helpErr := writeInterpreterHelp(flags, help); helpErr != nil {
				return settings, helpErr
			}
		}
		return settings, err
	}
	if flags.NArg() != 0 {
		return settings, fmt.Errorf("лишние позиционные аргументы: %s", strings.Join(flags.Args(), " "))
	}
	return settings, validateInterpreterOptions(settings)
}

// validateInterpreterOptions checks paths and memory bounds, including omitted range flags.
func validateInterpreterOptions(settings interpreterOptions) error {
	if strings.TrimSpace(settings.input) == "" || strings.TrimSpace(settings.output) == "" {
		return errors.New("обязательны -input и -output; справка: -help")
	}
	if settings.start < 0 || settings.end < settings.start || settings.end >= vm.MemorySize {
		return fmt.Errorf("обязательны -start и -end: 0 <= start <= end < %d", vm.MemorySize)
	}
	return nil
}

// writeInterpreterHelp renders interpreter flags and propagates stream failures.
func writeInterpreterHelp(flags *flag.FlagSet, output io.Writer) error {
	var usage bytes.Buffer
	fmt.Fprintln(&usage, "Использование: interpreter -input program.bin -output dump.csv -start 0 -end 255")
	flags.SetOutput(&usage)
	flags.PrintDefaults()
	_, err := io.Copy(output, &usage)
	return err
}

// interpret loads and executes the binary, then atomically saves the requested CSV interval.
func interpret(settings interpreterOptions) error {
	code, err := os.ReadFile(settings.input)
	if err != nil {
		return fmt.Errorf("чтение двоичной программы: %w", err)
	}
	machine, err := vm.New(code)
	if err != nil {
		return err
	}
	if err := machine.Run(); err != nil {
		return err
	}
	var dump bytes.Buffer
	if err := machine.Dump(&dump, settings.start, settings.end); err != nil {
		return err
	}
	return writeBinary(settings.input, settings.output, dump.Bytes())
}
