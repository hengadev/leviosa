package logger

import (
	"fmt"
	"log/slog"
	"os"
)

// TODO: make sure to write the logs right depending on the style that you have
// I want to use the ELK stack to handle my logs
func SetHandler(level, style string) (slog.Handler, error) {
	logLevel, ok := loggerLevels[loggerLevel(level)]
	if !ok {
		return nil, fmt.Errorf("invalid log level supplied: %q", level)
	}
	logStyle := loggerStyle(style)
	var slogHandler slog.Handler
	switch logStyle {
	case JSON:
		slogHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	case Text:
		slogHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
		// TODO: add the thing when defined
	case Dev:
		slogHandler = NewDevHandler(os.Stdout, logLevel)
	default:
		return nil, fmt.Errorf("invalid log style supplied: %q", style)
	}
	return slogHandler, nil
}
