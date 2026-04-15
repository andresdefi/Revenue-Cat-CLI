package commanddocs_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	rootcmd "github.com/andresdefi/rc/cmd"
	"github.com/andresdefi/rc/internal/commanddocs"
)

func TestCommandReferenceIsCurrent(t *testing.T) {
	got, err := commanddocs.Generate(rootcmd.NewRootCmd())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	path := filepath.Join("..", "..", "docs", "COMMANDS.md")
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s is stale; run `make docs`", path)
	}
}
