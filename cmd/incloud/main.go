package main

import (
	"fmt"
	"os"

	cmd "github.com/inhandnet/incloud-cli/internal/cmd"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

var version = "dev"

func main() {
	f := factory.New()
	rootCmd := cmd.NewCmdRoot(f, version)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
