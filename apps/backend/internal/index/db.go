// Package index adalah repository metadata di Postgres: accounts, tokens,
// folders (VFS), files_index, dan file_blocks.
package index

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store memegang pool koneksi dan menyediakan repo per-entitas.
type Store struct {
	pool *pgxpool.Pool
}

// Open membuka pool dan menunggu DB siap (Postgres di compose bisa telat bangun).
func Open(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("buka pool: %w", err)
	}

	var lastErr error
	for i := 0; i < 30; i++ {
		if lastErr = pool.Ping(ctx); lastErr == nil {
			return &Store{pool: pool}, nil
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	pool.Close()
	return nil, fmt.Errorf("postgres tak merespons: %w", lastErr)
}

func (s *Store) Close() { s.pool.Close() }

func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

func (s *Store) Pool() *pgxpool.Pool { return s.pool }

func (s *Store) Accounts() *AccountRepo { return &AccountRepo{pool: s.pool} }
func (s *Store) Files() *FileRepo       { return &FileRepo{pool: s.pool} }
func (s *Store) Folders() *FolderRepo   { return &FolderRepo{pool: s.pool} }

// EnsureUser membuat baris user default (mode single-user v1, doc 06).
func (s *Store) EnsureUser(ctx context.Context, id, email string) error {
	_, err := s.pool.Exec(ctx, `
		insert into users (id, email) values ($1, $2)
		on conflict (id) do nothing`, id, email)
	if err != nil {
		return fmt.Errorf("ensure user: %w", err)
	}
	return nil
}
