package config

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func (c *Config) GetSecurity() *SecurityConfig {
	return c.security
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	EncryptionKey []byte // 32 bytes for AES-256
	Pepper        []byte // Additional security for password hashing
	Argon2Params  *Argon2Params
}

// NewSecurityConfig creates a new security configuration
// func (c *Config) setSecurityConfig(context.Context) (*SecurityConfig, error) {
func (c *Config) setSecurityConfig(context.Context) error {
	// Convert hex-encoded key to bytes
	key, err := hex.DecodeString(c.viper.GetString("user.encryption.key"))
	if err != nil {
		return err
	}

	// Generate a random pepper
	pepper := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, pepper); err != nil {
		return err
	}

	c.security = &SecurityConfig{
		EncryptionKey: key,
		Pepper:        pepper,
		Argon2Params:  DefaultArgon2Params(),
	}
	return nil
}

// Argon2Params defines the parameters for Argon2id
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params returns recommended parameters for Argon2id
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}
