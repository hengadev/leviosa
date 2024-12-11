package serverutil

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/GaryHY/event-reservation-app/pkg/flags"
)

type loggerLevel string
type loggerStyle string

const (
	Info  loggerLevel = "info"
	Debug             = "debug"
	Error             = "error"
	Warn              = "warn"
)

const (
	JSON loggerStyle = "json"
	Text             = "text"
	Dev              = "dev"
)

var loggerLevels = map[loggerLevel]slog.Level{
	Info:  slog.LevelInfo,
	Debug: slog.LevelDebug,
	Error: slog.LevelError,
	Warn:  slog.LevelWarn,
}

func SetLoggerOptions(env mode.EnvMode, level, style *string) error {
	if env == mode.ModeDev {
		*level = string(Debug)
		*style = string(Dev)
		return nil
	}
	fmt.Println("outside the if that interest me")
	var defaultLevel string
	switch env {
	case mode.ModeProd:
		defaultLevel = Error
	case mode.ModeStaging:
		defaultLevel = string(Info)
	case mode.ModeDev:
	default:
		return fmt.Errorf("APP_ENV does not exist")
	}

	flag.StringVar(level, "logger-level", defaultLevel, "Set logger level")
	flag.StringVar(style, "logger-style", string(JSON), "Set logger style")
	flag.Parse()

	return nil
}

// TODO: make sure to write the logs to the right depending on the style that you have
// I want to use the ELK stack to handle my logs
func SetLoggerHandler(level, style string) (slog.Handler, error) {
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
