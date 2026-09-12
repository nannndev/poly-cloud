package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// RcloneDaemon bicara ke `rclone rcd` lewat RC API (ADR-011).
// Daemon diikat ke jaringan privat + basic auth; backend satu-satunya pemanggil.
type RcloneDaemon struct {
	BaseURL string // mis. http://rclone:5572
	User    string
	Pass    string
	HTTP    *http.Client
}

func NewRcloneDaemon(baseURL, user, pass string) *RcloneDaemon {
	return &RcloneDaemon{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		User:    user,
		Pass:    pass,
		// Tanpa timeout global: upload/download bisa berjam-jam. Batas waktu
		// dikendalikan lewat context per-operasi.
		HTTP: &http.Client{},
	}
}

var _ Engine = (*RcloneDaemon)(nil)

// fsOf menormalkan target jadi bentuk fs rclone. Menerima nama remote polos
// ("acc_x" -> "acc_x:") maupun remote berikut bucket/prefix ("acc_x:bucket"),
// yang sudah berbentuk fs dan dibiarkan apa adanya.
func fsOf(remote string) string {
	if strings.Contains(remote, ":") {
		return remote
	}
	return remote + ":"
}

func (d *RcloneDaemon) newRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, d.BaseURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	if d.User != "" {
		req.SetBasicAuth(d.User, d.Pass)
	}
	return req, nil
}

// rcError membawa pesan error mentah dari daemon untuk klasifikasi di §classify.
type rcError struct {
	Endpoint string
	Status   int
	Message  string
}

func (e *rcError) Error() string {
	return fmt.Sprintf("rclone rc %s: status %d: %s", e.Endpoint, e.Status, e.Message)
}

// call memanggil satu endpoint RC dengan payload JSON dan men-decode hasilnya ke out.
func (d *RcloneDaemon) call(ctx context.Context, endpoint string, in any, out any) error {
	payload, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", endpoint, err)
	}
	req, err := d.newRequest(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("rclone rc %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("rclone rc %s: baca respons: %w", endpoint, err)
	}
	if resp.StatusCode != http.StatusOK {
		return classify(&rcError{Endpoint: endpoint, Status: resp.StatusCode, Message: errMessage(raw)})
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("rclone rc %s: parse respons: %w", endpoint, err)
	}
	return nil
}

// errMessage menarik field "error" dari body error RC; fallback ke body mentah.
func errMessage(raw []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(raw, &e) == nil && e.Error != "" {
		return e.Error
	}
	msg := strings.TrimSpace(string(raw))
	if len(msg) > 500 {
		msg = msg[:500]
	}
	return msg
}

// classify memetakan error rclone ke sentinel domain (doc 03 §8) supaya
// lapisan HTTP bisa memilih error code tanpa tahu soal rclone.
func classify(e *rcError) error {
	low := strings.ToLower(e.Message)
	switch {
	case strings.Contains(low, "directory not found"),
		strings.Contains(low, "object not found"),
		strings.Contains(low, "couldn't find"),
		e.Status == http.StatusNotFound:
		return fmt.Errorf("%w: %s", domain.ErrNotFound, e.Message)
	case strings.Contains(low, "rate limit"),
		strings.Contains(low, "too many requests"),
		strings.Contains(low, "userratelimitexceeded"),
		e.Status == http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", domain.ErrRateLimited, e.Message)
	case strings.Contains(low, "token expired"),
		strings.Contains(low, "invalid_grant"),
		strings.Contains(low, "unauthenticated"),
		strings.Contains(low, "invalid credentials"):
		return fmt.Errorf("%w: %s", domain.ErrNeedsReconnect, e.Message)
	case strings.Contains(low, "doesn't support"),
		strings.Contains(low, "not supported"):
		return fmt.Errorf("%w: %s", domain.ErrUnsupported, e.Message)
	default:
		return fmt.Errorf("%w: %s", domain.ErrProvider, e.Message)
	}
}

