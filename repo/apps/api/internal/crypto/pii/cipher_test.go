package pii

import (
	"encoding/hex"
	"testing"
)

func TestCipher_roundTrip(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}
	c := &Cipher{key: key}
	plain := []byte("13800138000")
	ct, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if len(ct) < nonceSize+1 {
		t.Fatalf("ciphertext too short")
	}
	got, err := c.Decrypt(ct)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestNewCipherFromHex(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	h := hex.EncodeToString(key)
	c, err := NewCipherFromHex(h)
	if err != nil || c == nil || !c.Valid() {
		t.Fatalf("cipher: %v valid=%v", err, c != nil && c.Valid())
	}
	if _, err := c.EncryptString("test"); err != nil {
		t.Fatal(err)
	}
}

func TestNewCipherFromHex_empty(t *testing.T) {
	c, err := NewCipherFromHex("")
	if err != nil || c != nil {
		t.Fatalf("want nil cipher, err=%v", err)
	}
}

func TestEncrypt_noKey(t *testing.T) {
	var c *Cipher
	if _, err := c.Encrypt([]byte("x")); err == nil || err != ErrNoKey {
		t.Fatalf("err=%v", err)
	}
}
