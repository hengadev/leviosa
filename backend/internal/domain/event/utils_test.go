package eventService_test

import (
	"encoding/hex"
	"testing"

	"github.com/GaryHY/leviosa/pkg/config"
)

func prepareEncryptionConfig(t *testing.T) *config.SecurityConfig {
	t.Helper()
	key, err := hex.DecodeString("fb259f76a783dc70d8e7c73a4c056b496dcb52942248d467e81c931609aef4f7")
	if err != nil {
		t.Errorf("DecodeString for encryption %s", err)
	}
	return &config.SecurityConfig{
		EncryptionKey: key,                          // []byte
		Pepper:        []byte(""),                   // []byte
		Argon2Params:  config.DefaultArgon2Params(), // *Argon2Params
	}
}