// ---- Provisioning ----

func (d *RcloneDaemon) CreateRemote(ctx context.Context, name, rType string, params map[string]any) error {
	return d.call(ctx, "/config/create", map[string]any{
		"name":       name,
		"type":       rType,
		"parameters": params,
		// nonInteractive: jangan pernah blokir menunggu input wizard.
		"opt": map[string]any{"nonInteractive": true, "obscure": false, "noObscure": true},
	}, nil)
}

func (d *RcloneDaemon) UpdateRemote(ctx context.Context, name string, params map[string]any) error {
	return d.call(ctx, "/config/update", map[string]any{
		"name":       name,
		"parameters": params,
		"opt":        map[string]any{"nonInteractive": true, "obscure": false, "noObscure": true},
	}, nil)
}

func (d *RcloneDaemon) DeleteRemote(ctx context.Context, name string) error {
	return d.call(ctx, "/config/delete", map[string]any{"name": name}, nil)
}

func (d *RcloneDaemon) ListRemotes(ctx context.Context) ([]string, error) {
	var out struct {
		Remotes []string `json:"remotes"`
	}
	if err := d.call(ctx, "/config/listremotes", map[string]any{}, &out); err != nil {
		return nil, err
	}
	return out.Remotes, nil
}

func (d *RcloneDaemon) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return d.call(ctx, "/rc/noop", map[string]any{}, nil)
}

// ---- Operasi file ----

// rcEntry = bentuk item pada respons operations/list & operations/stat.
type rcEntry struct {
	Path     string `json:"Path"`
	Name     string `json:"Name"`
	Size     int64  `json:"Size"`
	MimeType string `json:"MimeType"`
	ModTime  string `json:"ModTime"`
	IsDir    bool   `json:"IsDir"`
	ID       string `json:"ID"`
}

func (e rcEntry) toDomain() domain.RemoteEntry {
	var mod time.Time
	if e.ModTime != "" {
		if t, err := time.Parse(time.RFC3339Nano, e.ModTime); err == nil {
			mod = t
		}
	}
	return domain.RemoteEntry{
		Path: e.Path, Name: e.Name, Size: e.Size,
		MimeType: e.MimeType, ModTime: mod, IsDir: e.IsDir, ID: e.ID,
	}
}

func (d *RcloneDaemon) List(ctx context.Context, remote, p string, recurse bool) ([]domain.RemoteEntry, error) {
	var out struct {
		List []rcEntry `json:"list"`
	}
	err := d.call(ctx, "/operations/list", map[string]any{
		"fs":     fsOf(remote),
		"remote": strings.Trim(p, "/"),
		"opt":    map[string]any{"recurse": recurse, "showOrigIDs": true},
	}, &out)
	if err != nil {
		return nil, err
	}
	entries := make([]domain.RemoteEntry, 0, len(out.List))
	for _, e := range out.List {
		entries = append(entries, e.toDomain())
	}
	return entries, nil
}

func (d *RcloneDaemon) Stat(ctx context.Context, remote, p string) (*domain.RemoteEntry, error) {
	var out struct {
		Item *rcEntry `json:"item"`
	}
	err := d.call(ctx, "/operations/stat", map[string]any{
		"fs":     fsOf(remote),
		"remote": strings.Trim(p, "/"),
		"opt":    map[string]any{"showOrigIDs": true},
	}, &out)
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, fmt.Errorf("%w: %s/%s", domain.ErrNotFound, remote, p)
	}
	e := out.Item.toDomain()
	return &e, nil
}

