// Package shortener derives short ids from URLs.
package shortener

import (
	"crypto/sha256"
	"encoding/hex"
)

// ShortID returns the first 5 hex characters of the SHA-256 of url.
func ShortID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[0:5]
}
