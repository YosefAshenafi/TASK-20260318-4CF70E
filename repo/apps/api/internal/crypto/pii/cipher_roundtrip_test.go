package pii

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestCipher_roundTrip_multipleValues(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}
	c := &Cipher{key: key}

	values := []string{
		"13800138000",
		"alice@example.com",
		"110101199001011234",
		"",
		strings.Repeat("A", 1024),
	}
	for _, v := range values {
		if v == "" {
			continue
		}
		t.Run(v[:min(len(v), 20)], func(t *testing.T) {
			ct, err := c.Encrypt([]byte(v))
			if err != nil {
				t.Fatal(err)
			}
			if len(ct) == 0 {
				t.Fatal("ciphertext should not be empty")
			}
			pt, err := c.Decrypt(ct)
			if err != nil {
				t.Fatal(err)
			}
			if string(pt) != v {
				t.Fatalf("round-trip failed: got %q want %q", string(pt), v)
			}
		})
	}
}

func TestCipher_differentCiphertexts(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i + 42)
	}
	c := &Cipher{key: key}
	plain := []byte("13800138000")

	ct1, _ := c.Encrypt(plain)
	ct2, _ := c.Encrypt(plain)
	if hex.EncodeToString(ct1) == hex.EncodeToString(ct2) {
		t.Fatal("AES-GCM should produce different ciphertexts per call (random nonce)")
	}

	pt1, _ := c.Decrypt(ct1)
	pt2, _ := c.Decrypt(ct2)
	if string(pt1) != string(pt2) || string(pt1) != string(plain) {
		t.Fatal("both should decrypt to same plaintext")
	}
}

func TestCipher_wrongKey(t *testing.T) {
	key1 := make([]byte, KeySize)
	for i := range key1 {
		key1[i] = byte(i)
	}
	key2 := make([]byte, KeySize)
	for i := range key2 {
		key2[i] = byte(i + 1)
	}
	c1 := &Cipher{key: key1}
	c2 := &Cipher{key: key2}

	ct, err := c1.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c2.Decrypt(ct)
	if err == nil {
		t.Fatal("expected decryption to fail with wrong key")
	}
}

func TestCipher_tampered(t *testing.T) {
	key := make([]byte, KeySize)
	c := &Cipher{key: key}
	ct, _ := c.Encrypt([]byte("secret"))
	ct[len(ct)-1] ^= 0xFF
	_, err := c.Decrypt(ct)
	if err == nil {
		t.Fatal("expected decryption to fail for tampered ciphertext")
	}
}

func TestCipher_EncryptString_DecryptString(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i + 10)
	}
	c := &Cipher{key: key}
	ct, err := c.EncryptString("phone: 13800138000")
	if err != nil {
		t.Fatal(err)
	}
	pt, err := c.DecryptString(ct)
	if err != nil {
		t.Fatal(err)
	}
	if pt != "phone: 13800138000" {
		t.Fatalf("got %q", pt)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
