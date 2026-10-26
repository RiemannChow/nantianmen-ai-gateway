// Package crypto provides transparent AES-256-GCM encryption for sensitive
// values (API keys, SMTP passwords, webhook secrets) stored in the control
// plane's settings / database.
//
// Ciphertext is stored with the prefix "enc:v1:" so that readers can detect
// whether a value was encrypted or is legacy plaintext. Decrypt is a no-op for
// values that do not carry the prefix, which keeps the system backward
// compatible with rows written before encryption was enabled.
//
// When no master key is configured (the common case in throw-away dev
// environments), Encrypt/Decrypt become transparent pass-throughs: nothing is
// encrypted, but existing call sites keep working. Production private
// deployments must set secret.encryption_key in config.yml.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/eolinker/go-common/cftool"
)

// Prefix marks a stored value as AES-GCM encrypted (version 1).
const Prefix = "enc:v1:"

// keyLength is the AES-256 master-key length in bytes.
const keyLength = 32

// SecretConfig binds the "secret" block of config.yml.
type SecretConfig struct {
	SecretKey string `yaml:"encryption_key"`
}

func init() {
	cftool.Register[SecretConfig]("secret")
}

// ICrypto is a marker interface so the autowire framework initializes imlCrypto
// (and thus installs the configured master key) at startup.
type ICrypto interface {
	OnInit()
}

// Init installs the master key derived from the configured secret string. It is
// intended to be called from an always-initialized component's OnInit (e.g. the
// settings service). A nil/empty key disables encryption (transparent
// pass-through) rather than failing, so dev environments keep working.
func Init(key string) error {
	k, err := resolveKey(key)
	if err != nil {
		defaultGCM = nil
		return err
	}
	block, err := aes.NewCipher(k)
	if err != nil {
		defaultGCM = nil
		return err
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		defaultGCM = nil
		return err
	}
	defaultGCM = g
	return nil
}

// defaultGCM is the package-level AEAD used by Encrypt/Decrypt. It is nil when
// no valid master key is configured.
var defaultGCM cipher.AEAD

// resolveKey normalizes the configured master-key string into a 32-byte key.
// A base64-encoded value is decoded when possible; otherwise the raw string is
// used as-is and padded/truncated to keyLength so callers may also set a plain
// passphrase. An empty string yields an error (no encryption).
func resolveKey(s string) ([]byte, error) {
	if strings.TrimSpace(s) == "" {
		return nil, errors.New("encryption_key is empty")
	}
	// Try base64 first (the recommended form for generated keys).
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == keyLength {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil && len(b) == keyLength {
		return b, nil
	}
	// Fall back to treating the raw string as the key material.
	key := make([]byte, keyLength)
	copy(key, []byte(s))
	return key, nil
}

// SetKey installs a raw 32-byte master key, primarily for tests and explicit
// initialization. It returns an error for keys of the wrong length.
func SetKey(key []byte) error {
	if len(key) != keyLength {
		return errors.New("crypto: master key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	defaultGCM = g
	return nil
}

// IsEnabled reports whether a master key is currently configured.
func IsEnabled() bool {
	return defaultGCM != nil
}

// Encrypt encrypts plaintext with the configured master key and returns a
// prefixed ciphertext. If encryption is disabled it returns the plaintext
// unchanged (no prefix).
func Encrypt(plain string) (string, error) {
	if !IsEnabled() {
		return plain, nil
	}
	return encryptWith(defaultGCM, plain)
}

// Decrypt reverses Encrypt. Values without the prefix (legacy plaintext or
// encryption disabled) are returned unchanged.
func Decrypt(ciphertext string) (string, error) {
	if !strings.HasPrefix(ciphertext, Prefix) {
		return ciphertext, nil
	}
	if !IsEnabled() {
		// Key missing but value was encrypted: cannot recover.
		return "", errors.New("crypto: value is encrypted but no master key is configured")
	}
	// Strip prefix before decoding.
	return decryptWith(defaultGCM, strings.TrimPrefix(ciphertext, Prefix))
}

func encryptWith(g cipher.AEAD, plain string) (string, error) {
	nonce := make([]byte, g.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := g.Seal(nonce, nonce, []byte(plain), nil)
	return Prefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func decryptWith(g cipher.AEAD, encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	ns := g.NonceSize()
	if len(raw) < ns {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, ct := raw[:ns], raw[ns:]
	opened, err := g.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(opened), nil
}
