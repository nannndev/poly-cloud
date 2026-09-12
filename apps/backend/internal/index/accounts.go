package index

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polycloud/platform/apps/backend/internal/auth"
	"github.com/polycloud/platform/apps/backend/internal/domain"
)

type AccountRepo struct{ pool *pgxpool.Pool }

const accountCols = `id, user_id, provider, label, rclone_remote, root_path,
	coalesce(total_bytes,0), coalesce(used_bytes,0), status, created_at, last_synced`

func scanAccount(row pgx.Row) (domain.Account, error) {
	var a domain.Account
	err := row.Scan(&a.ID, &a.UserID, &a.Provider, &a.Label, &a.RcloneRemote, &a.RootPath,
		&a.TotalBytes, &a.UsedBytes, &a.Status, &a.CreatedAt, &a.LastSynced)
	if err != nil {
		return a, err
	}
	a.FreeBytes = a.TotalBytes - a.UsedBytes
	if a.FreeBytes < 0 {
		a.FreeBytes = 0
	}
	return a, nil
}

func (r *AccountRepo) List(ctx context.Context, userID string) ([]domain.Account, error) {
	rows, err := r.pool.Query(ctx,
		`select `+accountCols+` from accounts where user_id = $1 order by created_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.Account, 0)
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// ListActive mengembalikan kandidat routing: account siap pakai saja.
func (r *AccountRepo) ListActive(ctx context.Context, userID string) ([]domain.Account, error) {
	rows, err := r.pool.Query(ctx,
		`select `+accountCols+` from accounts
		 where user_id = $1 and status = $2 order by created_at`, userID, domain.StatusActive)
	if err != nil {
		return nil, fmt.Errorf("list active accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.Account, 0)
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (r *AccountRepo) Get(ctx context.Context, userID, id string) (domain.Account, error) {
	a, err := scanAccount(r.pool.QueryRow(ctx,
		`select `+accountCols+` from accounts where user_id = $1 and id = $2`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return a, fmt.Errorf("%w: account %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return a, fmt.Errorf("get account: %w", err)
	}
	return a, nil
}

// Create menyisipkan account sekaligus menetapkan nama remote yang diturunkan
// dari UUID-nya (doc 10 §7) — dilakukan di satu statement supaya tak pernah ada
// baris dengan remote placeholder.
func (r *AccountRepo) Create(ctx context.Context, a domain.Account) (domain.Account, error) {
	row := r.pool.QueryRow(ctx, `
		with new as (select gen_random_uuid() as id)
		insert into accounts (id, user_id, provider, label, rclone_remote, root_path, total_bytes, used_bytes, status)
		select new.id, $1, $2, $3, 'acc_' || replace(new.id::text, '-', ''), $4, $5, $6, $7 from new
		returning `+accountCols,
		a.UserID, a.Provider, a.Label, a.RootPath, a.TotalBytes, a.UsedBytes, a.Status)
	created, err := scanAccount(row)
	if err != nil {
		return created, fmt.Errorf("create account: %w", err)
	}
	return created, nil
}

func (r *AccountRepo) UpdateQuota(ctx context.Context, id string, q domain.Quota) error {
	_, err := r.pool.Exec(ctx, `
		update accounts set total_bytes = $2, used_bytes = $3, last_synced = now()
		where id = $1`, id, q.Total, q.Used)
	if err != nil {
		return fmt.Errorf("update quota: %w", err)
	}
	return nil
}

func (r *AccountRepo) SetStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `update accounts set status = $2 where id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("set status: %w", err)
	}
	return nil
}

func (r *AccountRepo) TouchSynced(ctx context.Context, id string, t time.Time) error {
	_, err := r.pool.Exec(ctx, `update accounts set last_synced = $2 where id = $1`, id, t)
	if err != nil {
		return fmt.Errorf("touch synced: %w", err)
	}
	return nil
}

func (r *AccountRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `delete from accounts where user_id = $1 and id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: account %s", domain.ErrNotFound, id)
	}
	return nil
}

// ---- Token (account_tokens) ----

func (r *AccountRepo) SaveToken(ctx context.Context, accountID string, payload []byte, expiresAt *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		insert into account_tokens (account_id, enc_payload, expires_at) values ($1,$2,$3)
		on conflict (account_id) do update set enc_payload = excluded.enc_payload,
		                                        expires_at  = excluded.expires_at`,
		accountID, payload, expiresAt)
	if err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	return nil
}

func (r *AccountRepo) LoadToken(ctx context.Context, c *auth.Cipher, accountID string) (auth.TokenSet, error) {
	var payload []byte
	err := r.pool.QueryRow(ctx,
		`select enc_payload from account_tokens where account_id = $1`, accountID).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.TokenSet{}, fmt.Errorf("%w: token account %s", domain.ErrNotFound, accountID)
	}
	if err != nil {
		return auth.TokenSet{}, fmt.Errorf("load token: %w", err)
	}
	return c.Open(payload)
}

// HasToken dipakai untuk membedakan account OAuth (punya token) dari
// account berbasis key statis seperti S3/B2.
func (r *AccountRepo) HasToken(ctx context.Context, accountID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`select exists(select 1 from account_tokens where account_id = $1)`, accountID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("cek token: %w", err)
	}
	return exists, nil
}
