package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	KeyPrefix = "plc_live_"
	KeyRandomBytes = 24 // 48 hex chars + 9 prefix chars = 57 chars
)

// GeneratedKey memuat raw token (hanya diperlihatkan 1x), prefix display, dan hash DB.
type GeneratedKey struct {
	RawKey    string
	KeyPrefix string
	KeyHash   string
}

// GenerateAPIKey menghasilkan raw key acak kriptografis, prefix untuk identifikasi, dan SHA-256 hash.
func GenerateAPIKey() (*GeneratedKey, error) {
	b := make([]byte, KeyRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate random bytes: %w", err)
	}

	randomPart := hex.EncodeToString(b)
	raw := KeyPrefix + randomPart
	hash := HashKey(raw)

	// Prefix yang disimpan untuk display (misal: "plc_live_a1b2c3d4...")
	displayPrefix := raw[:16] + "..." + raw[len(raw)-4:]

	return &GeneratedKey{
		RawKey:    raw,
		KeyPrefix: displayPrefix,
		KeyHash:   hash,
	}, nil
}

// HashKey menghitung SHA-256 hash heksadesimal dari raw API key.
func HashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
