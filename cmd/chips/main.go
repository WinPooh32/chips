// Command chips is a local task tracker on plain markdown files.
package main

import (
	"fmt"
	"os"

	"github.com/WinPooh32/chips/cmd/chips/internal/commands"
)

func main() {
	if err := commands.NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
