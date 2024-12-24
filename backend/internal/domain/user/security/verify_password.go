package security

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// VerifyPassword checks if the provided password matches the hash using Argon2id
func (s *SecureUserData) VerifyPassword(password, encodedHash string) (bool, error) {
	// Extract the parameters, salt and hash from the encoded string
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Combine password with pepper
	peppered := append([]byte(password), s.config.Pepper...)

	// Compute the hash of the provided password using the same parameters
	otherHash := argon2.IDKey(
		peppered,
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(hash)),
	)

	// Compare the computed hash with the stored hash
	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}
