package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// HashPassword creates a secure hash of the password
func (s *SecureUserData) hashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, s.config.Argon2Params.SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// Combine password with pepper
	peppered := append([]byte(password), s.config.Pepper...)

	// Generate hash using Argon2id
	hash := argon2.IDKey(
		peppered,
		salt,
		s.config.Argon2Params.Iterations,
		s.config.Argon2Params.Memory,
		s.config.Argon2Params.Parallelism,
		s.config.Argon2Params.KeyLength,
	)

	// Encode params, salt, and hash into a string
	params := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.config.Argon2Params.Memory,
		s.config.Argon2Params.Iterations,
		s.config.Argon2Params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return params, nil
}
