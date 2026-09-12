// Package engine adalah adapter ke rclone. Interface Engine menyembunyikan
// rclone dari lapisan atas (ADR-002) — impl bisa diganti tanpa menyentuh StorageService.
package engine

import (
	"context"
	"io"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// Engine = kontrak operasi storage per-remote (doc 03 §3).
type Engine interface {
	// Provisioning remote (Opsi A — doc 10).
	CreateRemote(ctx context.Context, name, rType string, params map[string]any) error
	UpdateRemote(ctx context.Context, name string, params map[string]any) error
	DeleteRemote(ctx context.Context, name string) error
	ListRemotes(ctx context.Context) ([]string, error)

	// Operasi file.
	List(ctx context.Context, remote, path string, recurse bool) ([]domain.RemoteEntry, error)
	Stat(ctx context.Context, remote, path string) (*domain.RemoteEntry, error)
	About(ctx context.Context, remote string) (domain.Quota, error)
	Mkdir(ctx context.Context, remote, path string) error
	UploadStream(ctx context.Context, r io.Reader, remote, destDir, name string) error
	Download(ctx context.Context, remote, src string, w io.Writer) error
	Move(ctx context.Context, srcRemote, src, dstRemote, dst string) error
	Delete(ctx context.Context, remote, path string) error
	Ping(ctx context.Context) error
}
