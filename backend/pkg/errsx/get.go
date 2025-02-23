package errsx

func (m Map) Get(key string) string {
	if err := m[key]; err != nil {
		return err.Error()
	}
	return ""
}
