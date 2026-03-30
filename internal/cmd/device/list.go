package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	cmdutil.ListFlags
	Query        string
	Online       string
	Status       string
	Product      []string
	Group        []string
	Org          string
	Firmware     string
	Name         string
	SerialNumber string
	IP           string
	Label        []string
	ICCID        string
	MAC          string
}

var defaultListFields = []string{"_id", "name", "serialNumber", "online", "product", "firmware"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List devices",
		Long:    "List devices on the InCloud platform with optional filtering, searching, and pagination.",
		Aliases: []string{"ls"},
		Example: `  # Search by name or serial number
  incloud device list -q "router"

  # Filter by online status, product, org
  incloud device list --online true --product IR615 --org <org-id>

  # Expand related resources and output as JSON
  incloud device list --expand org,firmwareUpgradeStatus -o json

  # Export offline devices as CSV
  incloud device list --online false --jq '.result[] | [.name, .serialNumber] | @csv'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			if opts.Query != "" {
				q.Set("q", opts.Query)
			}
			if opts.Online != "" {
				q.Set("online", opts.Online)
			}
			if opts.Status != "" {
				switch opts.Status {
				case "online":
					q.Set("online", "true")
				case "offline":
					q.Set("online", "false")
				}
			}
			for _, p := range opts.Product {
				q.Add("product", p)
			}
			for _, g := range opts.Group {
				q.Add("devicegroupId", g)
			}
			if opts.Org != "" {
				q.Set("oid", opts.Org)
			}
			if opts.Firmware != "" {
				q.Set("firmware", opts.Firmware)
			}
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.SerialNumber != "" {
				q.Set("serial_number", opts.SerialNumber)
			}
			if opts.IP != "" {
				q.Set("ip", opts.IP)
			}
			for _, l := range opts.Label {
				q.Add("labels", l)
			}
			if opts.ICCID != "" {
				q.Set("iccid", opts.ICCID)
			}
			if opts.MAC != "" {
				q.Set("mac", opts.MAC)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/devices", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Search by name or serial number")
	cmd.Flags().StringVar(&opts.Online, "online", "", "Filter by online status (true/false)")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (online/offline)")
	cmd.Flags().StringArrayVar(&opts.Product, "product", nil, "Filter by product (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Group, "group", nil, "Filter by device group ID (can be repeated)")
	cmd.Flags().StringVar(&opts.Org, "org", "", "Filter by organization ID")
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Filter by firmware version")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name (exact match)")
	cmd.Flags().StringVar(&opts.SerialNumber, "serial-number", "", "Filter by serial number (exact match)")
	cmd.Flags().StringVar(&opts.IP, "ip", "", "Filter by IP address")
	cmd.Flags().StringArrayVar(&opts.Label, "label", nil, "Filter by label key=value (can be repeated)")
	cmd.Flags().StringVar(&opts.ICCID, "iccid", "", "Filter by ICCID")
	cmd.Flags().StringVar(&opts.MAC, "mac", "", "Filter by MAC address")
	opts.ListFlags.RegisterExpand(cmd)

	return cmd
}
