package errsx

func (m *Map) Has(key string) bool {
	_, ok := (*m)[key]
	return ok
}
