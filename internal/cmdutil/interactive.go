package cmdutil

import (
	"fmt"

	"github.com/andresdefi/rc/internal/output"
	"github.com/charmbracelet/huh"
)

// PromptIfEmpty prompts for a value if it is empty and stdout is a TTY.
// If not a TTY and the value is empty, returns an error.
func PromptIfEmpty(value *string, title, placeholder string) error {
	if *value != "" {
		return nil
	}
	if !output.IsTTY() {
		return fmt.Errorf("missing required value: %s", title)
	}
	return huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(value).
		Run()
}

// PromptSelect prompts the user to select from options if value is empty and stdout is a TTY.
// If not a TTY and the value is empty, returns an error.
func PromptSelect(value *string, title string, options []string) error {
	if *value != "" {
		return nil
	}
	if !output.IsTTY() {
		return fmt.Errorf("missing required value: %s", title)
	}
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}
	return huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Value(value).
		Run()
}

// PromptConfirm prompts for yes/no confirmation. Returns false if not a TTY.
func PromptConfirm(title string) (bool, error) {
	if !output.IsTTY() {
		return false, nil
	}
	var confirmed bool
	err := huh.NewConfirm().
		Title(title).
		Value(&confirmed).
		Run()
	return confirmed, err
}
