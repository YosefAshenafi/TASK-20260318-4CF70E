// Package pii provides AES-256-GCM encryption for candidate PII at rest (design.md §11.1).
package pii

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	// KeySize is the required AES-256 key length in bytes.
	KeySize = 32
	// nonceSize is the GCM standard nonce length.
	nonceSize = 12
)

var (
	// ErrNoKey means PII_AES_KEY_HEX was not set or was invalid.
	ErrNoKey = errors.New("PII AES key not configured")
	// ErrDecrypt means ciphertext could not be authenticated or decoded.
	ErrDecrypt = errors.New("PII decrypt failed")
)

// Cipher encrypts and decrypts PII blobs stored in VARBINARY columns.
type Cipher struct {
	key []byte
}

// NewCipherFromHex loads a 32-byte key from 64 hex characters (PII_AES_KEY_HEX).
func NewCipherFromHex(hexKey string) (*Cipher, error) {
	if hexKey == "" {
		return nil, nil
	}
	raw, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("PII_AES_KEY_HEX: %w", err)
	}
	if len(raw) != KeySize {
		return nil, fmt.Errorf("PII_AES_KEY_HEX must decode to %d bytes, got %d", KeySize, len(raw))
	}
	return &Cipher{key: append([]byte(nil), raw...)}, nil
}

// Valid reports whether encryption is available.
func (c *Cipher) Valid() bool {
	return c != nil && len(c.key) == KeySize
}

// Encrypt returns nonce||ciphertext (includes GCM tag). Empty plaintext yields nil, nil.
func (c *Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, nil
	}
	if !c.Valid() {
		return nil, ErrNoKey
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, nonceSize+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

// Decrypt reverses Encrypt. Empty blob yields empty plaintext.
func (c *Cipher) Decrypt(blob []byte) ([]byte, error) {
	if len(blob) == 0 {
		return nil, nil
	}
	if !c.Valid() {
		return nil, ErrNoKey
	}
	if len(blob) < nonceSize {
		return nil, ErrDecrypt
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := blob[:nonceSize]
	ct := blob[nonceSize:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrDecrypt
	}
	return pt, nil
}

// EncryptString encrypts UTF-8 text to a blob.
func (c *Cipher) EncryptString(s string) ([]byte, error) {
	return c.Encrypt([]byte(s))
}

// DecryptString decrypts to UTF-8 string.
func (c *Cipher) DecryptString(blob []byte) (string, error) {
	b, err := c.Decrypt(blob)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DigestHex returns deterministic HMAC-SHA256 digest for normalized duplicate keys.
func (c *Cipher) DigestHex(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !c.Valid() {
		return "", ErrNoKey
	}
	mac := hmac.New(sha256.New, c.key)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
