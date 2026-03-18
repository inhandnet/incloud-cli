package main

import (
	"fmt"
	"os"

	cmd "github.com/inhandnet/incloud-cli/internal/cmd"
	activityCmd "github.com/inhandnet/incloud-cli/internal/cmd/activity"
	alertCmd "github.com/inhandnet/incloud-cli/internal/cmd/alert"
	apiCmd "github.com/inhandnet/incloud-cli/internal/cmd/api"
	authCmd "github.com/inhandnet/incloud-cli/internal/cmd/auth"
	configCmd "github.com/inhandnet/incloud-cli/internal/cmd/config"
	deviceCmd "github.com/inhandnet/incloud-cli/internal/cmd/device"
	firmwareCmd "github.com/inhandnet/incloud-cli/internal/cmd/firmware"
	networkCmd "github.com/inhandnet/incloud-cli/internal/cmd/network"
	orgCmd "github.com/inhandnet/incloud-cli/internal/cmd/org"
	overviewCmd "github.com/inhandnet/incloud-cli/internal/cmd/overview"
	productCmd "github.com/inhandnet/incloud-cli/internal/cmd/product"
	roleCmd "github.com/inhandnet/incloud-cli/internal/cmd/role"
	userCmd "github.com/inhandnet/incloud-cli/internal/cmd/user"
	versionCmd "github.com/inhandnet/incloud-cli/internal/cmd/version"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func main() {
	f := factory.New()
	rootCmd := cmd.NewCmdRoot(f)
	rootCmd.AddCommand(activityCmd.NewCmdActivity(f))
	rootCmd.AddCommand(alertCmd.NewCmdAlert(f))
	rootCmd.AddCommand(configCmd.NewCmdConfig(f))
	rootCmd.AddCommand(apiCmd.NewCmdApi(f))
	rootCmd.AddCommand(authCmd.NewCmdAuth(f))
	rootCmd.AddCommand(deviceCmd.NewCmdDevice(f))
	rootCmd.AddCommand(firmwareCmd.NewCmdFirmware(f))
	rootCmd.AddCommand(networkCmd.NewCmdNetwork(f))
	rootCmd.AddCommand(orgCmd.NewCmdOrg(f))
	rootCmd.AddCommand(overviewCmd.NewCmdOverview(f))
	rootCmd.AddCommand(productCmd.NewCmdProduct(f))
	rootCmd.AddCommand(roleCmd.NewCmdRole(f))
	rootCmd.AddCommand(userCmd.NewCmdUser(f))
	rootCmd.AddCommand(versionCmd.NewCmdVersion(f))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
