package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
)

// mode is one way of running the binary: `leviosa [mode] [args...]`. Each mode lives in
// its own file in this package and is registered in modes below.
type mode func(ctx context.Context, args []string) error

// modes maps each run mode name to its entry point. With no argument the binary serves.
var modes = map[string]mode{
	"check-email":     checkEmail,
	"check-sms":       checkSMS,
	"serve":           serve,
	"migrate":         migrate,
	"bootstrap-admin": bootstrapAdmin,
	"verify":          verify,
}

func main() {
	ctx := context.Background()
	if err := dispatch(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitCode(err))
	}
}

// exitCode is the process exit status for a mode's error: 1, unless the error
// carries its own positive status (a mode with distinct failures, like verify).
func exitCode(err error) int {
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) && coded.ExitCode() > 0 {
		return coded.ExitCode()
	}
	return 1
}

func dispatch(ctx context.Context, args []string) error {
	name := "serve"
	if len(args) > 0 {
		name, args = args[0], args[1:]
	}

	run, ok := modes[name]
	if !ok {
		names := make([]string, 0, len(modes))
		for n := range modes {
			names = append(names, n)
		}
		slices.Sort(names)
		return fmt.Errorf("unknown mode %q (available: %s)", name, strings.Join(names, ", "))
	}
	return run(ctx, args)
}

func setupLogger() *slog.Logger {
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
