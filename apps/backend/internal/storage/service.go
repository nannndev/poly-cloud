// Package storage adalah orkestrator operasi file. Interface Service adalah
// kontrak stabil (doc 03 §2): Model A dipenuhi WholeFileStore, Model B nanti
// oleh ChunkedStore tanpa mengubah HTTP/API.
package storage

import (
	"context"
	"io"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// Progress dilaporkan selama transfer berlangsung (dipakai SSE).
type Progress func(bytes int64, total int64)

// UploadResult = hasil akhir upload, termasuk keputusan router.
type UploadResult struct {
	File         domain.FileEntry `json:"file"`
	AccountID    string           `json:"account_id"`
	AccountLabel string           `json:"account_label"`
	RoutedBy     string           `json:"routed_by"`
}

// Download membawa stream file beserta metadata untuk header respons.
type Download struct {
	Name      string
	Mime      string
	SizeBytes int64
	Body      io.ReadCloser
}

type Service interface {
	List(ctx context.Context, userID string, folderID *string, q domain.SearchQuery) (domain.FileListResult, error)
	Search(ctx context.Context, userID string, q domain.SearchQuery) (domain.FileListResult, error)
	Upload(ctx context.Context, userID string, folderID *string, name string, r io.Reader, size int64, onProgress Progress) (UploadResult, error)
	Download(ctx context.Context, userID, fileID string) (*Download, error)
	MoveToAccount(ctx context.Context, userID, fileID, destAccountID string) error
	UpdateOrganization(ctx context.Context, userID, fileID string, name *string, folderID **string) (domain.FileEntry, error)
	Delete(ctx context.Context, userID, fileID string) error
	Quota(ctx context.Context, userID string) (domain.QuotaReport, error)
}
