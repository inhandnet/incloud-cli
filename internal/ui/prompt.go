package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// IsTTY returns true if the factory's stdin is an interactive terminal.
func IsTTY(f *factory.Factory) bool {
	file, ok := f.IO.In.(*os.File)
	return ok && (isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd()))
}

// Select displays an interactive selection prompt and returns the chosen value.
// options is a slice of huh.Option created via huh.NewOption(label, value).
// Returns an error if stdin is not a terminal.
func Select[T comparable](f *factory.Factory, title string, options []huh.Option[T]) (T, error) {
	var zero T
	if !IsTTY(f) {
		return zero, fmt.Errorf("terminal is non-interactive; specify the value via flags")
	}

	var result T
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[T]().
				Title(title).
				Options(options...).
				Value(&result),
		),
	).
		WithTheme(huh.ThemeBase()).
		WithOutput(f.IO.ErrOut).
		WithInput(f.IO.In).
		Run()
	if err != nil {
		return zero, err
	}
	return result, nil
}

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
