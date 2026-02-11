package tickets

import (
	"crypto/rand"
	"math/big"
)

const base62Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// generateID generates a random 3-character base62 ID.
func generateID() string {
	b := make([]byte, 3)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		b[i] = base62Chars[n.Int64()]
	}
	return string(b)
}

// generateUniqueID generates an ID that doesn't conflict with existing IDs in the file.
func generateUniqueID(existing map[string]bool) string {
	for {
		id := generateID()
		if !existing[id] {
			return id
		}
	}
}
