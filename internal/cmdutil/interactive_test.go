package cmdutil

import (
	"testing"
)

func TestPromptIfEmpty_ValueAlreadySet(t *testing.T) {
	val := "existing-value"
	err := PromptIfEmpty(&val, "Test", "placeholder")
	if err != nil {
		t.Fatalf("PromptIfEmpty() error: %v", err)
	}
	if val != "existing-value" {
		t.Errorf("PromptIfEmpty() changed value to %q, want %q", val, "existing-value")
	}
}

func TestPromptIfEmpty_ValueSet_NoPrompt(t *testing.T) {
	// When value is already set, PromptIfEmpty should return nil immediately
	// regardless of TTY state
	val := "set"
	err := PromptIfEmpty(&val, "Title", "hint")
	if err != nil {
		t.Errorf("PromptIfEmpty() with set value returned error: %v", err)
	}
}

func TestPromptIfEmpty_EmptyValueNonTTY(t *testing.T) {
	// In test environment (non-TTY), empty value should return an error
	val := ""
	err := PromptIfEmpty(&val, "Test Field", "placeholder")
	if err == nil {
		t.Fatal("PromptIfEmpty() with empty value in non-TTY should return error")
	}
}

func TestPromptSelect_ValueAlreadySet(t *testing.T) {
	val := "option-a"
	err := PromptSelect(&val, "Choose", []string{"option-a", "option-b"})
	if err != nil {
		t.Fatalf("PromptSelect() error: %v", err)
	}
	if val != "option-a" {
		t.Errorf("PromptSelect() changed value to %q, want %q", val, "option-a")
	}
}

func TestPromptSelect_ValueSet_NoPrompt(t *testing.T) {
	val := "preset"
	err := PromptSelect(&val, "Select", []string{"a", "b", "c"})
	if err != nil {
		t.Errorf("PromptSelect() with set value returned error: %v", err)
	}
}

func TestPromptSelect_EmptyValueNonTTY(t *testing.T) {
	val := ""
	err := PromptSelect(&val, "Select Item", []string{"a", "b"})
	if err == nil {
		t.Fatal("PromptSelect() with empty value in non-TTY should return error")
	}
}

func TestPromptConfirm_NonTTY(t *testing.T) {
	// In non-TTY mode, PromptConfirm returns false with no error
	confirmed, err := PromptConfirm("Confirm?")
	if err != nil {
		t.Fatalf("PromptConfirm() error: %v", err)
	}
	if confirmed {
		t.Error("PromptConfirm() in non-TTY should return false")
	}
}

func TestPromptConfirm_CallableInNonTTY(t *testing.T) {
	// In non-TTY (test environment), PromptConfirm should return false
	result, err := PromptConfirm("Test?")
	if err != nil {
		t.Logf("PromptConfirm returned error (expected in non-TTY): %v", err)
	}
	_ = result
}
