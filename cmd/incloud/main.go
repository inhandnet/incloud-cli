package main

import (
	"fmt"
	"os"

	cmd "github.com/inhandnet/incloud-cli/internal/cmd"
	configCmd "github.com/inhandnet/incloud-cli/internal/cmd/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

var version = "dev"

func main() {
	f := factory.New()
	rootCmd := cmd.NewCmdRoot(f, version)
	rootCmd.AddCommand(configCmd.NewCmdConfig(f))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
