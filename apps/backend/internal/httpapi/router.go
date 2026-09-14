package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/polycloud/platform/apps/backend/internal/config"
	"github.com/polycloud/platform/apps/backend/internal/index"
)

// Health = pemeriksa kesiapan dependensi (DB, rclone daemon).
type Health struct {
	DB     func() error
	Engine func() error
}

// NewRouter merakit seluruh rute sesuai doc 06.
func NewRouter(cfg *config.Config, api *API, hub *Hub, keys *index.KeyRepo, health Health, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		status := map[string]string{"status": "ok", "db": "ok", "engine": "ok"}
		code := http.StatusOK
		if err := health.DB(); err != nil {
			status["status"], status["db"] = "degraded", err.Error()
			code = http.StatusServiceUnavailable
		}
		if err := health.Engine(); err != nil {
			// Engine mati hanya melumpuhkan operasi file, bukan browse index.
			status["status"], status["engine"] = "degraded", err.Error()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"service": "polycloud-api", "version": "v1"})
	})

	// Accounts
	mux.HandleFunc("GET /api/v1/accounts", api.ListAccounts)
	mux.HandleFunc("POST /api/v1/accounts/connect", api.ConnectAccount)
	mux.HandleFunc("POST /api/v1/accounts/callback", api.CallbackAccount)
	mux.HandleFunc("POST /api/v1/accounts/{id}/sync", api.SyncAccount)
	mux.HandleFunc("DELETE /api/v1/accounts/{id}", api.DeleteAccount)
	mux.HandleFunc("GET /api/v1/providers", api.Providers)
	mux.HandleFunc("GET /api/v1/settings", api.Settings)

	// Files
	mux.HandleFunc("GET /api/v1/files", api.ListFiles)
	mux.HandleFunc("GET /api/v1/files/search", api.SearchFiles)
	mux.HandleFunc("POST /api/v1/files/upload", api.UploadFile)
	mux.HandleFunc("GET /api/v1/files/{id}/download", api.DownloadFile)
	mux.HandleFunc("POST /api/v1/files/{id}/move", api.MoveFile)
	mux.HandleFunc("PATCH /api/v1/files/{id}", api.PatchFile)
	mux.HandleFunc("DELETE /api/v1/files/{id}", api.DeleteFile)

	// Folders (VFS)
	mux.HandleFunc("GET /api/v1/folders", api.ListFolders)
	mux.HandleFunc("POST /api/v1/folders", api.CreateFolder)
	mux.HandleFunc("PATCH /api/v1/folders/{id}", api.PatchFolder)
	mux.HandleFunc("DELETE /api/v1/folders/{id}", api.DeleteFolder)

	// Quota & events
	mux.HandleFunc("GET /api/v1/quota", api.Quota)
	mux.HandleFunc("GET /api/v1/events/uploads/{jobId}", hub.HandleUploadEvents)

	// Model Context Protocol (MCP) endpoint
	mux.HandleFunc("POST /api/v1/mcp", api.HandleMCP)

	// Developer API Keys
	mux.HandleFunc("GET /api/v1/api-keys", api.ListAPIKeys)
	mux.HandleFunc("POST /api/v1/api-keys", api.CreateAPIKey)
	mux.HandleFunc("DELETE /api/v1/api-keys/{id}", api.DeleteAPIKey)

	// Authentication
	mux.HandleFunc("POST /api/v1/auth/login", api.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", api.Logout)
	mux.HandleFunc("GET /api/v1/auth/me", api.Me)

	return withRecover(log, withLogging(log, withCORS(cfg.CORSOrigins, withAuth(cfg, keys, log, mux))))
}
