// Command interpreter executes variant 17 binary programs and saves a CSV memory dump.
package main

import (
	"os"

	"AssemblerKonfig/src/cli"
)

// main passes interpreter arguments and streams to the CLI.
func main() {
	os.Exit(cli.RunInterpreter(os.Args[1:], os.Stdout, os.Stderr))
}
