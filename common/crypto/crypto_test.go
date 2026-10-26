package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, keyLength)
	for i := range key {
		key[i] = byte(i)
	}
	if err := SetKey(key); err != nil {
		t.Fatalf("SetKey: %v", err)
	}

	plain := "sk-1234567890abcdef"
	cipher, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if cipher == plain {
		t.Fatal("ciphertext must differ from plaintext")
	}
	if !hasPrefix(cipher, Prefix) {
		t.Fatalf("ciphertext must carry prefix, got %q", cipher)
	}

	got, err := Decrypt(cipher)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plain {
		t.Fatalf("round-trip mismatch: got %q want %q", got, plain)
	}
}

func TestEncryptEmptyAndUnicode(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, keyLength)
	if err := SetKey(key); err != nil {
		t.Fatalf("SetKey: %v", err)
	}
	for _, plain := range []string{"", "中文秘钥 secret ✓"} {
		c, err := Encrypt(plain)
		if err != nil {
			t.Fatalf("Encrypt(%q): %v", plain, err)
		}
		g, err := Decrypt(c)
		if err != nil {
			t.Fatalf("Decrypt(%q): %v", plain, err)
		}
		if g != plain {
			t.Fatalf("round-trip mismatch for %q: got %q", plain, g)
		}
	}
}

func TestDecryptNonPrefixedIsPassthrough(t *testing.T) {
	// No key configured (fresh package state is hard to reset, so just verify
	// that a value without the prefix is returned unchanged regardless of key).
	legacy := "already-plaintext-secret"
	got, err := Decrypt(legacy)
	if err != nil {
		t.Fatalf("Decrypt legacy: %v", err)
	}
	if got != legacy {
		t.Fatalf("legacy plaintext must pass through, got %q", got)
	}
}

func TestDecryptTamperedFails(t *testing.T) {
	key := bytes.Repeat([]byte{0x07}, keyLength)
	if err := SetKey(key); err != nil {
		t.Fatalf("SetKey: %v", err)
	}
	c, err := Encrypt("topsecret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	// Flip a byte in the base64 payload (after the prefix).
	body := []byte(c[len(Prefix):])
	body[0] ^= 0xff
	tampered := Prefix + string(body)
	if _, err := Decrypt(tampered); err == nil {
		t.Fatal("tampered ciphertext must fail to decrypt")
	}
}

func TestSetKeyWrongLength(t *testing.T) {
	if err := SetKey([]byte("tooshort")); err == nil {
		t.Fatal("SetKey with wrong length must error")
	}
}

func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}
