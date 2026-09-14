package auth

import (
	"testing"
	"time"
)

func TestSessionTokenLifecycle(t *testing.T) {
	secret := []byte("very-secure-secret-key-12345678")
	userID := "usr-123"
	email := "admin@polycloud.local"

	// 1. Create valid token
	token, err := CreateSessionToken(secret, userID, email, 1*time.Hour)
	if err != nil {
		t.Fatalf("CreateSessionToken failed: %v", err)
	}

	// 2. Verify valid token
	claims, err := VerifySessionToken(secret, token)
	if err != nil {
		t.Fatalf("VerifySessionToken failed: %v", err)
	}
	if claims.UserID != userID || claims.Email != email {
		t.Errorf("claims mismatch: got %+v", claims)
	}

	// 3. Verify tampered token fails
	tampered := token + "tamper"
	if _, err := VerifySessionToken(secret, tampered); err == nil {
		t.Error("expected error for tampered token, got nil")
	}

	// 4. Verify token signed with different secret fails
	otherSecret := []byte("different-secret-key-87654321000")
	if _, err := VerifySessionToken(otherSecret, token); err == nil {
		t.Error("expected error for wrong secret, got nil")
	}

	// 5. Verify expired token fails
	expiredToken, err := CreateSessionToken(secret, userID, email, -1*time.Minute)
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}
	if _, err := VerifySessionToken(secret, expiredToken); err != ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}
