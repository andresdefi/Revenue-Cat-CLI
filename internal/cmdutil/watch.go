package cmdutil

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/andresdefi/rc/internal/output"
)

const DefaultWatchInterval = 5 * time.Second

// Watch repeatedly calls fn at the given interval until the context is canceled
// or Ctrl+C is pressed. Errors are printed to stderr but do not stop the loop.
func Watch(ctx context.Context, interval time.Duration, fn func(ctx context.Context) error) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	if interval <= 0 {
		interval = DefaultWatchInterval
	}

	for {
		// Clear screen if TTY
		if output.IsTTY() {
			fmt.Fprint(os.Stdout, "\033[2J\033[H")
		}

		if err := fn(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		if output.IsTTY() {
			fmt.Fprintf(os.Stderr, "\nRefreshing every %s... (Ctrl+C to stop)\n", interval)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
		}
	}
}
