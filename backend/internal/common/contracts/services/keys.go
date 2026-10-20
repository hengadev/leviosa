package services

import "fmt"

// ServiceAPIKeyPath is the path under which a service key store holds the hash of a
// service's API key; the auth middleware's RequireServiceAuth reads it there.
func ServiceAPIKeyPath(serviceName string) string {
	return fmt.Sprintf("secret/data/services/%s/api-key", serviceName)
}
