package update

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/build"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

const (
	repoOwner = "inhandnet"
	repoName  = "incloud-cli"
)

func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	var (
		checkOnly   bool
		targetVer   string
		skipConfirm bool
		output      string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update incloud CLI to the latest version",
		Long: `Check for and install newer versions of the incloud CLI.

By default, downloads and installs the latest release from GitHub.
Use --check to only check without installing.`,
		Example: `  # Update to latest version
  incloud update

  # Check for updates without installing
  incloud update --check

  # Check for updates (JSON output, useful for scripts)
  incloud update --check -o json

  # Update to a specific version
  incloud update --version v0.3.0

  # Skip confirmation prompt (for scripts/CI)
  incloud update --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd.Context(), f, updateOptions{
				checkOnly:   checkOnly,
				targetVer:   targetVer,
				skipConfirm: skipConfirm,
				output:      output,
			})
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
	cmd.Flags().StringVar(&targetVer, "version", "", "Update to a specific version (e.g. v0.3.0)")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output format for --check (json)")

	return cmd
}

type updateOptions struct {
	checkOnly   bool
	targetVer   string
	skipConfirm bool
	output      string
}

type checkResult struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable bool   `json:"update_available"`
}

func runUpdate(ctx context.Context, f *factory.Factory, opts updateOptions) error {
	io := f.IO

	currentVer := build.Version
	if currentVer == "dev" {
		return fmt.Errorf("development build cannot determine current version; use 'make install' to update from source")
	}

	fmt.Fprintf(io.ErrOut, "Checking for updates... current: %s\n", currentVer)

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("initializing updater: %w", err)
	}

	repo := selfupdate.ParseSlug(repoOwner + "/" + repoName)

	if opts.targetVer != "" {
		return updateToVersion(ctx, io, updater, repo, opts)
	}

	latest, found, err := updater.DetectLatest(ctx, repo)
	if err != nil {
		return fmt.Errorf("checking for updates: %w\nHint: check your network connection or proxy settings", err)
	}
	if !found {
		return fmt.Errorf("no release found for %s/%s (%s)", runtime.GOOS, runtime.GOARCH, repoOwner+"/"+repoName)
	}

	updateAvailable := latest.GreaterThan(currentVer)

	if opts.checkOnly {
		return printCheckResult(io, opts.output, currentVer, latest.Version(), updateAvailable)
	}

	if !updateAvailable {
		fmt.Fprintf(io.ErrOut, "Already up to date: %s\n", currentVer)
		return nil
	}

	fmt.Fprintf(io.ErrOut, "\nNew version available: %s (released %s)\n",
		latest.Version(), latest.PublishedAt.Format("2006-01-02"))
	if latest.ReleaseNotes != "" {
		fmt.Fprintf(io.ErrOut, "\nRelease notes:\n%s\n", latest.ReleaseNotes)
	}

	if cancelled := confirmUpdate(io, opts.skipConfirm); cancelled {
		return nil
	}

	return doUpdate(ctx, io, updater, latest)
}

func updateToVersion(ctx context.Context, io *iostreams.IOStreams, updater *selfupdate.Updater, repo selfupdate.Repository, opts updateOptions) error {
	ver := opts.targetVer

	release, found, err := updater.DetectVersion(ctx, repo, ver)
	if err != nil {
		return fmt.Errorf("finding version %s: %w", ver, err)
	}
	if !found {
		return fmt.Errorf("version %s not found for %s/%s", ver, runtime.GOOS, runtime.GOARCH)
	}

	fmt.Fprintf(io.ErrOut, "Found version: %s (released %s)\n",
		release.Version(), release.PublishedAt.Format("2006-01-02"))

	if cancelled := confirmUpdate(io, opts.skipConfirm); cancelled {
		return nil
	}

	return doUpdate(ctx, io, updater, release)
}

// confirmUpdate prompts the user unless skipConfirm is set or stdout is not a TTY.
// Returns true if the user cancelled.
func confirmUpdate(io *iostreams.IOStreams, skipConfirm bool) bool {
	if skipConfirm || !io.IsStdoutTTY() {
		return false
	}
	fmt.Fprintf(io.ErrOut, "\nDownload and install? [Y/n] ")
	var answer string
	_, _ = fmt.Fscanln(io.In, &answer)
	if answer != "" && answer != "y" && answer != "Y" && answer != "yes" {
		fmt.Fprintln(io.ErrOut, "Update cancelled.")
		return true
	}
	return false
}

func doUpdate(ctx context.Context, io *iostreams.IOStreams, updater *selfupdate.Updater, release *selfupdate.Release) error {
	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("locating executable: %w", err)
	}

	fmt.Fprintf(io.ErrOut, "Downloading %s...\n", release.AssetName)

	if err := updater.UpdateTo(ctx, release, exe); err != nil {
		return fmt.Errorf("updating binary: %w\nHint: if permission denied, try: sudo incloud update", err)
	}

	fmt.Fprintf(io.ErrOut, "Updated successfully: %s → %s\n", build.Version, release.Version())
	return nil
}

func printCheckResult(io *iostreams.IOStreams, output, current, latest string, available bool) error {
	if output == "json" {
		result := checkResult{
			Current:         current,
			Latest:          latest,
			UpdateAvailable: available,
		}
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintln(io.Out, string(data))
		return nil
	}

	if available {
		fmt.Fprintf(io.ErrOut, "New version available: %s (current: %s)\n", latest, current)
		fmt.Fprintf(io.ErrOut, "Run 'incloud update' to install.\n")
	} else {
		fmt.Fprintf(io.ErrOut, "Already up to date: %s\n", current)
	}
	return nil
}
