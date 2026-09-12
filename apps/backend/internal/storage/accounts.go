package storage

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/domain"
	"github.com/polycloud/platform/apps/backend/internal/engine"
	"github.com/polycloud/platform/apps/backend/internal/index"
)

// AccountService mengurus siklus hidup account: connect (OAuth / key),
// provisioning remote rclone, sync index & kuota, dan pencabutan.
type AccountService struct {
	deps *Deps
}

func NewAccountService(d *Deps) *AccountService { return &AccountService{deps: d} }

func (s *AccountService) List(ctx context.Context, userID string) ([]domain.Account, error) {
	return s.deps.Accounts.List(ctx, userID)
}

func (s *AccountService) Get(ctx context.Context, userID, id string) (domain.Account, error) {
	return s.deps.Accounts.Get(ctx, userID, id)
}

// StartConnect mengembalikan URL consent provider (doc 07 §1 langkah 2-3).
func (s *AccountService) StartConnect(ctx context.Context, userID, provider, label string) (string, error) {
	return s.deps.OAuth.AuthURL(provider, label, userID)
}

// CompleteConnect menukar code jadi token, menyimpan account + token, lalu
// membuat remote rclone dengan token itu (doc 10 §3 langkah 5-9).
func (s *AccountService) CompleteConnect(ctx context.Context, code, state string) (domain.Account, error) {
	st, err := s.deps.OAuth.ConsumeState(state)
	if err != nil {
		return domain.Account{}, err
	}
	provider := st.Provider_()
	label := st.Label_()
	userID := st.UserID_()
	if label == "" {
		label = provider
	}

	ts, err := s.deps.OAuth.Exchange(ctx, provider, code)
	if err != nil {
		return domain.Account{}, err
	}

	account, err := s.deps.Accounts.Create(ctx, domain.Account{
		UserID: userID, Provider: provider, Label: label,
		Status: domain.StatusSyncing,
	})
	if err != nil {
		return domain.Account{}, err
	}
	// Nama remote diturunkan dari UUID account: unik, tak bocor identitas,
	// tak bentrok antar akun/provider/user (doc 10 §7).
	remote := account.RcloneRemote

	app := s.deps.Cfg.OAuthApps[provider]
	rcloneType, params, err := auth.RcloneParams(provider, ts, app)
	if err != nil {
		s.cleanup(ctx, userID, account.ID, "")
		return domain.Account{}, err
	}
	if err := s.deps.Engine.CreateRemote(ctx, remote, rcloneType, params); err != nil {
		s.cleanup(ctx, userID, account.ID, "")
		return domain.Account{}, fmt.Errorf("provisioning remote: %w", err)
	}

	payload, err := s.deps.Cipher.Seal(ts)
	if err != nil {
		s.cleanup(ctx, userID, account.ID, remote)
		return domain.Account{}, err
	}
	var expires *time.Time
	if !ts.Expiry.IsZero() {
		e := ts.Expiry
		expires = &e
	}
	if err := s.deps.Accounts.SaveToken(ctx, account.ID, payload, expires); err != nil {
		s.cleanup(ctx, userID, account.ID, remote)
		return domain.Account{}, err
	}

	// Siapkan folder fisik tempat objek ditaruh (doc 09 §2). Provider yang tak
	// punya konsep folder akan membuatnya implisit saat upload pertama.
	if err := s.deps.Engine.Mkdir(ctx, account.FsTarget(), s.deps.BaseDir); err != nil {
		s.deps.Log.Warn("mkdir base dir gagal (dilanjut)", "remote", remote, "err", err)
	}

	if err := s.deps.Accounts.SetStatus(ctx, account.ID, domain.StatusActive); err != nil {
		return domain.Account{}, err
	}
	account.Status = domain.StatusActive
	return account, nil
}

