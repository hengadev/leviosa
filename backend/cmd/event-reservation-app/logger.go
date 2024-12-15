package main

import (
	"fmt"
	"log/slog"

	"github.com/GaryHY/event-reservation-app/pkg/serverutil/logger"
)

func setLogger() (*slog.Logger, error) {
	if err := logger.SetOptions(opts.mode, &opts.logger.level, &opts.logger.style); err != nil {
		return nil, fmt.Errorf("set logger options: %w", err)
	}
	slogHandler, err := logger.SetHandler(opts.logger.level, opts.logger.style)
	if err != nil {
		return nil, fmt.Errorf("create logger handler: %w", err)
	}
	logger := slog.New(slogHandler)
	return logger, nil
}
