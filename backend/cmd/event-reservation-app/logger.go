package main

import (
	"fmt"
	"log/slog"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil"
)

func setLogger() (*slog.Logger, error) {
	if err := serverutil.SetLoggerOptions(opts.mode, &opts.logger.level, &opts.logger.style); err != nil {
		return nil, fmt.Errorf("set logger options: %w", err)
	}
	slogHandler, err := serverutil.SetLoggerHandler(opts.logger.level, opts.logger.style)
	if err != nil {
		return nil, fmt.Errorf("create logger handler: %w", err)
	}
	logger := slog.New(slogHandler)
	return logger, nil
}
