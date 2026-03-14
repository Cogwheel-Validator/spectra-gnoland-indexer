package key

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateApiKey() (string, string, [32]byte, error) {
	rawKey := make([]byte, 32)
	_, err := rand.Read(rawKey)
	if err != nil {
		return "", "", [32]byte{}, err
	}
	rawKeyHex := "api_" + hex.EncodeToString(rawKey)
	keyHash := sha256.Sum256([]byte(rawKeyHex))
	keyPrefix := rawKeyHex[:10]
	return rawKeyHex, keyPrefix, keyHash, nil
}
