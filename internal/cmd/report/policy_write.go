package report

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var (
	weekdayNames = []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}
	periods      = []string{"LAST_1_DAY", "LAST_3_DAYS", "LAST_7_DAYS", "LAST_30_DAYS"}
	formats      = []string{"PDF", "XLSX"}
	applyToTypes = []string{"ALL", "ORG", "GROUP"}
	frequencies  = []string{"daily", "weekly", "monthly"}
)

// policyOptions holds the flags shared by report policy create and update.
type policyOptions struct {
	Name        string
	ApplyTo     string
	Targets     []string
	Templates   []string
	Format      string
	Period      string
	Recipient   string
	Description string
	Timezone    string
	Frequency   string
	At          string
	DayOfWeek   string
	DayOfMonth  int
	From        string
	To          string
}

func (o *policyOptions) registerFlags(cmd *cobra.Command, create bool) {
	fl := cmd.Flags()
	fl.StringVar(&o.Name, "name", "", "Policy name (max 128 characters)")
	fl.StringVar(&o.ApplyTo, "apply-to", "", "Devices covered: all, org or group (default all on create)")
	fl.StringArrayVar(&o.Targets, "target", nil, "Organization or group ID for --apply-to org/group (can be repeated)")
	fl.StringArrayVar(&o.Templates, "template", nil, "Report template (can be repeated; see 'incloud report policy templates')")
	fl.StringVar(&o.Format, "format", "", "Report file format: pdf or xlsx (default pdf on create)")
	fl.StringVar(&o.Period, "period", "", "Statistics window: last_1_day (yesterday), last_3_days, last_7_days or last_30_days")
	fl.StringVar(&o.Recipient, "recipient", "", "Email address that receives the report")
	fl.StringVar(&o.Description, "description", "", "Description (max 512 characters)")
	fl.StringVar(&o.Timezone, "timezone", "", "IANA timezone for the schedule and statistics window (e.g. Asia/Shanghai; default: server timezone)")
	fl.StringVar(&o.Frequency, "frequency", "", "Schedule: daily, weekly or monthly")
	fl.StringVar(&o.At, "at", "", "Time of day the scheduled report runs (HH:MM, in --timezone)")
	fl.StringVar(&o.DayOfWeek, "day-of-week", "", "Day for weekly schedules: MONDAY..SUNDAY or 1-7 (1 = Monday)")
	fl.IntVar(&o.DayOfMonth, "day-of-month", 0, "Day for monthly schedules (1-28)")
	if create {
		fl.StringVar(&o.From, "from", "", "One-off report: first day (YYYY-MM-DD, in --timezone)")
		fl.StringVar(&o.To, "to", "", "One-off report: last day, inclusive (YYYY-MM-DD, in --timezone)")
	}
}

// normalizeEnum upper-cases v and checks it against allowed.
func normalizeEnum(flag, v string, allowed []string) (string, error) {
	u := strings.ToUpper(strings.TrimSpace(v))
	if !slices.Contains(allowed, u) {
		return "", fmt.Errorf("invalid --%s %q: must be one of %s", flag, v, strings.ToLower(strings.Join(allowed, ", ")))
	}
	return u, nil
}

func (o *policyOptions) templateList() ([]string, error) {
	out := make([]string, 0, len(o.Templates))
	for _, t := range o.Templates {
		name := strings.ToUpper(strings.TrimSpace(t))
		if _, ok := lookupTemplate(name); !ok {
			return nil, fmt.Errorf("unknown --template %q (use 'incloud report policy templates' to list templates)", t)
		}
		out = append(out, name)
	}
	return out, nil
}

// checkXLSXOnly rejects templates that cannot be rendered as PDF.
func checkXLSXOnly(templates []string, format string) error {
	if format != "PDF" {
		return nil
	}
	for _, t := range templates {
		if def, _ := lookupTemplate(t); def.XLSXOnly {
			return fmt.Errorf("template %s is only available in xlsx format; pass --format xlsx", t)
		}
	}
	return nil
}

