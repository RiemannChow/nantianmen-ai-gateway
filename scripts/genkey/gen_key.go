// Command gen_key generates a random 32-byte AES-256 master key suitable for the
// secret.encryption_key field in config.yml. The output is base64-encoded so it
// can be pasted directly into config.yml.
//
// Usage:
//
//	go run ./scripts/gen_key.go
package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(key))
}
