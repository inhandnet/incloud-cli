package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// Confirm displays an interactive yes/no prompt and returns the user's choice.
// Returns an error if stdin is not a terminal (use --yes to skip).
func Confirm(f *factory.Factory, message string) (bool, error) {
	file, ok := f.IO.In.(*os.File)
	if !ok || (!isatty.IsTerminal(file.Fd()) && !isatty.IsCygwinTerminal(file.Fd())) {
		return false, fmt.Errorf("terminal is non-interactive; use --yes to confirm")
	}

	var confirmed bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	).
		WithTheme(huh.ThemeBase()).
		WithOutput(f.IO.ErrOut).
		WithInput(f.IO.In).
		Run()
	if err != nil {
		return false, err
	}
	return confirmed, nil
}
