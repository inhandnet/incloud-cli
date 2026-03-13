package main

import (
	"fmt"
	"os"

	cmd "github.com/inhandnet/incloud-cli/internal/cmd"
)

var version = "dev"

func main() {
	rootCmd := cmd.NewCmdRoot(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
