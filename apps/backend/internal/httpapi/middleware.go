package httpapi

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-File-Name")
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
