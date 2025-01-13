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
	"time"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/pkg/errsx"
)

// EncryptUser encrypts sensitive fields in the provided user model and clears the original plaintext values.
//
// Parameters:
//   - user: A pointer to the `models.User` struct containing fields to be encrypted, such as birthdate,
//     last name, first name, email, and password.
//
// Returns:
//   - errsx.Map: A map containing errors for any encryption failures. The map contains key-value pairs
//     where the key is the name of the field (e.g., "encrypt birthdate") and the value is the corresponding error.
//     If no errors occur, an empty map is returned.
func (s *SecureUserData) EncryptUser(user *models.User) errsx.Map {
	var errs errsx.Map
	// TODO: do the same later with created at and loggedinat
	if !user.BirthDate.IsZero() {
		dateStr := user.BirthDate.Format(time.RFC3339)
		encrypted, pbms := s.encrypt(dateStr)
		if len(pbms) > 0 {
			errs.Set("encrypt birthdate", pbms.Error())
		}
		// Store encrypted string back in a temp field
		user.EncryptedBirthDate = encrypted
		// Zero out the original time.Time
		user.BirthDate = time.Time{}
	}

	fields := []struct {
		value *string
	}{
		{&user.LastName},
		{&user.FirstName},
		{&user.Gender},
		{&user.Telephone},
		{&user.PostalCode},
		{&user.City},
		{&user.Address1},
		{&user.Address2},
		{&user.GoogleID},
		{&user.AppleID},
	}

	for _, field := range fields {
		if *field.value != "" {
			encrypted, pbms := s.encrypt(*field.value)
			if len(pbms) > 0 {
				errs.Set("encrypt field", pbms.Error())
			}
			*field.value = encrypted
		}
	}
	// Handle email specially - we need both a hash for searching and encrypted value
	if user.Email != "" {
		user.EmailHash = HashEmail(user.Email)

		encrypted, pbms := s.encrypt(user.Email)
		if len(pbms) > 0 {
			errs.Set("encrypt field", pbms.Error())
		}
		user.EncryptedEmail = encrypted
		user.Email = "" // Clear the plain text email
	}

	// Hash password if present
	if user.Password != "" {
		hash, err := s.hashPassword(user.Password)
		if err != nil {
			errs.Set("hash password", err)
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
		errs.Set("aes create cypher", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		errs.Set("ciper create GCM", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		errs.Set("gcm nonce", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), errs
}
