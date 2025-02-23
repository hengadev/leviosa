package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func SetHandler(level, style string) (slog.Handler, error) {
	logLevel, ok := loggerLevels[loggerLevel(level)]
	if !ok {
		return nil, fmt.Errorf("invalid log level supplied: %q", level)
	}
	logStyle := loggerStyle(style)
	var slogHandler slog.Handler
	switch logStyle {
	case JSON:
		// TODO: write logs to some file to use with ELK stack
		slogHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case Text:
		slogHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case Dev:
		slogHandler = NewDevHandler(os.Stdout, logLevel)
	default:
		return nil, fmt.Errorf("invalid log style supplied: %q", style)
	}
	return slogHandler, nil
}