func (o *policyOptions) applyTo() (map[string]any, error) {
	typ := "ALL"
	if o.ApplyTo != "" {
		var err error
		if typ, err = normalizeEnum("apply-to", o.ApplyTo, applyToTypes); err != nil {
			return nil, err
		}
	}
	if typ == "ALL" {
		if len(o.Targets) > 0 {
			return nil, fmt.Errorf("--target is only used with --apply-to org or group")
		}
		return map[string]any{"type": typ}, nil
	}
	if len(o.Targets) == 0 {
		return nil, fmt.Errorf("--apply-to %s requires at least one --target", strings.ToLower(typ))
	}
	for _, id := range o.Targets {
		if !cmdutil.IsObjectID(id) {
			return nil, fmt.Errorf("invalid --target %q: must be a 24-character hex ID", id)
		}
	}
	return map[string]any{"type": typ, "ids": o.Targets}, nil
}

// schedule builds the schedule object from --frequency/--at/--day-of-*.
func (o *policyOptions) schedule() (map[string]any, error) {
	freq := strings.ToLower(strings.TrimSpace(o.Frequency))
	if !slices.Contains(frequencies, freq) {
		return nil, fmt.Errorf("invalid --frequency %q: must be one of daily, weekly, monthly", o.Frequency)
	}
	if o.At == "" {
		return nil, fmt.Errorf("--frequency requires --at HH:MM")
	}
	t, err := time.Parse("15:04", o.At)
	if err != nil {
		return nil, fmt.Errorf("invalid --at %q: expected HH:MM (24-hour)", o.At)
	}
	s := map[string]any{"hourOfDay": t.Hour(), "minuteOfHour": t.Minute()}

	switch freq {
	case "weekly":
		if o.DayOfWeek == "" {
			return nil, fmt.Errorf("--frequency weekly requires --day-of-week")
		}
		d, err := parseDayOfWeek(o.DayOfWeek)
		if err != nil {
			return nil, err
		}
		s["dayOfWeek"] = d
	case "monthly":
		if o.DayOfMonth < 1 || o.DayOfMonth > 28 {
			return nil, fmt.Errorf("--frequency monthly requires --day-of-month between 1 and 28")
		}
		s["dayOfMonth"] = o.DayOfMonth
	}
	if freq != "weekly" && o.DayOfWeek != "" {
		return nil, fmt.Errorf("--day-of-week is only used with --frequency weekly")
	}
	if freq != "monthly" && o.DayOfMonth != 0 {
		return nil, fmt.Errorf("--day-of-month is only used with --frequency monthly")
	}
	return s, nil
}

func parseDayOfWeek(v string) (int, error) {
	v = strings.TrimSpace(v)
	if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 7 {
		return n, nil
	}
	if i := slices.Index(weekdayNames, strings.ToUpper(v)); i >= 0 {
		return i + 1, nil
	}
	return 0, fmt.Errorf("invalid --day-of-week %q: must be MONDAY..SUNDAY or 1-7 (1 = Monday)", v)
}

// oneOffRange converts --from/--to days into [from 00:00, day after to 00:00)
// in the policy timezone (or the local timezone when --timezone is not set).
func (o *policyOptions) oneOffRange() (string, string, error) {
	loc := time.Local
	if o.Timezone != "" {
		var err error
		if loc, err = time.LoadLocation(o.Timezone); err != nil {
			return "", "", err
		}
	}
	from, err := time.ParseInLocation("2006-01-02", o.From, loc)
	if err != nil {
		return "", "", fmt.Errorf("invalid --from %q: expected YYYY-MM-DD", o.From)
	}
	to, err := time.ParseInLocation("2006-01-02", o.To, loc)
	if err != nil {
		return "", "", fmt.Errorf("invalid --to %q: expected YYYY-MM-DD", o.To)
	}
	if to.Before(from) {
		return "", "", fmt.Errorf("--to must not be before --from")
	}
	end := to.AddDate(0, 0, 1)
	return from.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339), nil
}

