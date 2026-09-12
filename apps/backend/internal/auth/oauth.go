package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/config"
	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// ProviderSpec mendeskripsikan satu provider OAuth + pemetaannya ke tipe rclone
// (doc 10 §4). Menambah provider = menambah satu entri di sini.
type ProviderSpec struct {
	Provider   string
	RcloneType string
	AuthURL    string
	TokenURL   string
	Scopes     []string
	// ExtraAuth = query tambahan pada auth URL (mis. minta refresh_token dari Google).
	ExtraAuth map[string]string
	// RcloneExtra = parameter tambahan saat config/create.
	RcloneExtra map[string]any
}

// OAuthProviders = provider yang memakai redirect consent.
var OAuthProviders = map[string]ProviderSpec{
	"gdrive": {
		Provider:   "gdrive",
		RcloneType: "drive",
		AuthURL:    "https://accounts.google.com/o/oauth2/auth",
		TokenURL:   "https://oauth2.googleapis.com/token",
		Scopes:     []string{"https://www.googleapis.com/auth/drive"},
		// access_type=offline + prompt=consent wajib agar refresh_token dikirim.
		ExtraAuth:   map[string]string{"access_type": "offline", "prompt": "consent"},
		RcloneExtra: map[string]any{"scope": "drive"},
	},
	"dropbox": {
		Provider:    "dropbox",
		RcloneType:  "dropbox",
		AuthURL:     "https://www.dropbox.com/oauth2/authorize",
		TokenURL:    "https://api.dropboxapi.com/oauth2/token",
		ExtraAuth:   map[string]string{"token_access_type": "offline"},
		RcloneExtra: map[string]any{},
	},
	"onedrive": {
		Provider:    "onedrive",
		RcloneType:  "onedrive",
		AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		Scopes:      []string{"Files.ReadWrite.All", "offline_access"},
		RcloneExtra: map[string]any{"drive_type": "personal"},
	},
}

// KeyProviderSpec mendeskripsikan provider berbasis kredensial statis
// (tanpa redirect consent, doc 10 §4).
type KeyProviderSpec struct {
	RcloneType string
	// Fields = parameter yang diteruskan ke rclone config/create.
	Fields []string
	// Required = field yang wajib diisi user.
	Required []string
	// NeedsBucket menandai provider penyimpanan objek: akar remote-nya adalah
	// daftar bucket, jadi objek harus ditaruh di dalam sebuah bucket. Nilai
	// "bucket" dari form disimpan sebagai accounts.root_path, bukan dikirim
	// ke rclone sebagai parameter remote.
	NeedsBucket bool
}

var KeyProviders = map[string]KeyProviderSpec{
	"s3": {
		RcloneType:  "s3",
		Fields:      []string{"provider", "access_key_id", "secret_access_key", "region", "endpoint", "location_constraint", "acl"},
		Required:    []string{"access_key_id", "secret_access_key"},
		NeedsBucket: true,
	},
	"b2": {
		RcloneType:  "b2",
		Fields:      []string{"account", "key", "hard_delete"},
		Required:    []string{"account", "key"},
		NeedsBucket: true,
	},
	"r2": {
		RcloneType:  "s3",
		Fields:      []string{"access_key_id", "secret_access_key", "endpoint"},
		Required:    []string{"access_key_id", "secret_access_key", "endpoint"},
		NeedsBucket: true,
	},
}

// FormFields mengembalikan daftar field yang diminta dari user di UI,
// termasuk "bucket" untuk provider penyimpanan objek.
func (s KeyProviderSpec) FormFields() []string {
	if !s.NeedsBucket {
		return s.Fields
	}
	return append(append([]string{}, s.Fields...), "bucket")
}

// RequiredFields mengembalikan field wajib yang ditampilkan di UI.
func (s KeyProviderSpec) RequiredFields() []string {
	if !s.NeedsBucket {
		return s.Required
	}
	return append(append([]string{}, s.Required...), "bucket")
}

// pendingState menyimpan state OAuth antara /connect dan /callback.
type pendingState struct {
	Provider  string
	Label     string
	UserID    string
	ExpiresAt time.Time
}

// Manager membangun auth URL, menukar code jadi token, dan me-refresh token.
type Manager struct {
	cfg    *config.Config
	http   *http.Client
	mu     sync.Mutex
	states map[string]pendingState
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg:    cfg,
		http:   &http.Client{Timeout: 30 * time.Second},
		states: make(map[string]pendingState),
	}
}

