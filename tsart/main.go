package main

import (
	"fmt"
	"os"

	"github.com/microsoft/typescript-go/tsart/cmd"
)

func main() {
	// plugin.Main owns --manifest/--location. Let it receive those flags without
	// Cobra consuming them first.
	if len(os.Args) > 1 && os.Args[1] == "nvim" {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		cmd.RunNvimPlugin()
		return
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
