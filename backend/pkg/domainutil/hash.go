package domainutil

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashWithSalt(value string, salt string) string {
	h := sha256.New()
	h.Write([]byte(value + salt))
	return hex.EncodeToString(h.Sum(nil))
}
