package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/inhandnet/incloud-cli/internal/api"
	cmd "github.com/inhandnet/incloud-cli/internal/cmd"
	activityCmd "github.com/inhandnet/incloud-cli/internal/cmd/activity"
	alertCmd "github.com/inhandnet/incloud-cli/internal/cmd/alert"
	apiCmd "github.com/inhandnet/incloud-cli/internal/cmd/api"
	authCmd "github.com/inhandnet/incloud-cli/internal/cmd/auth"
	configCmd "github.com/inhandnet/incloud-cli/internal/cmd/config"
	connectorCmd "github.com/inhandnet/incloud-cli/internal/cmd/connector"
	deviceCmd "github.com/inhandnet/incloud-cli/internal/cmd/device"
	feedbackCmd "github.com/inhandnet/incloud-cli/internal/cmd/feedback"
	firmwareCmd "github.com/inhandnet/incloud-cli/internal/cmd/firmware"
	licenseCmd "github.com/inhandnet/incloud-cli/internal/cmd/license"
	oobmCmd "github.com/inhandnet/incloud-cli/internal/cmd/oobm"
	orgCmd "github.com/inhandnet/incloud-cli/internal/cmd/org"
	overviewCmd "github.com/inhandnet/incloud-cli/internal/cmd/overview"
	productCmd "github.com/inhandnet/incloud-cli/internal/cmd/product"
	roleCmd "github.com/inhandnet/incloud-cli/internal/cmd/role"
	sdwanCmd "github.com/inhandnet/incloud-cli/internal/cmd/sdwan"
	tunnelCmd "github.com/inhandnet/incloud-cli/internal/cmd/tunnel"
	updateCmd "github.com/inhandnet/incloud-cli/internal/cmd/update"
	userCmd "github.com/inhandnet/incloud-cli/internal/cmd/user"
	versionCmd "github.com/inhandnet/incloud-cli/internal/cmd/version"
	webhookCmd "github.com/inhandnet/incloud-cli/internal/cmd/webhook"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func main() {
	f := factory.New()
	rootCmd := cmd.NewCmdRoot(f)
	rootCmd.AddCommand(activityCmd.NewCmdActivity(f))
	rootCmd.AddCommand(alertCmd.NewCmdAlert(f))
	rootCmd.AddCommand(configCmd.NewCmdConfig(f))
	rootCmd.AddCommand(connectorCmd.NewCmdConnector(f))
	rootCmd.AddCommand(apiCmd.NewCmdApi(f))
	rootCmd.AddCommand(authCmd.NewCmdAuth(f))
	rootCmd.AddCommand(deviceCmd.NewCmdDevice(f))
	rootCmd.AddCommand(feedbackCmd.NewCmdFeedback(f))
	rootCmd.AddCommand(firmwareCmd.NewCmdFirmware(f))
	rootCmd.AddCommand(licenseCmd.NewCmdLicense(f))
	rootCmd.AddCommand(oobmCmd.NewCmdOobm(f))
	rootCmd.AddCommand(orgCmd.NewCmdOrg(f))
	rootCmd.AddCommand(overviewCmd.NewCmdOverview(f))
	rootCmd.AddCommand(productCmd.NewCmdProduct(f))
	rootCmd.AddCommand(roleCmd.NewCmdRole(f))
	rootCmd.AddCommand(sdwanCmd.NewCmdSdwan(f))
	rootCmd.AddCommand(tunnelCmd.NewCmdTunnel(f))
	rootCmd.AddCommand(updateCmd.NewCmdUpdate(f))
	rootCmd.AddCommand(userCmd.NewCmdUser(f))
	rootCmd.AddCommand(versionCmd.NewCmdVersion(f))
	rootCmd.AddCommand(webhookCmd.NewCmdWebhook(f))

	cmd.SetupSuperAdminFlags(rootCmd, f)

	// Top-level shortcuts: `incloud login` → `incloud auth login`
	loginAlias := authCmd.NewCmdLogin(f)
	loginAlias.Use = "login"
	loginAlias.Hidden = true
	rootCmd.AddCommand(loginAlias)

	executedCmd, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		// Show usage for flag-related errors (missing required flags, unknown flags)
		if isFlagError(err) && executedCmd != nil {
			fmt.Fprintln(os.Stderr)
			fmt.Fprint(os.Stderr, executedCmd.UsageString())
		}
		// Hint re-authentication on 401 Unauthorized
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 401 {
			fmt.Fprintln(os.Stderr, "Hint: run 'incloud auth login' to re-authenticate")
		}
		os.Exit(1)
	}
}

func isFlagError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "required flag") ||
		strings.Contains(msg, "unknown flag") ||
		strings.Contains(msg, "flag needs an argument")
}