func (o *policyOptions) createBody() (map[string]any, error) {
	if err := cmdutil.ValidateTimezone(o.Timezone); err != nil {
		return nil, err
	}
	if o.Name == "" {
		return nil, fmt.Errorf("--name is required")
	}
	if len(o.Templates) == 0 {
		return nil, fmt.Errorf("at least one --template is required (see 'incloud report policy templates')")
	}

	applyTo, err := o.applyTo()
	if err != nil {
		return nil, err
	}
	templates, err := o.templateList()
	if err != nil {
		return nil, err
	}
	format := "PDF"
	if o.Format != "" {
		if format, err = normalizeEnum("format", o.Format, formats); err != nil {
			return nil, err
		}
	}
	if err := checkXLSXOnly(templates, format); err != nil {
		return nil, err
	}

	body := map[string]any{
		"name":      o.Name,
		"applyTo":   applyTo,
		"templates": templates,
		"format":    format,
	}
	if o.Period != "" {
		p, err := normalizeEnum("period", o.Period, periods)
		if err != nil {
			return nil, err
		}
		body["period"] = p
	}

	if o.Frequency != "" {
		if o.From != "" || o.To != "" {
			return nil, fmt.Errorf("--from/--to are for one-off reports and cannot be combined with --frequency")
		}
		if o.Period == "" {
			return nil, fmt.Errorf("scheduled policies require --period")
		}
		s, err := o.schedule()
		if err != nil {
			return nil, err
		}
		body["scheduled"] = true
		body["schedule"] = s
	} else {
		if o.At != "" || o.DayOfWeek != "" || o.DayOfMonth != 0 {
			return nil, fmt.Errorf("--at/--day-of-week/--day-of-month require --frequency")
		}
		switch {
		case o.From != "" || o.To != "":
			if o.Period != "" {
				return nil, fmt.Errorf("--period and --from/--to cannot be combined")
			}
			if o.From == "" || o.To == "" {
				return nil, fmt.Errorf("one-off reports need both --from and --to")
			}
			from, to, err := o.oneOffRange()
			if err != nil {
				return nil, err
			}
			body["startTime"] = from
			body["endTime"] = to
		case o.Period == "":
			return nil, fmt.Errorf("pass --frequency for a scheduled policy, or --period or --from/--to for a one-off report")
		}
		body["scheduled"] = false
	}

	if o.Recipient != "" {
		body["recipient"] = o.Recipient
	}
	if o.Description != "" {
		body["description"] = o.Description
	}
	if o.Timezone != "" {
		body["timezone"] = o.Timezone
	}
	return body, nil
}

// updateBody sends only the fields whose flags were passed; the API leaves the
// rest of the policy unchanged.
func (o *policyOptions) updateBody(cmd *cobra.Command) (map[string]any, error) {
	changed := cmd.Flags().Changed
	body := map[string]any{}

	if changed("timezone") {
		if o.Timezone == "" {
			return nil, fmt.Errorf("--timezone cannot be cleared; pass an IANA timezone name such as Asia/Shanghai")
		}
		if err := cmdutil.ValidateTimezone(o.Timezone); err != nil {
			return nil, err
		}
		body["timezone"] = o.Timezone
	}
	if changed("name") {
		if o.Name == "" {
			return nil, fmt.Errorf("--name cannot be empty")
		}
		body["name"] = o.Name
	}
	if changed("target") && !changed("apply-to") {
		return nil, fmt.Errorf("--target requires --apply-to org or group")
	}
	if changed("apply-to") {
		a, err := o.applyTo()
		if err != nil {
			return nil, err
		}
		body["applyTo"] = a
	}

	var templates []string
	if changed("template") {
		var err error
		if templates, err = o.templateList(); err != nil {
			return nil, err
		}
		if len(templates) == 0 {
			return nil, fmt.Errorf("--template cannot be empty")
		}
		body["templates"] = templates
		if !changed("format") {
			for _, t := range templates {
				if def, _ := lookupTemplate(t); def.XLSXOnly {
					return nil, fmt.Errorf("template %s is only available in xlsx format; pass --format xlsx", t)
				}
			}
		}
	}
	if changed("format") {
		format, err := normalizeEnum("format", o.Format, formats)
		if err != nil {
			return nil, err
		}
		if err := checkXLSXOnly(templates, format); err != nil {
			return nil, err
		}
		body["format"] = format
	}
	if changed("period") {
		p, err := normalizeEnum("period", o.Period, periods)
		if err != nil {
			return nil, err
		}
		body["period"] = p
	}

	// The schedule is replaced as a whole, so it has to be given in full.
	if changed("frequency") || changed("at") || changed("day-of-week") || changed("day-of-month") {
		if !changed("frequency") || !changed("at") {
			return nil, fmt.Errorf("changing the schedule requires both --frequency and --at (plus --day-of-week/--day-of-month for weekly/monthly)")
		}
		s, err := o.schedule()
		if err != nil {
			return nil, err
		}
		body["schedule"] = s
	}

	if changed("recipient") {
		body["recipient"] = o.Recipient
	}
	if changed("description") {
		body["description"] = o.Description
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("nothing to update; pass at least one flag")
	}
	return body, nil
}

