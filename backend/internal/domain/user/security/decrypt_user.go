package security

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
	"github.com/GaryHY/event-reservation-app/pkg/errsx"
)

// DecryptUser decrypts sensitive fields of a user
func (s *SecureUserData) DecryptUser(user *models.User) errsx.Map {
	var errs errsx.Map
	if user.EncryptedBirthDate != "" {
		decrypted, err := s.decrypt(user.EncryptedBirthDate)
		if err != nil {
			errs.Set("encrypted birthdate", err)
		}
		parsedTime, err := time.Parse(time.RFC3339, decrypted)
		if err != nil {
			errs.Set("parsing decrypted birthdate", err)
		}
		user.BirthDate = parsedTime
	}
	// Decrypt email if present
	if user.EncryptedEmail != "" {
		decrypted, err := s.decrypt(user.EncryptedEmail)
		if err != nil {
			errs.Set("encrypted email", err)
		}
		user.Email = decrypted
	}

	fields := []struct {
		name  string
		value *string
	}{
		{"lastname", &user.LastName},
		{"firstname", &user.FirstName},
		{"gender", &user.Gender},
		{"telephone", &user.Telephone},
		{"postal code", &user.PostalCode},
		{"city", &user.City},
		{"address 1", &user.Address1},
		{"address 2", &user.Address2},
		{"google ID", &user.GoogleID},
		{"apple ID", &user.AppleID},
	}

	for _, field := range fields {
		if *field.value != "" {
			decrypted, err := s.decrypt(*field.value)
			if err != nil {
				errs.Set(field.name, err)
			}
			*field.value = decrypted
		}
	}

	return errs
}

// decrypt decrypts sensitive data
func (s *SecureUserData) decrypt(encryptedData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.config.EncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