// ConnectWithKeys menghubungkan provider non-OAuth (S3/B2/R2) memakai kredensial
// statis dari form (doc 10 §4).
func (s *AccountService) ConnectWithKeys(ctx context.Context, userID, provider, label string, fields map[string]string) (domain.Account, error) {
	spec, ok := auth.KeyProviders[provider]
	if !ok {
		return domain.Account{}, fmt.Errorf("%w: provider %q tak mendukung koneksi berbasis key", domain.ErrInvalidArgument, provider)
	}
	for _, req := range spec.RequiredFields() {
		if strings.TrimSpace(fields[req]) == "" {
			return domain.Account{}, fmt.Errorf("%w: field %q wajib diisi", domain.ErrInvalidArgument, req)
		}
	}
	// Provider objek menaruh semua file di dalam satu bucket; nama bucket jadi
	// akar penyimpanan account, bukan parameter remote.
	rootPath := ""
	if spec.NeedsBucket {
		rootPath = strings.Trim(strings.TrimSpace(fields["bucket"]), "/")
	}
	// Hanya field yang dikenal yang diteruskan ke rclone.
	params := map[string]any{}
	for _, f := range spec.Fields {
		if v := strings.TrimSpace(fields[f]); v != "" {
			params[f] = v
		}
	}
	if provider == "r2" {
		params["provider"] = "Cloudflare"
	}
	if label == "" {
		label = provider
	}

	account, err := s.deps.Accounts.Create(ctx, domain.Account{
		UserID: userID, Provider: provider, Label: label,
		RootPath: rootPath, Status: domain.StatusSyncing,
	})
	if err != nil {
		return domain.Account{}, err
	}
	remote := account.RcloneRemote

	if err := s.deps.Engine.CreateRemote(ctx, remote, spec.RcloneType, params); err != nil {
		s.cleanup(ctx, userID, account.ID, "")
		return domain.Account{}, fmt.Errorf("provisioning remote: %w", err)
	}

	if err := s.deps.Engine.Mkdir(ctx, account.FsTarget(), s.deps.BaseDir); err != nil {
		s.deps.Log.Warn("mkdir base dir gagal (dilanjut)", "remote", remote, "err", err)
	}
	if err := s.deps.Accounts.SetStatus(ctx, account.ID, domain.StatusActive); err != nil {
		return domain.Account{}, err
	}
	account.Status = domain.StatusActive
	return account, nil
}

// cleanup membatalkan provisioning setengah jadi agar tak meninggalkan
// account yatim atau remote nyasar.
func (s *AccountService) cleanup(ctx context.Context, userID, accountID, remote string) {
	if remote != "" {
		if err := s.deps.Engine.DeleteRemote(ctx, remote); err != nil {
			s.deps.Log.Warn("rollback hapus remote gagal", "remote", remote, "err", err)
		}
	}
	if err := s.deps.Accounts.Delete(ctx, userID, accountID); err != nil {
		s.deps.Log.Warn("rollback hapus account gagal", "account", accountID, "err", err)
	}
}

// SyncResult melaporkan hasil satu putaran sync.
type SyncResult struct {
	Synced       bool  `json:"synced"`
	FilesIndexed int   `json:"files_indexed"`
	TotalBytes   int64 `json:"total_bytes"`
	UsedBytes    int64 `json:"used_bytes"`
	FreeBytes    int64 `json:"free_bytes"`
}