func NewCmdPolicyCreate(f *factory.Factory) *cobra.Command {
	opts := &policyOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a report policy",
		Long: `Create a report policy.

Scheduled policy: pass --frequency (daily, weekly or monthly) with --at and --period.
Each run covers the --period window that ends when it runs and emails the report to
--recipient.

One-off report: omit --frequency and pass --period, or --from/--to for a fixed range of
days. The report is generated immediately; one-off policies cannot be updated and are
not shown by 'incloud report policy list' (find the report with 'incloud report list').

--timezone sets the IANA timezone in which --at, the day of week/month, the statistics
window and the times shown in the report are evaluated. Without it the server's
timezone is used, so pass it explicitly when you schedule reports for a specific
region.`,
		Example: `  # Daily report at 08:00 Shanghai time covering yesterday
  incloud report policy create --name "Daily overview" \
    --template TOTAL_DEVICES --template OFFLINE_TIMES_TOP_10 \
    --period last_1_day --frequency daily --at 08:00 \
    --timezone Asia/Shanghai --recipient ops@example.com

  # Weekly XLSX report for two device groups, every Monday 09:30
  incloud report policy create --name "Weekly usage" \
    --apply-to group --target 507f1f77bcf86cd799439011 --target 507f1f77bcf86cd799439012 \
    --template DEVICE_DATA_USAGE_DETAILS --format xlsx \
    --period last_7_days --frequency weekly --day-of-week MONDAY --at 09:30 \
    --timezone Europe/Berlin

  # One-off report for a range of days
  incloud report policy create --name "October" --template USAGE_SUMMARY \
    --from 2026-10-01 --to 2026-10-31 --timezone Asia/Tokyo`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := opts.createBody()
			if err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			resp, err := client.Post(policiesPath, body)
			if err != nil {
				return err
			}
			printPolicyID(f, resp, "created")

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(resp, f.IO, output)
		},
	}

	opts.registerFlags(cmd, true)

	return cmd
}

func NewCmdPolicyUpdate(f *factory.Factory) *cobra.Command {
	opts := &policyOptions{}

	cmd := &cobra.Command{
		Use:   "update <policy-id>",
		Short: "Update a scheduled report policy",
		Long: `Update a scheduled report policy. Only the flags you pass are changed; everything
else is kept.

The schedule is replaced as a whole: to change it, pass --frequency and --at (plus
--day-of-week / --day-of-month for weekly / monthly). Changing the schedule or the
timezone recalculates the next run.

--timezone can be changed but not cleared. One-off policies cannot be updated.

Templates marked xlsxOnly need --format xlsx in the same command. Changing --format
to pdf alone is not checked against the policy's existing templates.`,
		Example: `  # Move a policy to New York time
  incloud report policy update 507f1f77bcf86cd799439011 --timezone America/New_York

  # Run monthly on the 1st at 07:00
  incloud report policy update 507f1f77bcf86cd799439011 \
    --frequency monthly --day-of-month 1 --at 07:00 --period last_30_days`,
		Args: cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "report policy id", "incloud report policy list --name %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := opts.updateBody(cmd)
			if err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			resp, err := client.Put(policiesPath+"/"+args[0], body)
			if err != nil {
				return err
			}
			printPolicyID(f, resp, "updated")

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(resp, f.IO, output)
		},
	}

	opts.registerFlags(cmd, false)

	return cmd
}

func printPolicyID(f *factory.Factory, body []byte, verb string) {
	var resp struct {
		Result struct {
			ID string `json:"_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err == nil && resp.Result.ID != "" {
		fmt.Fprintf(f.IO.ErrOut, "Report policy %s. (id: %s)\n", verb, resp.Result.ID)
	}
}
