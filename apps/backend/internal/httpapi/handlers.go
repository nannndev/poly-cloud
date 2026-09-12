package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/config"
	"github.com/polycloud/platform/apps/backend/internal/domain"
	"github.com/polycloud/platform/apps/backend/internal/index"
	"github.com/polycloud/platform/apps/backend/internal/storage"
)

// API menyatukan dependensi handler.
type API struct {
	cfg      *config.Config
	files    storage.Service
	accounts *storage.AccountService
	folders  *index.FolderRepo
	hub      *Hub
	log      *slog.Logger
}

func NewAPI(cfg *config.Config, files storage.Service, accounts *storage.AccountService, folders *index.FolderRepo, hub *Hub, log *slog.Logger) *API {
	return &API{cfg: cfg, files: files, accounts: accounts, folders: folders, hub: hub, log: log}
}

// userID mengembalikan pemilik request. v1 single-user: selalu user default
// (doc 06). Saat multi-user aktif, di sinilah sesi dibaca.
func (a *API) userID(_ *http.Request) string { return a.cfg.DefaultUserID }

// ---- Accounts ----

func (a *API) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := a.accounts.List(r.Context(), a.userID(r))
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

type connectRequest struct {
	Provider string            `json:"provider"`
	Label    string            `json:"label"`
	Fields   map[string]string `json:"fields"`
}

// ConnectAccount memulai OAuth (balas auth_url) atau langsung membuat account
// untuk provider berbasis key (S3/B2/R2).
func (a *API) ConnectAccount(w http.ResponseWriter, r *http.Request) {
	var req connectRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, a.log, err)
		return
	}
	req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
	if req.Provider == "" {
		writeError(w, a.log, errInvalid("field provider wajib diisi"))
		return
	}

	if _, isKeyProvider := auth.KeyProviders[req.Provider]; isKeyProvider {
		account, err := a.accounts.ConnectWithKeys(r.Context(), a.userID(r), req.Provider, req.Label, req.Fields)
		if err != nil {
			writeError(w, a.log, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"account_id": account.ID, "status": account.Status, "account": account,
		})
		return
	}

	authURL, err := a.accounts.StartConnect(r.Context(), a.userID(r), req.Provider, req.Label)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"auth_url": authURL})
}

type callbackRequest struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	State    string `json:"state"`
}

// CallbackAccount menyelesaikan OAuth: tukar code, simpan token, buat remote,
// lalu jalankan initial sync di latar belakang.
func (a *API) CallbackAccount(w http.ResponseWriter, r *http.Request) {
	var req callbackRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, a.log, err)
		return
	}
	if req.Code == "" || req.State == "" {
		writeError(w, a.log, errInvalid("field code dan state wajib diisi"))
		return
	}

	account, err := a.accounts.CompleteConnect(r.Context(), req.Code, req.State)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	a.syncInBackground(account.UserID, account.ID)

	writeJSON(w, http.StatusCreated, map[string]any{
		"account_id": account.ID, "status": account.Status, "account": account,
	})
}

