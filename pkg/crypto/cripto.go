package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/danielmesquitta/openfinance/internal/config"
)

type Crypto struct {
	env *config.Env
}

func NewCrypto(env *config.Env) *Crypto {
	return &Crypto{env: env}
}

func (c *Crypto) Decrypt(hashed string) (string, error) {
	key, err := hex.DecodeString(c.env.HashSecret)
	if err != nil {
		return "", fmt.Errorf("error decoding hash secret: %w", err)
	}

	ciphertext, err := base64.URLEncoding.DecodeString(hashed)
	if err != nil {
		return "", fmt.Errorf("error decoding hash: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error creating new cipher: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return "", fmt.Errorf("error checking ciphertext length: %w", err)
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(ciphertext, ciphertext)

	decrypted := string(ciphertext)

	return decrypted, nil
}

func (c *Crypto) Encrypt(plaintext string) (string, error) {
	key, err := hex.DecodeString(c.env.HashSecret)
	if err != nil {
		return "", fmt.Errorf("error decoding hash secret: %w", err)
	}

	plaintextInBytes := []byte(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("error creating new cipher: %w", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintextInBytes))
	iv := ciphertext[:aes.BlockSize]
	_, err = io.ReadFull(rand.Reader, iv)
	if err != nil {
		return "", fmt.Errorf("error reading random bytes: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintextInBytes)

	hashed := base64.URLEncoding.EncodeToString(ciphertext)

	return hashed, nil
}
