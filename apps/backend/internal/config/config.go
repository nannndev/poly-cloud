// Package config memuat seluruh konfigurasi runtime dari environment.
// Satu-satunya tempat backend membaca os.Getenv — lapisan lain menerima Config.
package config

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// OAuthApp = kredensial aplikasi OAuth milik platform untuk satu provider.
// Opsi A (ADR-010): platform yang memegang client_id/secret, bukan user.
type OAuthApp struct {
	ClientID     string
	ClientSecret string
}

func (a OAuthApp) Configured() bool { return a.ClientID != "" && a.ClientSecret != "" }

type Config struct {
	Port     string
	LogLevel string

	DatabaseURL string

	// TokenEncKey = kunci AES-256-GCM, diturunkan SHA-256 dari TOKEN_ENC_KEY
	// supaya panjang secret di env bebas.
	TokenEncKey [32]byte

	// OAuthRedirectURL = halaman frontend yang menerima ?code&state dari provider.
	OAuthRedirectURL string

	// rclone RC daemon (ADR-011): localhost-only + basic auth.
	RcloneRCURL  string
	RcloneRCUser string
	RcloneRCPass string

	// RemoteBaseDir = folder fisik tempat semua objek ditaruh di tiap akun (doc 09 §2).
	RemoteBaseDir string

	RoutingStrategy string

	// v1 single-user: semua request dipetakan ke user ini (doc 06).
	DefaultUserID    string
	DefaultUserEmail string

	CORSOrigins []string

	// OAuthApps di-key provider ('gdrive'|'dropbox'|'onedrive').
	OAuthApps map[string]OAuthApp

	SyncRecurse bool
}

const DefaultUserID = "00000000-0000-0000-0000-000000000001"

func Load() (*Config, error) {
	c := &Config{
		Port:             env("PORT", "8080"),
		LogLevel:         env("LOG_LEVEL", "info"),
		DatabaseURL:      env("DATABASE_URL", ""),
		OAuthRedirectURL: env("OAUTH_REDIRECT_URL", "http://localhost:3000/connect/callback"),
		RcloneRCURL:      strings.TrimSuffix(env("RCLONE_RC_URL", "http://rclone:5572"), "/"),
		RcloneRCUser:     env("RCLONE_RC_USER", ""),
		RcloneRCPass:     env("RCLONE_RC_PASS", ""),
		RemoteBaseDir:    strings.Trim(env("RCLONE_BASE_DIR", "PolyCloud"), "/"),
		RoutingStrategy:  env("ROUTING_STRATEGY", "most-free"),
		DefaultUserID:    env("DEFAULT_USER_ID", DefaultUserID),
		DefaultUserEmail: env("DEFAULT_USER_EMAIL", "owner@polycloud.local"),
		CORSOrigins:      splitList(env("CORS_ORIGINS", "http://localhost:3000")),
		SyncRecurse:      envBool("SYNC_RECURSE", true),
		OAuthApps: map[string]OAuthApp{
			"gdrive":   {ClientID: env("GOOGLE_CLIENT_ID", ""), ClientSecret: env("GOOGLE_CLIENT_SECRET", "")},
			"dropbox":  {ClientID: env("DROPBOX_CLIENT_ID", ""), ClientSecret: env("DROPBOX_CLIENT_SECRET", "")},
			"onedrive": {ClientID: env("ONEDRIVE_CLIENT_ID", ""), ClientSecret: env("ONEDRIVE_CLIENT_SECRET", "")},
		},
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL wajib diisi")
	}
	secret := env("TOKEN_ENC_KEY", "")
	if len(secret) < 16 {
		return nil, fmt.Errorf("TOKEN_ENC_KEY wajib diisi (minimal 16 karakter)")
	}
	c.TokenEncKey = sha256.Sum256([]byte(secret))

	return c, nil
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func splitList(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
