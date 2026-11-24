// Command assembler translates variant 17 programs into binary machine code.
package main

import (
	"os"

	"AssemblerKonfig/src/cli"
)

// main passes process arguments and streams to the CLI.
func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
