package sqliteutil

import "fmt"

func BuildDSN(connStr string) string {
	return fmt.Sprintf("%s.db", connStr)
}
