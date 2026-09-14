package auth

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	key, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if !strings.HasPrefix(key.RawKey, KeyPrefix) {
		t.Errorf("expected prefix %q, got %q", KeyPrefix, key.RawKey)
	}

	if len(key.RawKey) < 30 {
		t.Errorf("raw key is too short: %d", len(key.RawKey))
	}

	if len(key.KeyHash) != 64 {
		t.Errorf("expected 64-char sha256 hex, got len %d: %s", len(key.KeyHash), key.KeyHash)
	}

	// Verify hashing is deterministic
	hash2 := HashKey(key.RawKey)
	if hash2 != key.KeyHash {
		t.Errorf("hash mismatch: expected %s, got %s", key.KeyHash, hash2)
	}

	// Verify display prefix mask
	if !strings.HasPrefix(key.KeyPrefix, KeyPrefix) || !strings.Contains(key.KeyPrefix, "...") {
		t.Errorf("unexpected display prefix: %s", key.KeyPrefix)
	}
}
