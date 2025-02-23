package errsx

import (
	"errors"
)

// Set associates the given error with the given key. The map is lazily instanciated if it is nil
func (m *Map) Set(field string, msg any) {
	if *m == nil {
		*m = make(Map)
	}
	var err error
	switch msg := msg.(type) {
	case error:
		if msg == nil {
			return
		}
		err = msg
	case string:
		if msg == "" {
			return
		}
		err = errors.New(msg)
	default:
		panic("want error or string message")
	}
	(*m)[field] = err
}
