package main

import (
	"fmt"
	"os"

	cmd "github.com/inhandnet/incloud-cli/internal/cmd"
	apiCmd "github.com/inhandnet/incloud-cli/internal/cmd/api"
	authCmd "github.com/inhandnet/incloud-cli/internal/cmd/auth"
	configCmd "github.com/inhandnet/incloud-cli/internal/cmd/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

var version = "dev"

func main() {
	f := factory.New()
	rootCmd := cmd.NewCmdRoot(f, version)
	rootCmd.AddCommand(configCmd.NewCmdConfig(f))
	rootCmd.AddCommand(apiCmd.NewCmdApi(f))
	rootCmd.AddCommand(authCmd.NewCmdAuth(f))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