func (a *API) SyncAccount(w http.ResponseWriter, r *http.Request) {
	res, err := a.accounts.Sync(r.Context(), a.userID(r), r.PathValue("id"))
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (a *API) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if err := a.accounts.Disconnect(r.Context(), a.userID(r), r.PathValue("id")); err != nil {
		writeError(w, a.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Providers melaporkan provider mana yang siap dipakai — UI memakainya untuk
// menandai provider yang kredensial OAuth-nya belum diisi.
func (a *API) Providers(w http.ResponseWriter, _ *http.Request) {
	type providerInfo struct {
		Provider   string   `json:"provider"`
		Kind       string   `json:"kind"` // "oauth" | "keys"
		Configured bool     `json:"configured"`
		Fields     []string `json:"fields,omitempty"`
		Required   []string `json:"required,omitempty"`
	}
	out := make([]providerInfo, 0, len(auth.OAuthProviders)+len(auth.KeyProviders))
	for name := range auth.OAuthProviders {
		out = append(out, providerInfo{
			Provider: name, Kind: "oauth", Configured: a.cfg.OAuthApps[name].Configured(),
		})
	}
	for name, spec := range auth.KeyProviders {
		out = append(out, providerInfo{
			Provider: name, Kind: "keys", Configured: true,
			Fields: spec.FormFields(), Required: spec.RequiredFields(),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- Files ----

// parseQuery membaca parameter list/search yang dipakai bersama.
func parseQuery(r *http.Request) domain.SearchQuery {
	q := domain.SearchQuery{
		Q:         strings.TrimSpace(r.URL.Query().Get("q")),
		Mime:      strings.TrimSpace(r.URL.Query().Get("type")),
		AccountID: strings.TrimSpace(r.URL.Query().Get("account_id")),
		MinSize:   queryInt64(r, "min_size", 0),
		MaxSize:   queryInt64(r, "max_size", 0),
		Page:      queryInt(r, "page", 1),
		PerPage:   queryInt(r, "per_page", 100),
		Sort:      strings.TrimSpace(r.URL.Query().Get("sort")),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 500 {
		q.PerPage = 100
	}
	return q
}

// resolveFolder menerjemahkan ?folder_id= atau ?path= jadi folder id (nil = root).
func (a *API) resolveFolder(r *http.Request) (*string, error) {
	if id := optionalID(r, "folder_id"); id != nil {
		return id, nil
	}
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" || path == "/" {
		return nil, nil
	}
	folder, err := a.folders.GetByPath(r.Context(), a.userID(r), path)
	if err != nil {
		return nil, err
	}
	return &folder.ID, nil
}

func (a *API) ListFiles(w http.ResponseWriter, r *http.Request) {
	folderID, err := a.resolveFolder(r)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	res, err := a.files.List(r.Context(), a.userID(r), folderID, parseQuery(r))
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (a *API) SearchFiles(w http.ResponseWriter, r *http.Request) {
	res, err := a.files.Search(r.Context(), a.userID(r), parseQuery(r))
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// UploadFile menerima stream (multipart atau body mentah) dan meneruskannya
// ke provider tanpa menyentuh disk (doc 03 §7). Progres disiarkan lewat SSE
// pada job id yang dikembalikan di header X-Job-Id.
func (a *API) UploadFile(w http.ResponseWriter, r *http.Request) {
	folderID, err := a.resolveFolder(r)
	if err != nil {
		writeError(w, a.log, err)
		return
	}

	jobID := strings.TrimSpace(r.URL.Query().Get("job_id"))
	if jobID == "" {
		jobID = newJobID()
	}
	w.Header().Set("X-Job-Id", jobID)

	var (
		body io.ReadCloser = r.Body
		name string
		size = r.ContentLength
	)
	defer r.Body.Close()

	if ct := r.Header.Get("Content-Type"); strings.HasPrefix(ct, "multipart/") {
		// Ambil part pertama saja — stream, bukan ParseMultipartForm (yang menulis ke disk).
		mr, err := r.MultipartReader()
		if err != nil {
			writeError(w, a.log, errInvalid("multipart tidak valid: "+err.Error()))
			return
		}
		part, err := mr.NextPart()
		if err != nil {
			writeError(w, a.log, errInvalid("tak ada file di body multipart"))
			return
		}
		defer part.Close()
		body = part
		name = part.FileName()
		// Ukuran asli tak diketahui dari part; pakai hint dari query bila ada.
		size = queryInt64(r, "size", 0)
	}

	if n := strings.TrimSpace(r.Header.Get("X-File-Name")); n != "" {
		name = n
	}
	if n := strings.TrimSpace(r.URL.Query().Get("name")); n != "" {
		name = n
	}
	if name == "" {
		writeError(w, a.log, errInvalid("nama file wajib (query ?name= atau header X-File-Name)"))
		return
	}
	name = filepath.Base(name)

	onProgress := func(sent, total int64) {
		a.hub.Publish(jobID, "progress", map[string]any{
			"job": jobID, "bytes": sent, "total": total,
		})
	}

	res, err := a.files.Upload(r.Context(), a.userID(r), folderID, name, body, size, onProgress)
	if err != nil {
		a.hub.Publish(jobID, "error", map[string]any{"job": jobID, "message": err.Error()})
		writeError(w, a.log, err)
		return
	}

	a.hub.Publish(jobID, "done", map[string]any{
		"job": jobID, "file_id": res.File.ID, "account_label": res.AccountLabel,
	})
	w.Header().Set("X-Account-Label", res.AccountLabel)
	w.Header().Set("X-Routed-By", res.RoutedBy)
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": res.File.ID, "name": res.File.Name,
		"account_id": res.AccountID, "account_label": res.AccountLabel,
		"routed_by": res.RoutedBy, "job_id": jobID, "file": res.File,
	})
}

func (a *API) DownloadFile(w http.ResponseWriter, r *http.Request) {
	dl, err := a.files.Download(r.Context(), a.userID(r), r.PathValue("id"))
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	defer dl.Body.Close()

	// `?inline=1` menyajikan file untuk ditampilkan di halaman (preview PDF,
	// gambar, video) alih-alih memicu unduhan. Tanpa itu, iframe PDF akan
	// men-download berkas, bukan merendernya.
	disposition := "attachment"
	if queryBool(r, "inline") {
		disposition = "inline"
	}

	w.Header().Set("Content-Type", dl.Mime)
	w.Header().Set("Content-Disposition",
		mime.FormatMediaType(disposition, map[string]string{"filename": dl.Name}))
	if dl.SizeBytes > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", dl.SizeBytes))
	}

	// Header sudah terkirim di titik ini, jadi error di tengah stream hanya
	// bisa dilaporkan lewat log + koneksi yang terputus.
	if _, err := io.Copy(w, dl.Body); err != nil {
		a.log.Warn("stream download terputus", "file", r.PathValue("id"), "err", err)
	}
}

type moveRequest struct {
	DestAccountID string `json:"dest_account_id"`
}

func (a *API) MoveFile(w http.ResponseWriter, r *http.Request) {
	var req moveRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, a.log, err)
		return
	}
	if req.DestAccountID == "" {
		writeError(w, a.log, errInvalid("field dest_account_id wajib diisi"))
		return
	}
	if err := a.files.MoveToAccount(r.Context(), a.userID(r), r.PathValue("id"), req.DestAccountID); err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"moved": true})
}

// patchFileRequest memakai pointer supaya "field tak dikirim" bisa dibedakan
// dari "field dikirim bernilai null" (pindah ke root).
type patchFileRequest struct {
	Name     *string  `json:"name"`
	FolderID **string `json:"folder_id"`
}

func (a *API) PatchFile(w http.ResponseWriter, r *http.Request) {
	var raw map[string]any
	if err := decodeJSON(r, &raw); err != nil {
		writeError(w, a.log, err)
		return
	}
	name, folderID, err := parseOrganizationPatch(raw)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	entry, err := a.files.UpdateOrganization(r.Context(), a.userID(r), r.PathValue("id"), name, folderID)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (a *API) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if err := a.files.Delete(r.Context(), a.userID(r), r.PathValue("id")); err != nil {
		writeError(w, a.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// parseOrganizationPatch membaca body PATCH yang field-nya semua opsional.
func parseOrganizationPatch(raw map[string]any) (*string, **string, error) {
	var (
		name     *string
		folderID **string
	)
	for k, v := range raw {
		switch k {
		case "name":
			s, ok := v.(string)
			if !ok || strings.TrimSpace(s) == "" {
				return nil, nil, errInvalid("field name harus string tak kosong")
			}
			trimmed := strings.TrimSpace(s)
			name = &trimmed
		case "folder_id", "parent_id":
			if v == nil {
				var null *string
				folderID = &null
				continue
			}
			s, ok := v.(string)
			if !ok || strings.TrimSpace(s) == "" {
				return nil, nil, errInvalid("field " + k + " harus string UUID atau null")
			}
			trimmed := strings.TrimSpace(s)
			p := &trimmed
			folderID = &p
		default:
			return nil, nil, errInvalid("field tak dikenal: " + k)
		}
	}
	if name == nil && folderID == nil {
		return nil, nil, errInvalid("tak ada field yang diubah")
	}
	return name, folderID, nil
}

// Settings melaporkan konfigurasi runtime backend (read-only). Nilai-nilai ini
// berasal dari environment saat proses start — mengubahnya butuh restart, jadi
// UI menampilkannya, bukan menyuntingnya.
func (a *API) Settings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"routing_strategy": a.cfg.RoutingStrategy,
		"remote_base_dir":  a.cfg.RemoteBaseDir,
		// Model A (whole-file). Model B (chunked) belum diimplementasikan (ADR-001).
		"storage_model": "A",
		"chunking":      false,
		"sync_recurse":  a.cfg.SyncRecurse,
		"multi_user":    false,
	})
}

// ---- Quota ----

func (a *API) Quota(w http.ResponseWriter, r *http.Request) {
	report, err := a.files.Quota(r.Context(), a.userID(r))
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// syncInBackground menjalankan initial sync tanpa menahan respons HTTP.
func (a *API) syncInBackground(userID, accountID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if _, err := a.accounts.Sync(ctx, userID, accountID); err != nil {
			a.log.Warn("initial sync gagal", "account", accountID, "err", err)
		}
	}()
}

// newJobID membuat id acak untuk satu job upload (dipakai sebagai kanal SSE).
func newJobID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand praktis tak pernah gagal; fallback berbasis waktu tetap unik
		// untuk keperluan kanal SSE berumur pendek.
		return "job-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
}
