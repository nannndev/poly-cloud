package index

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

type KeyRepo struct {
	pool *pgxpool.Pool
}

const keyCols = `id, user_id, name, key_prefix, key_hash, scopes, last_used_at, expires_at, created_at`

func scanKey(row pgx.Row) (domain.APIKey, error) {
	var k domain.APIKey
	err := row.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyPrefix, &k.KeyHash, &k.Scopes, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt)
	return k, err
}

func (r *KeyRepo) List(ctx context.Context, userID string) ([]domain.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`select `+keyCols+` from api_keys where user_id = $1 order by created_at desc`, userID)
	if err != nil {
		return nil, fmt.Errorf("list api_keys: %w", err)
	}
	defer rows.Close()

	keys := make([]domain.APIKey, 0)
	for rows.Next() {
		k, err := scanKey(rows)
		if err != nil {
			return nil, fmt.Errorf("scan api_key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (r *KeyRepo) Create(ctx context.Context, userID, name, prefix, hash string, scopes []string, expiresAt *time.Time) (domain.APIKey, error) {
	if len(scopes) == 0 {
		scopes = []string{"read", "write"}
	}

	row := r.pool.QueryRow(ctx,
		`insert into api_keys (user_id, name, key_prefix, key_hash, scopes, expires_at)
		 values ($1, $2, $3, $4, $5, $6)
		 returning `+keyCols,
		userID, name, prefix, hash, scopes, expiresAt)

	k, err := scanKey(row)
	if err != nil {
		return k, fmt.Errorf("create api_key: %w", err)
	}
	return k, nil
}

func (r *KeyRepo) Delete(ctx context.Context, userID, keyID string) error {
	tag, err := r.pool.Exec(ctx,
		`delete from api_keys where id = $1 and user_id = $2`, keyID, userID)
	if err != nil {
		return fmt.Errorf("delete api_key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// FindByHash mencari key berdasarkan SHA-256 hash dan memperbarui last_used_at secara otomatis.
func (r *KeyRepo) FindByHash(ctx context.Context, hash string) (domain.APIKey, error) {
	row := r.pool.QueryRow(ctx,
		`update api_keys set last_used_at = now()
		 where key_hash = $1 and (expires_at is null or expires_at > now())
		 returning `+keyCols, hash)

	k, err := scanKey(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return k, domain.ErrNotFound
		}
		return k, fmt.Errorf("find api_key by hash: %w", err)
	}
	return k, nil
}
