package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidSession = errors.New("invalid session token")
	ErrSessionExpired = errors.New("session token expired")
)

type SessionClaims struct {
	UserID    string `json:"uid"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"exp"`
}

// CreateSessionToken membuat signed session token berbasis HMAC-SHA256 yang aman dan tamper-proof.
func CreateSessionToken(secret []byte, userID, email string, duration time.Duration) (string, error) {
	claims := SessionClaims{
		UserID:    userID,
		Email:     email,
		ExpiresAt: time.Now().Add(duration).Unix(),
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal session claims: %w", err)
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payloadB64))
	sig := hex.EncodeToString(mac.Sum(nil))

	return payloadB64 + "." + sig, nil
}

// VerifySessionToken memvalidasi tanda tangan HMAC-SHA256 dan waktu kedaluwarsa token sesi.
func VerifySessionToken(secret []byte, tokenStr string) (*SessionClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidSession
	}

	payloadB64, providedSig := parts[0], parts[1]

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payloadB64))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return nil, ErrInvalidSession
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, ErrInvalidSession
	}

	var claims SessionClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidSession
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrSessionExpired
	}

	return &claims, nil
}
