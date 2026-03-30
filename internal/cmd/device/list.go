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
		Example: `  # List devices with default pagination
  incloud device list

  # Paginate
  incloud device list --page 2 --limit 50

  # Filter by online status
  incloud device list --online true

  # Search by name or serial number
  incloud device list -q "router"

  # Filter by product
  incloud device list --product IR615

  # Filter by org
  incloud device list --org <org-id>

  # Filter by firmware version
  incloud device list --firmware 2.0.0

  # Filter by device name (exact)
  incloud device list --name "my-router"

  # Filter by serial number (exact)
  incloud device list --serial-number IR6151234567890

  # Filter by IP address
  incloud device list --ip 192.168.1.1

  # Filter by label
  incloud device list --label env=prod --label region=us

  # Filter by ICCID
  incloud device list --iccid 89860000000000000000

  # Filter by MAC address
  incloud device list --mac 00:11:22:33:44:55

  # Sort results
  incloud device list --sort "name,asc"

  # Expand related resources (e.g. org info, firmware upgrade status)
  incloud device list --expand org,firmwareUpgradeStatus

  # Table output with selected fields
  incloud device list -o table -f name -f serialNumber -f online

  # Extract names with jq
  incloud device list --jq '.result[].name'

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