// Sync membaca ulang kuota + isi folder PolyCloud di satu account dan menulisnya
// ke index (doc 03 §6). File yang sudah ada di-upsert by virtual_path, jadi
// operasi ini idempotent dan aman diulang.
func (s *AccountService) Sync(ctx context.Context, userID, accountID string) (SyncResult, error) {
	account, err := s.deps.Accounts.Get(ctx, userID, accountID)
	if err != nil {
		return SyncResult{}, err
	}
	if err := s.deps.EnsureRemote(ctx, account); err != nil {
		return SyncResult{}, err
	}

	var result SyncResult

	// Kuota dulu — berguna untuk routing walau listing gagal.
	if q, err := s.deps.Engine.About(ctx, account.FsTarget()); err == nil {
		if err := s.deps.Accounts.UpdateQuota(ctx, account.ID, q); err != nil {
			return result, err
		}
		result.TotalBytes, result.UsedBytes = q.Total, q.Used
		result.FreeBytes = q.Total - q.Used
	} else {
		// Provider tanpa dukungan `about` bukan kondisi fatal (doc 01 §9).
		s.deps.Log.Warn("about tak didukung/gagal", "account", account.ID, "err", err)
	}

	entries, err := s.deps.Engine.List(ctx, account.FsTarget(), s.deps.BaseDir, s.deps.Cfg.SyncRecurse)
	if err != nil {
		return result, s.deps.HandleEngineError(ctx, account, err)
	}

	for _, e := range entries {
		if e.IsDir {
			continue
		}
		name := e.Name
		if name == "" {
			name = path.Base(e.Path)
		}
		virtualPath := index.JoinPath("", name)

		// Jangan timpa organisasi yang sudah dibuat user: kalau path ini sudah
		// terdaftar, biarkan baris yang ada (upsert hanya menyegarkan metadata).
		// Provider yang melaporkan MIME dipercaya; sisanya ditebak dari nama.
		mimeType := e.MimeType
		if mimeType == "" {
			mimeType = domain.MimeByName(name)
		}
		var mimePtr *string
		if mimeType != "" {
			mimePtr = &mimeType
		}
		var modPtr *time.Time
		if !e.ModTime.IsZero() {
			m := e.ModTime
			modPtr = &m
		}
		providerRef := e.ID
		if providerRef == "" {
			providerRef = engine.Join(s.deps.BaseDir, e.Path)
		}

		if _, err := s.deps.Files.Insert(ctx, index.NewFile{
			UserID: userID, FolderID: nil, Name: name, VirtualPath: virtualPath,
			Mime: mimePtr, SizeBytes: e.Size, ModifiedAt: modPtr,
			AccountID: account.ID, ProviderRef: providerRef,
		}); err != nil {
			s.deps.Log.Warn("index file gagal", "file", name, "err", err)
			continue
		}
		result.FilesIndexed++
	}

	if err := s.deps.Accounts.TouchSynced(ctx, account.ID, time.Now()); err != nil {
		return result, err
	}
	if account.Status != domain.StatusActive {
		if err := s.deps.Accounts.SetStatus(ctx, account.ID, domain.StatusActive); err != nil {
			return result, err
		}
	}
	result.Synced = true
	return result, nil
}

// SyncAll menyegarkan semua account milik user; error per-account dicatat
// tapi tak menghentikan yang lain.
func (s *AccountService) SyncAll(ctx context.Context, userID string) (map[string]SyncResult, error) {
	accounts, err := s.deps.Accounts.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]SyncResult, len(accounts))
	for _, a := range accounts {
		res, err := s.Sync(ctx, userID, a.ID)
		if err != nil {
			s.deps.Log.Warn("sync account gagal", "account", a.ID, "err", err)
		}
		out[a.ID] = res
	}
	return out, nil
}

// Disconnect mencabut account: hapus remote rclone, index terkait, lalu barisnya
// (token & block ikut terhapus lewat cascade).
func (s *AccountService) Disconnect(ctx context.Context, userID, accountID string) error {
	account, err := s.deps.Accounts.Get(ctx, userID, accountID)
	if err != nil {
		return err
	}
	if err := s.deps.Files.DeleteByAccount(ctx, userID, accountID); err != nil {
		return err
	}
	if account.RcloneRemote != "" {
		if err := s.deps.Engine.DeleteRemote(ctx, account.RcloneRemote); err != nil {
			s.deps.Log.Warn("hapus remote gagal (dilanjut)", "remote", account.RcloneRemote, "err", err)
		}
	}
	return s.deps.Accounts.Delete(ctx, userID, accountID)
}
