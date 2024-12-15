package logger

import (
	"flag"
	"fmt"

	"github.com/GaryHY/event-reservation-app/pkg/flags"
)

func SetOptions(env mode.EnvMode, level, style *string) error {
	if env == mode.ModeDev {
		*level = string(Debug)
		*style = string(Dev)
		return nil
	}
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
