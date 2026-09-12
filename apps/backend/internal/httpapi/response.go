// Package httpapi memuat handler REST + SSE sesuai kontrak doc 06.
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// apiError adalah bentuk error seragam: { "error": { "code", "message" } }.
type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError memetakan sentinel domain ke status + error code (doc 06 §Error Codes).
func writeError(w http.ResponseWriter, log *slog.Logger, err error) {
	status, code := http.StatusInternalServerError, "INTERNAL"
	switch {
	case errors.Is(err, domain.ErrNoRoom):
		status, code = http.StatusConflict, "NO_ROOM"
	case errors.Is(err, domain.ErrNeedsReconnect):
		status, code = http.StatusUnauthorized, "ACCOUNT_NEEDS_RECONNECT"
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrPathExists):
		status, code = http.StatusConflict, "PATH_EXISTS"
	case errors.Is(err, domain.ErrFolderNotEmpty):
		status, code = http.StatusConflict, "FOLDER_NOT_EMPTY"
	case errors.Is(err, domain.ErrRateLimited):
		status, code = http.StatusTooManyRequests, "RATE_LIMITED"
	case errors.Is(err, domain.ErrInvalidArgument):
		status, code = http.StatusBadRequest, "INVALID_ARGUMENT"
	case errors.Is(err, domain.ErrUnsupported):
		status, code = http.StatusNotImplemented, "UNSUPPORTED"
	case errors.Is(err, domain.ErrProvider):
		status, code = http.StatusBadGateway, "PROVIDER_ERROR"
	}
	if status >= 500 {
		log.Error("request gagal", "code", code, "err", err)
	} else {
		log.Warn("request ditolak", "code", code, "err", err)
	}

	var body apiError
	body.Error.Code = code
	body.Error.Message = err.Error()
	writeJSON(w, status, body)
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return errInvalid("body JSON tidak valid: " + err.Error())
	}
	return nil
}

func errInvalid(msg string) error {
	return &wrappedErr{sentinel: domain.ErrInvalidArgument, msg: msg}
}

type wrappedErr struct {
	sentinel error
	msg      string
}

func (e *wrappedErr) Error() string { return e.sentinel.Error() + ": " + e.msg }
func (e *wrappedErr) Unwrap() error { return e.sentinel }

// ---- query helper ----

func queryInt(r *http.Request, key string, def int) int {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func queryInt64(r *http.Request, key string, def int64) int64 {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func queryBool(r *http.Request, key string) bool {
	b, _ := strconv.ParseBool(r.URL.Query().Get(key))
	return b
}

// optionalID mengembalikan pointer ke query param bila diisi, nil bila kosong.
func optionalID(r *http.Request, key string) *string {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return nil
	}
	return &v
}
