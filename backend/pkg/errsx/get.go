package errsx

func (m Map) Get(key string, err error) string {
	if err := m[key]; err != nil {
		return err.Error()
	}
	return ""
}
