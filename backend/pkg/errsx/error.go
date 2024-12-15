package errsx

import (
	"fmt"
	"strings"
)

// Error implements the error interface for ErrorMap
func (m Map) Error() string {
	if m == nil {
		return "<nil>"
	}
	var builder strings.Builder
	for field, err := range m {
		builder.WriteString(fmt.Sprintf("%s: %s; ", field, err.Error()))
	}

	return builder.String()
}
