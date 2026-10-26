package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
)

// publicKey is the vendor's Ed25519 public key. A License is valid only if its
// signature was produced by the matching private key (held by the vendor).
//
// NOTE: This is a DEMO key baked into the open-source build so the feature can
// be exercised end-to-end. For a real commercial product, replace this with the
// vendor's production public key and keep the private key off the repository.
const publicKeyB64 = "n6jEm7Yxfs/w6DF80ng5UC6oMxTCnRvq4q3wCqvcZc8="

var publicKey ed25519.PublicKey

func init() {
	b, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		panic("license: invalid embedded public key: " + err.Error())
	}
	publicKey = ed25519.PublicKey(b)
}

// sign produces a base64 signature over the canonical JSON of the signable
// fields. priv is the vendor private key (used only by the license generator).
func sign(priv ed25519.PrivateKey, l *License) (string, error) {
	raw, err := json.Marshal(l.signableCopy())
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, raw)
	return base64.StdEncoding.EncodeToString(sig), nil
}

// Sign signs a license with the given base64-encoded Ed25519 private key and
// returns the license with its Signature field populated. Used by the license
// generator (scripts/gen_license.go), which holds the vendor private key.
func Sign(privB64 string, l *License) (*License, error) {
	privRaw, err := base64.StdEncoding.DecodeString(privB64)
	if err != nil {
		return nil, err
	}
	if len(privRaw) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	sig, err := sign(ed25519.PrivateKey(privRaw), l)
	if err != nil {
		return nil, err
	}
	l.Signature = sig
	return l, nil
}

// verify checks the license signature against the embedded public key.
func verify(l *License) error {
	if l.Signature == "" {
		return errors.New("missing signature")
	}
	sig, err := base64.StdEncoding.DecodeString(l.Signature)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(l.signableCopy())
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, raw, sig) {
		return errors.New("signature mismatch")
	}
	return nil
}
