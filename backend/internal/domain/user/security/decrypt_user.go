package security

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user/models"
)

// DecryptUser decrypts sensitive fields of a user
func (s *SecureUserData) DecryptUser(user *models.User) error {
	if user.EncryptedBirthDate != "" {
		decrypted, err := s.decrypt(user.EncryptedBirthDate)
		if err != nil {
			return err
		}
		parsedTime, err := time.Parse(time.RFC3339, decrypted)
		if err != nil {
			return err
		}
		user.BirthDate = parsedTime
	}
	// Decrypt email if present
	if user.EncryptedEmail != "" {
		decrypted, err := s.decrypt(user.EncryptedEmail)
		if err != nil {
			return err
		}
		user.Email = decrypted
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
	}

	for _, field := range fields {
		if *field.value != "" {
			decrypted, err := s.decrypt(*field.value)
			if err != nil {
				return err
			}
			*field.value = decrypted
		}
	}

	return nil
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
