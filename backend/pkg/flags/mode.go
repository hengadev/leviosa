package mode

import "fmt"

type EnvMode uint

const (
	ModeDev EnvMode = iota
	ModeProd
	ModeStaging
)

var modeStr = []string{
	"development",
	"production",
	"staging",
}

func (m *EnvMode) Set(value string) error {
	switch value {
	case "development":
		*m = ModeDev
	case "production":
		*m = ModeProd
	case "staging":
		*m = ModeStaging
	default:
		return fmt.Errorf("mode value can only be 'development', 'production' or 'staging', got : %q", value)
	}
	return nil
}

func (m *EnvMode) String() string {
	return modeStr[*m]
}
