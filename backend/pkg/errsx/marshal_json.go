package errsx

import (
	"fmt"
	"strings"
)

// MarshalJSON implements the json.Marshaler interface.
func (m Map) MarshalJSON() ([]byte, error) {
	errs := make([]string, 0, len(m))
	for field, err := range m {
		errs = append(errs, fmt.Sprintf("%q:%q", field, err.Error()))
	}
	return []byte(fmt.Sprintf("{%v}", strings.Join(errs, ", "))), nil
}