func (d *RcloneDaemon) About(ctx context.Context, remote string) (domain.Quota, error) {
	var out struct {
		Total *int64 `json:"total"`
		Used  *int64 `json:"used"`
		Free  *int64 `json:"free"`
	}
	if err := d.call(ctx, "/operations/about", map[string]any{"fs": fsOf(remote)}, &out); err != nil {
		return domain.Quota{}, err
	}
	q := domain.Quota{}
	if out.Total != nil {
		q.Total = *out.Total
	}
	if out.Used != nil {
		q.Used = *out.Used
	}
	// Provider tanpa `free` eksplisit (mis. S3): turunkan dari total-used bila bisa.
	switch {
	case out.Free != nil:
		q.Free = *out.Free
	case q.Total > 0:
		q.Free = q.Total - q.Used
	}
	return q, nil
}

func (d *RcloneDaemon) Mkdir(ctx context.Context, remote, p string) error {
	return d.call(ctx, "/operations/mkdir", map[string]any{
		"fs": fsOf(remote), "remote": strings.Trim(p, "/"),
	}, nil)
}

// UploadStream mengalirkan reader ke remote:destDir/name tanpa menyentuh disk server
// (doc 03 §7). Body multipart dibangun on-the-fly lewat io.Pipe sehingga file
// tak pernah dibuffer utuh di memori.
func (d *RcloneDaemon) UploadStream(ctx context.Context, r io.Reader, remote, destDir, name string) error {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		part, err := mw.CreateFormFile("file", name)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, r); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if err := mw.Close(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = pw.Close()
	}()

	q := url.Values{}
	q.Set("fs", fsOf(remote))
	q.Set("remote", strings.Trim(destDir, "/"))
	endpoint := "/operations/uploadfile?" + q.Encode()

	req, err := d.newRequest(ctx, http.MethodPost, endpoint, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := d.HTTP.Do(req)
	if err != nil {
		_ = pr.CloseWithError(err)
		return fmt.Errorf("rclone rc uploadfile: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return classify(&rcError{Endpoint: "/operations/uploadfile", Status: resp.StatusCode, Message: errMessage(raw)})
	}
	return nil
}

// Download menstream objek dari remote ke writer lewat HTTP serve daemon
// (`--rc-serve`). Tak ada file perantara di server.
func (d *RcloneDaemon) Download(ctx context.Context, remote, src string, w io.Writer) error {
	// Bentuk URL yang dimengerti serve: /[fs]/path/to/object — nama fs WAJIB
	// dikurung siku, kalau tidak daemon membaca segmen pertama sebagai nama
	// direktori biasa dan menjawab 404. Tiap segmen di-escape; daemon menerima
	// bentuk ter-escape maupun mentah.
	target := "/" + url.PathEscape("["+fsOf(remote)+"]")
	for _, seg := range strings.Split(strings.Trim(src, "/"), "/") {
		if seg == "" {
			continue
		}
		target += "/" + url.PathEscape(seg)
	}

	req, err := d.newRequest(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	resp, err := d.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("rclone serve download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return classify(&rcError{Endpoint: target, Status: resp.StatusCode, Message: errMessage(raw)})
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("stream download %s: %w", target, err)
	}
	return nil
}

// Move memindah objek antar remote (transfer data nyata — beda dari pindah folder VFS).
func (d *RcloneDaemon) Move(ctx context.Context, srcRemote, src, dstRemote, dst string) error {
	return d.call(ctx, "/operations/movefile", map[string]any{
		"srcFs": fsOf(srcRemote), "srcRemote": strings.Trim(src, "/"),
		"dstFs": fsOf(dstRemote), "dstRemote": strings.Trim(dst, "/"),
	}, nil)
}

func (d *RcloneDaemon) Delete(ctx context.Context, remote, p string) error {
	err := d.call(ctx, "/operations/deletefile", map[string]any{
		"fs": fsOf(remote), "remote": strings.Trim(p, "/"),
	}, nil)
	// Objek sudah tak ada = hasil akhir yang diinginkan (idempotent, doc 03 §8).
	if err != nil && errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	return err
}

// Join menyusun path objek di dalam remote.
func Join(parts ...string) string {
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.Trim(p, "/"); p != "" {
			cleaned = append(cleaned, p)
		}
	}
	return path.Join(cleaned...)
}
