// Package auth menangani OAuth provider, enkripsi token, dan lifecycle refresh.
package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// TokenSet = bentuk token yang disimpan (dan yang dimengerti rclone).
// Field JSON-nya sengaja mengikuti format token rclone: {access_token, token_type,
// refresh_token, expiry} — supaya bisa langsung disuntik ke `config/create`.
type TokenSet struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// Expired melaporkan token sudah lewat masa berlaku (dengan margin aman 60 detik).
func (t TokenSet) Expired() bool {
	if t.Expiry.IsZero() {
		return false // provider tanpa expiry (mis. token statis)
	}
	return time.Now().Add(60 * time.Second).After(t.Expiry)
}

// Cipher membungkus AES-256-GCM untuk token at-rest (ADR-006).
type Cipher struct {
	aead cipher.AEAD
}

func NewCipher(key [32]byte) (*Cipher, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

// Seal mengenkripsi TokenSet jadi blob untuk kolom account_tokens.enc_payload.
// Nonce ditaruh di depan ciphertext.
func (c *Cipher) Seal(t TokenSet) ([]byte, error) {
	plain, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("marshal token: %w", err)
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	return c.aead.Seal(nonce, nonce, plain, nil), nil
}

// Open mendekripsi blob dari DB kembali jadi TokenSet.
func (c *Cipher) Open(payload []byte) (TokenSet, error) {
	var t TokenSet
	ns := c.aead.NonceSize()
	if len(payload) < ns {
		return t, fmt.Errorf("payload terlalu pendek")
	}
	plain, err := c.aead.Open(nil, payload[:ns], payload[ns:], nil)
	if err != nil {
		return t, fmt.Errorf("decrypt token: %w", err)
	}
	if err := json.Unmarshal(plain, &t); err != nil {
		return t, fmt.Errorf("unmarshal token: %w", err)
	}
	return t, nil
}
