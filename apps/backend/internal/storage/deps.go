package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/config"
	"github.com/polycloud/platform/apps/backend/internal/domain"
	"github.com/polycloud/platform/apps/backend/internal/engine"
	"github.com/polycloud/platform/apps/backend/internal/index"
	"github.com/polycloud/platform/apps/backend/internal/routing"
)

// Deps mengumpulkan kolaborator yang dipakai bersama oleh Service dan
// AccountService, termasuk urusan lintas-potong: token lifecycle & kuota.
type Deps struct {
	Cfg      *config.Config
	Engine   engine.Engine
	Accounts *index.AccountRepo
	Files    *index.FileRepo
	Folders  *index.FolderRepo
	Router   *routing.Router
	OAuth    *auth.Manager
	Cipher   *auth.Cipher
	Log      *slog.Logger
	BaseDir  string
}

// EnsureRemote menjamin remote rclone milik account siap dipakai: token masih
// valid, dan bila sudah lewat masa berlaku, backend me-refresh lalu menyuntikkan
// token baru ke rclone (doc 03 §5 / doc 07 §4).
func (d *Deps) EnsureRemote(ctx context.Context, a domain.Account) error {
	if a.Status == domain.StatusNeedsReconnect {
		return fmt.Errorf("%w: account %s", domain.ErrNeedsReconnect, a.Label)
	}
	// Account berbasis key statis (S3/B2) tak punya token — tak ada yang perlu di-refresh.
	hasToken, err := d.Accounts.HasToken(ctx, a.ID)
	if err != nil {
		return err
	}
	if !hasToken {
		return nil
	}

	ts, err := d.Accounts.LoadToken(ctx, d.Cipher, a.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	if !ts.Expired() {
		return nil
	}

	fresh, err := d.OAuth.Refresh(ctx, a.Provider, ts)
	if err != nil {
		// Refresh gagal permanen → minta user re-auth (doc 03 §5).
		_ = d.Accounts.SetStatus(ctx, a.ID, domain.StatusNeedsReconnect)
		return fmt.Errorf("%w: refresh token %s gagal: %v", domain.ErrNeedsReconnect, a.Label, err)
	}
	if err := d.SaveTokenAndSyncRemote(ctx, a, fresh); err != nil {
		return err
	}
	d.Log.Info("token di-refresh", "account", a.ID, "provider", a.Provider)
	return nil
}

// SaveTokenAndSyncRemote menyimpan token terenkripsi ke DB sekaligus memperbarui
// token di remote rclone, supaya engine tak memakai token basi.
func (d *Deps) SaveTokenAndSyncRemote(ctx context.Context, a domain.Account, ts auth.TokenSet) error {
	payload, err := d.Cipher.Seal(ts)
	if err != nil {
		return err
	}
	var expires *time.Time
	if !ts.Expiry.IsZero() {
		e := ts.Expiry
		expires = &e
	}
	if err := d.Accounts.SaveToken(ctx, a.ID, payload, expires); err != nil {
		return err
	}

	app := d.Cfg.OAuthApps[a.Provider]
	_, params, err := auth.RcloneParams(a.Provider, ts, app)
	if err != nil {
		return err
	}
	if err := d.Engine.UpdateRemote(ctx, a.RcloneRemote, params); err != nil {
		return fmt.Errorf("sinkron token ke remote %s: %w", a.RcloneRemote, err)
	}
	return nil
}

// HandleEngineError menandai account yang tokennya tak lagi bisa dipakai,
// supaya UI menampilkan tombol reconnect.
func (d *Deps) HandleEngineError(ctx context.Context, a domain.Account, err error) error {
	if errors.Is(err, domain.ErrNeedsReconnect) {
		if serr := d.Accounts.SetStatus(ctx, a.ID, domain.StatusNeedsReconnect); serr != nil {
			d.Log.Error("gagal set status needs_reconnect", "account", a.ID, "err", serr)
		}
	}
	return err
}

// RefreshQuotaAsync menyegarkan cache kuota di latar belakang setelah operasi
// yang mengubah pemakaian ruang. Gagal di sini tak boleh menggagalkan request.
func (d *Deps) RefreshQuotaAsync(accountID, remote string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		q, err := d.Engine.About(ctx, remote)
		if err != nil {
			d.Log.Warn("refresh kuota gagal", "account", accountID, "err", err)
			return
		}
		if err := d.Accounts.UpdateQuota(ctx, accountID, q); err != nil {
			d.Log.Warn("simpan kuota gagal", "account", accountID, "err", err)
		}
	}()
}
