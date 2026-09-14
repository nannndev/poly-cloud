package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/config"
)

func TestUserContext(t *testing.T) {
	ctx := context.Background()
	if u := UserFromContext(ctx); u != "" {
		t.Errorf("expected empty user from blank context, got %q", u)
	}

	ctx = ContextWithUser(ctx, "user-123")
	if u := UserFromContext(ctx); u != "user-123" {
		t.Errorf("expected 'user-123', got %q", u)
	}
}

func TestWithAuthFallbackWhenAuthNotRequired(t *testing.T) {
	defaultUser := "default-user-999"
	var recordedUser string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recordedUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.Config{
		DefaultUserID: defaultUser,
		AuthRequired:  false,
	}

	handler := withAuth(cfg, nil, discardLogger(), next)

	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if recordedUser != defaultUser {
		t.Errorf("expected recorded user %q, got %q", defaultUser, recordedUser)
	}
}

func TestWithAuthRejectsUnauthenticatedWhenAuthRequired(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.Config{
		DefaultUserID: "user-123",
		AuthRequired:  true,
	}

	handler := withAuth(cfg, nil, discardLogger(), next)

	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
	}
}

func TestWithAuthAcceptsValidSessionToken(t *testing.T) {
	secret := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	cfg := &config.Config{
		DefaultUserID: "default-user",
		AuthRequired:  true,
		SessionSecret: secret,
	}

	tokenUser := "session-user-555"
	sessionToken, err := auth.CreateSessionToken(secret[:], tokenUser, "admin@polycloud.local", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create session token: %v", err)
	}

	var recordedUser string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recordedUser = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := withAuth(cfg, nil, discardLogger(), next)

	// Test via Authorization: Bearer <sessionToken>
	req := httptest.NewRequest("GET", "/api/v1/files", nil)
	req.Header.Set("Authorization", "Bearer "+sessionToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if recordedUser != tokenUser {
		t.Errorf("expected %q, got %q", tokenUser, recordedUser)
	}

	// Test via cookie polycloud_session
	recordedUser = ""
	reqCookie := httptest.NewRequest("GET", "/api/v1/files", nil)
	reqCookie.AddCookie(&http.Cookie{Name: "polycloud_session", Value: sessionToken})
	recCookie := httptest.NewRecorder()
	handler.ServeHTTP(recCookie, reqCookie)

	if recCookie.Code != http.StatusOK {
		t.Fatalf("expected 200 OK via cookie, got %d", recCookie.Code)
	}
	if recordedUser != tokenUser {
		t.Errorf("expected %q via cookie, got %q", tokenUser, recordedUser)
	}
}
