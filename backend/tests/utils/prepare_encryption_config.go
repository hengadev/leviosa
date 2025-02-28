package test

import (
	"encoding/hex"

	"github.com/hengadev/leviosa/pkg/config"
)

func PrepareEncryptionConfig() *config.SecurityConfig {
	key, _ := hex.DecodeString("fb259f76a783dc70d8e7c73a4c056b496dcb52942248d467e81c931609aef4f7")
	return &config.SecurityConfig{
		EncryptionKey: key,                          // []byte
		Pepper:        []byte(""),                   // []byte
		Argon2Params:  config.DefaultArgon2Params(), // *Argon2Params
	}
}
