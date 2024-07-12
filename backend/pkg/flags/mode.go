package mode

import "fmt"

type EnvMode uint

const (
	ModeDev EnvMode = iota
	ModeProd
	ModeStaging
)

func (m *EnvMode) Set(value string) error {
	switch value {
	case "dev":
		*m = ModeDev
	case "prod":
		*m = ModeProd
	case "staging":
		*m = ModeStaging
	default:
		return fmt.Errorf("mode value can only be 'dev', 'prod' or 'staging', got : %q", *m)
	}
	return nil
}
func (m *EnvMode) String() string {
	modes := []string{
		"dev",
		"prod",
		"staging",
	}
	return modes[*m]
}
