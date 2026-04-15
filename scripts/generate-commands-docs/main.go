package main

import (
	"fmt"
	"os"
	"path/filepath"

	rootcmd "github.com/andresdefi/rc/cmd"
	"github.com/andresdefi/rc/internal/commanddocs"
)

func main() {
	root := rootcmd.NewRootCmd()
	data, err := commanddocs.Generate(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate command docs: %v\n", err)
		os.Exit(1)
	}

	path := filepath.Join("docs", "COMMANDS.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "create docs dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
		os.Exit(1)
	}
}
