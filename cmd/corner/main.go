package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"corner/internal/config"
	"corner/internal/pipeline"
)

func main() {
	if err := run(os.Args, os.Stdout, os.Stderr); err != nil {
		// flag has already printed its own diagnostic for the parse-failure
		// and help-requested cases; avoid duplicating it here. For
		// usage-requested we additionally exit successfully, matching the
		// convention of most CLIs.
		switch {
		case errors.Is(err, config.ErrUsageRequested):
			return
		case errors.Is(err, config.ErrFlagParse):
			os.Exit(2)
		default:
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	configureLogging(stderr)
	if len(args) == 0 {
		return errors.New("missing argv[0]")
	}

	cfg, err := config.Load(args[0], args[1:], stderr)
	if err != nil {
		return err
	}
	if cfg.Moo {
		_, _ = fmt.Fprint(stdout, mooASCII)
		return nil
	}
	stats, err := pipeline.NewProcessor(cfg).Run()
	if stats.Processed > 0 || stats.Failed > 0 {
		if stats.Failed == 0 {
			_, _ = fmt.Fprintf(stdout, "Processed %d file(s)\n", stats.Processed)
		} else {
			_, _ = fmt.Fprintf(stdout, "Processed %d file(s), %d failed\n", stats.Processed, stats.Failed)
		}
	}
	return err
}

// configureLogging routes slog through a text handler aimed at humans on a
// CLI: timestamps off (we already have shell timing), info level, output to
// the same writer the binary uses for diagnostics.
func configureLogging(w io.Writer) {
	h := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	slog.SetDefault(slog.New(h))
}
