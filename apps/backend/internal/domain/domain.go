// Package domain memuat tipe inti yang dibagi seluruh lapisan backend.
// Tidak bergantung pada paket internal lain — menghindari import cycle.
package domain

import (
	"errors"
	"strings"
	"time"
)

// ---- Account ----

const (
	StatusActive         = "active"
	StatusNeedsReconnect = "needs_reconnect"
	StatusError          = "error"
	StatusSyncing        = "syncing"
)

// Account = 1 akun user pada 1 provider. Kuota di-cache di sini (doc 05 §3).
type Account struct {
	ID           string     `json:"id"`
	UserID       string     `json:"-"`
	Provider     string     `json:"provider"`
	Label        string     `json:"label"`
	RcloneRemote string     `json:"-"`
	RootPath     string     `json:"-"`
	TotalBytes   int64      `json:"total_bytes"`
	UsedBytes    int64      `json:"used_bytes"`
	FreeBytes    int64      `json:"free_bytes"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	LastSynced   *time.Time `json:"last_synced"`
}

// FsTarget mengembalikan alamat fs rclone untuk account ini: nama remote,
// ditambah bucket/prefix bila provider menyimpan objek di dalam bucket
// (mis. "acc_x:bucket-saya"). Inilah yang dioper ke Engine, bukan nama remote
// mentah — akar remote S3 adalah daftar bucket, bukan tempat menaruh objek.
func (a Account) FsTarget() string {
	if a.RootPath == "" {
		return a.RcloneRemote
	}
	return a.RcloneRemote + ":" + strings.Trim(a.RootPath, "/")
}

// Quota = hasil rclone `about` untuk satu remote.
type Quota struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
	Free  int64 `json:"free"`
}

type AccountQuota struct {
	AccountID  string `json:"account_id"`
	Label      string `json:"label"`
	Provider   string `json:"provider"`
	TotalBytes int64  `json:"total_bytes"`
	UsedBytes  int64  `json:"used_bytes"`
	FreeBytes  int64  `json:"free_bytes"`
}

type QuotaReport struct {
	Aggregate struct {
		TotalBytes int64 `json:"total_bytes"`
		UsedBytes  int64 `json:"used_bytes"`
		FreeBytes  int64 `json:"free_bytes"`
	} `json:"aggregate"`
	Accounts []AccountQuota `json:"accounts"`
}

// ---- Files & folders (VFS, doc 09) ----

// Folder = folder virtual; hanya hidup di DB, tak pernah dibuat di provider.
type Folder struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	ParentID  *string   `json:"parent_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// FileEntry = 1 baris di explorer. Gabungan files_index + block pertamanya.
type FileEntry struct {
	ID           string     `json:"id"`
	UserID       string     `json:"-"`
	FolderID     *string    `json:"folder_id"`
	Name         string     `json:"name"`
	VirtualPath  string     `json:"virtual_path"`
	Mime         *string    `json:"mime"`
	SizeBytes    int64      `json:"size_bytes"`
	ModifiedAt   *time.Time `json:"modified_at"`
	IsChunked    bool       `json:"is_chunked"`
	AccountID    string     `json:"account_id"`
	AccountLabel string     `json:"account_label"`
	Provider     string     `json:"provider"`
}

// FileBlock = lokasi fisik satu potongan file. Model A: tepat 1 baris (seq=0).
type FileBlock struct {
	ID          string
	FileID      string
	AccountID   string
	ProviderRef string
	Seq         int
	SizeBytes   int64
	Checksum    *string
}

// RemoteEntry = hasil listing dari engine (provider-agnostic).
type RemoteEntry struct {
	Path     string
	Name     string
	Size     int64
	MimeType string
	ModTime  time.Time
	IsDir    bool
	ID       string
}

// SearchQuery = filter untuk pencarian lintas account.
type SearchQuery struct {
	Q         string
	Mime      string
	AccountID string
	MinSize   int64
	MaxSize   int64
	FolderID  *string
	Page      int
	PerPage   int
	Sort      string
}

// MimePatterns memecah Mime jadi daftar pola. Filter kategori di UI memetakan
// satu kategori ke beberapa pola MIME ("media" = image/, video/, audio/), yang
// dikirim sebagai satu parameter dipisah koma dan dicocokkan sebagai OR.
func (q SearchQuery) MimePatterns() []string {
	if q.Mime == "" {
		return nil
	}
	out := make([]string, 0, 4)
	for _, part := range strings.Split(q.Mime, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

type FileListResult struct {
	Path  string      `json:"path"`
	Items []FileEntry `json:"items"`
	Page  int         `json:"page"`
	Total int         `json:"total"`
}

// ---- Error sentinel ----
// Dipetakan ke error code API di lapisan HTTP (doc 06 §Error Codes).

var (
	ErrNoRoom          = errors.New("no single account has enough free space")
	ErrNotFound        = errors.New("not found")
	ErrPathExists      = errors.New("path already exists")
	ErrFolderNotEmpty  = errors.New("folder is not empty")
	ErrNeedsReconnect  = errors.New("account needs reconnect")
	ErrProvider        = errors.New("provider error")
	ErrRateLimited     = errors.New("rate limited by provider")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrUnsupported     = errors.New("unsupported operation")
)
