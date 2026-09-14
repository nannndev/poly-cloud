package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/config"
	"github.com/polycloud/platform/apps/backend/internal/domain"
	"github.com/polycloud/platform/apps/backend/internal/index"
)

type ctxKey int

const userCtxKey ctxKey = 1

func ContextWithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userCtxKey, userID)
}

func UserFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userCtxKey).(string); ok && v != "" {
		return v
	}
	return ""
}

// withCORS mengizinkan origin frontend yang terdaftar di config.
func withCORS(origins []string, next http.Handler) http.Handler {
	allowAll := slices.Contains(origins, "*")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (allowAll || slices.Contains(origins, origin)) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-File-Name, X-API-Key")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition, X-Account-Label, X-Routed-By")
			w.Header().Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withAuth mengekstrak dan memvalidasi API Key atau Session Token.
// Mendukung Authorization: Bearer, X-API-Key, dan cookie polycloud_session.
func withAuth(cfg *config.Config, keys *index.KeyRepo, log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions ||
			strings.HasPrefix(r.URL.Path, "/healthz") ||
			r.URL.Path == "/api/v1" ||
			r.URL.Path == "/api/v1/auth/login" ||
			r.URL.Path == "/api/v1/accounts/callback" {
			next.ServeHTTP(w, r)
			return
		}

		token := ""
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			}
		}
		if token == "" {
			token = strings.TrimSpace(r.Header.Get("X-API-Key"))
		}
		if token == "" {
			if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
				token = strings.TrimSpace(cookie.Value)
			}
		}

		if token != "" {
			// 1. Developer API Key (plc_live_...)
			if strings.HasPrefix(token, auth.KeyPrefix) {
				if keys != nil {
					hash := auth.HashKey(token)
					key, err := keys.FindByHash(r.Context(), hash)
					if err != nil {
						writeError(w, log, domain.ErrUnauthorized)
						return
					}
					ctx := ContextWithUser(r.Context(), key.UserID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			} else {
				// 2. Session token HMAC-SHA256
				claims, err := auth.VerifySessionToken(cfg.SessionSecret[:], token)
				if err != nil {
					writeError(w, log, domain.ErrUnauthorized)
					return
				}
				ctx := ContextWithUser(r.Context(), claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Jika token tidak diberikan sama sekali:
		if cfg != nil && cfg.AuthRequired {
			writeError(w, log, domain.ErrUnauthorized)
			return
		}

		defaultUserID := config.DefaultUserID
		if cfg != nil && cfg.DefaultUserID != "" {
			defaultUserID = cfg.DefaultUserID
		}
		ctx := ContextWithUser(r.Context(), defaultUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// statusRecorder menangkap status code untuk log akses.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Flush meneruskan flush ke writer asli — dibutuhkan SSE.
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func withLogging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Jangan bising untuk health check.
		if strings.HasPrefix(r.URL.Path, "/healthz") {
			return
		}
		log.Info("http",
			"method", r.Method, "path", r.URL.Path,
			"status", rec.status, "dur", time.Since(start).String())
	})
}

func withRecover(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic di handler", "path", r.URL.Path, "panic", rec)
				var body apiError
				body.Error.Code = "INTERNAL"
				body.Error.Message = "terjadi kesalahan internal"
				writeJSON(w, http.StatusInternalServerError, body)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
