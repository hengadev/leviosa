package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// Encrypt sensitive fields
// func (s *SecureUserData) EncryptUser(user *models.User) error {
func (s *SecureUserData) EncryptUser(user *models.User) errsx.Map {
	var errs errsx.Map
	fields := []struct {
		value *string
	}{
		{&user.BirthDate},
		{&user.LastName},
		{&user.FirstName},
		{&user.Gender},
		{&user.Telephone},
		{&user.PostalCode},
		{&user.City},
		{&user.Address1},
		{&user.Address2},
	}

	for _, field := range fields {
		if *field.value != "" {
			// encrypted, err := s.encrypt(*field.value)
			encrypted, pbms := s.encrypt(*field.value)
			if len(pbms) > 0 {
				errs.Set("encrypt field", pbms.Error())
				// return errs
			}
			*field.value = encrypted
		}
	}
	// Handle email specially - we need both a hash for searching and encrypted value
	if user.Email != "" {
		// Create a hash for searching
		// emailHash := sha256.Sum256([]byte(strings.ToLower(user.Email)))
		// user.EmailHash = hex.EncodeToString(emailHash[:])
		user.EmailHash = HashEmail(user.Email)

		// Encrypt the actual email
		// encrypted, err := s.encrypt(user.Email)
		encrypted, pbms := s.encrypt(user.Email)
		// if err != nil {
		if len(pbms) > 0 {
			errs.Set("encrypt field", pbms.Error())
			// return err
		}
		user.EncryptedEmail = encrypted
		user.Email = "" // Clear the plain text email
	}

	// Hash password if present
	if user.Password != "" {
		hash, err := s.hashPassword(user.Password)
		if err != nil {
			errs.Set("hash password", err)
			// return err
		}
		user.PasswordHash = hash
		user.Password = "" // Clear plain text password
	}

	return nil
}

func HashEmail(email string) string {
	emailHash := sha256.Sum256([]byte(strings.ToLower(email)))
	return hex.EncodeToString(emailHash[:])
}

// encrypt is a helper function for the EncryptUser function
func (s *SecureUserData) encrypt(data string) (string, errsx.Map) {
	var errs errsx.Map
	block, err := aes.NewCipher(s.config.EncryptionKey)
	if err != nil {
		// return "", err
		errs.Set("aes create cypher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		// return "", err
		errs.Set("ciper create GCM", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		// return "", err
		errs.Set("gcm nonce", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	// return base64.StdEncoding.EncodeToString(ciphertext), nil
	return base64.StdEncoding.EncodeToString(ciphertext), errs
}
