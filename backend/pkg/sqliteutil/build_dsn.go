package sqliteutil

import "fmt"

func BuildDSN(connStr string) string {
	return fmt.Sprintf("%s.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", connStr)
}