// AuthURL memulai flow: bangun consent URL + simpan state acak yang terikat user.
func (m *Manager) AuthURL(provider, label, userID string) (string, error) {
	spec, ok := OAuthProviders[provider]
	if !ok {
		return "", fmt.Errorf("%w: provider %q bukan provider OAuth", domain.ErrInvalidArgument, provider)
	}
	app := m.cfg.OAuthApps[provider]
	if !app.Configured() {
		return "", fmt.Errorf("%w: kredensial OAuth %s belum diset (isi env client id & secret)",
			domain.ErrInvalidArgument, provider)
	}

	state, err := randomState()
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	m.pruneLocked()
	m.states[state] = pendingState{
		Provider: provider, Label: label, UserID: userID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	m.mu.Unlock()

	q := url.Values{}
	q.Set("client_id", app.ClientID)
	q.Set("redirect_uri", m.cfg.OAuthRedirectURL)
	q.Set("response_type", "code")
	q.Set("state", state)
	if len(spec.Scopes) > 0 {
		q.Set("scope", strings.Join(spec.Scopes, " "))
	}
	for k, v := range spec.ExtraAuth {
		q.Set(k, v)
	}
	return spec.AuthURL + "?" + q.Encode(), nil
}

// ConsumeState memvalidasi state dari callback dan memakainya sekali saja.
func (m *Manager) ConsumeState(state string) (pendingState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pruneLocked()

	st, ok := m.states[state]
	if !ok {
		return pendingState{}, fmt.Errorf("%w: state tidak dikenal atau kedaluwarsa", domain.ErrInvalidArgument)
	}
	delete(m.states, state)
	return st, nil
}

func (m *Manager) pruneLocked() {
	now := time.Now()
	for k, v := range m.states {
		if now.After(v.ExpiresAt) {
			delete(m.states, k)
		}
	}
}

// Provider mengembalikan provider & label yang tersimpan di state.
func (s pendingState) Provider_() string { return s.Provider }
func (s pendingState) Label_() string    { return s.Label }
func (s pendingState) UserID_() string   { return s.UserID }

// tokenResponse = bentuk umum respons token endpoint OAuth2.
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (m *Manager) postToken(ctx context.Context, spec ProviderSpec, form url.Values) (TokenSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, spec.TokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := m.http.Do(req)
	if err != nil {
		return TokenSet{}, fmt.Errorf("hubungi token endpoint %s: %w", spec.Provider, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return TokenSet{}, fmt.Errorf("baca respons token: %w", err)
	}

	var tr tokenResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return TokenSet{}, fmt.Errorf("parse respons token (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || tr.Error != "" {
		msg := tr.ErrorDescription
		if msg == "" {
			msg = tr.Error
		}
		return TokenSet{}, fmt.Errorf("%w: token endpoint %s menolak: %s",
			domain.ErrProvider, spec.Provider, msg)
	}

	ts := TokenSet{
		AccessToken:  tr.AccessToken,
		TokenType:    tr.TokenType,
		RefreshToken: tr.RefreshToken,
	}
	if ts.TokenType == "" {
		ts.TokenType = "Bearer"
	}
	if tr.ExpiresIn > 0 {
		ts.Expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	return ts, nil
}

// Exchange menukar authorization code jadi access + refresh token.
func (m *Manager) Exchange(ctx context.Context, provider, code string) (TokenSet, error) {
	spec, ok := OAuthProviders[provider]
	if !ok {
		return TokenSet{}, fmt.Errorf("%w: provider %q", domain.ErrInvalidArgument, provider)
	}
	app := m.cfg.OAuthApps[provider]
	if !app.Configured() {
		return TokenSet{}, fmt.Errorf("%w: kredensial OAuth %s belum diset", domain.ErrInvalidArgument, provider)
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", m.cfg.OAuthRedirectURL)
	form.Set("client_id", app.ClientID)
	form.Set("client_secret", app.ClientSecret)
	return m.postToken(ctx, spec, form)
}

// Refresh menukar refresh_token jadi access token baru. Backend yang berwenang
// refresh karena token milik platform (ADR-010).
func (m *Manager) Refresh(ctx context.Context, provider string, old TokenSet) (TokenSet, error) {
	spec, ok := OAuthProviders[provider]
	if !ok {
		return TokenSet{}, fmt.Errorf("%w: provider %q", domain.ErrInvalidArgument, provider)
	}
	if old.RefreshToken == "" {
		return TokenSet{}, fmt.Errorf("%w: tak ada refresh_token", domain.ErrNeedsReconnect)
	}
	app := m.cfg.OAuthApps[provider]
	if !app.Configured() {
		return TokenSet{}, fmt.Errorf("%w: kredensial OAuth %s belum diset", domain.ErrInvalidArgument, provider)
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", old.RefreshToken)
	form.Set("client_id", app.ClientID)
	form.Set("client_secret", app.ClientSecret)

	ts, err := m.postToken(ctx, spec, form)
	if err != nil {
		return TokenSet{}, err
	}
	// Sebagian provider tak mengirim ulang refresh_token → pertahankan yang lama.
	if ts.RefreshToken == "" {
		ts.RefreshToken = old.RefreshToken
	}
	return ts, nil
}

// RcloneParams menyusun parameter config/create untuk provider OAuth:
// token JSON + client app platform (doc 10 §3 langkah 8).
func RcloneParams(provider string, ts TokenSet, app config.OAuthApp) (rcloneType string, params map[string]any, err error) {
	spec, ok := OAuthProviders[provider]
	if !ok {
		return "", nil, fmt.Errorf("%w: provider %q", domain.ErrInvalidArgument, provider)
	}
	tokenJSON, err := json.Marshal(ts)
	if err != nil {
		return "", nil, fmt.Errorf("marshal token: %w", err)
	}
	params = map[string]any{
		"token":         string(tokenJSON),
		"client_id":     app.ClientID,
		"client_secret": app.ClientSecret,
	}
	for k, v := range spec.RcloneExtra {
		params[k] = v
	}
	return spec.RcloneType, params, nil
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("acak state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
