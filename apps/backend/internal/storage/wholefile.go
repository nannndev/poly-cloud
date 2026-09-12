package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/polycloud/platform/apps/backend/internal/domain"
	"github.com/polycloud/platform/apps/backend/internal/engine"
	"github.com/polycloud/platform/apps/backend/internal/index"
)

// WholeFileStore = implementasi Model A: 1 file = 1 objek utuh di 1 account (ADR-001).
type WholeFileStore struct {
	deps *Deps
}

var _ Service = (*WholeFileStore)(nil)

func NewWholeFileStore(d *Deps) *WholeFileStore { return &WholeFileStore{deps: d} }

func (s *WholeFileStore) List(ctx context.Context, userID string, folderID *string, q domain.SearchQuery) (domain.FileListResult, error) {
	items, total, err := s.deps.Files.ListByFolder(ctx, userID, folderID, q)
	if err != nil {
		return domain.FileListResult{}, err
	}
	path := "/"
	if folderID != nil {
		if f, err := s.deps.Folders.Get(ctx, userID, *folderID); err == nil {
			path = f.Path
		}
	}
	return domain.FileListResult{Path: path, Items: items, Page: q.Page, Total: total}, nil
}

func (s *WholeFileStore) Search(ctx context.Context, userID string, q domain.SearchQuery) (domain.FileListResult, error) {
	items, total, err := s.deps.Files.Search(ctx, userID, q)
	if err != nil {
		return domain.FileListResult{}, err
	}
	return domain.FileListResult{Path: "/", Items: items, Page: q.Page, Total: total}, nil
}

// Upload: router pilih account → stream ke provider → catat index (doc 07 §2).
// File tak pernah mendarat di disk server.
func (s *WholeFileStore) Upload(ctx context.Context, userID string, folderID *string, name string, r io.Reader, size int64, onProgress Progress) (UploadResult, error) {
	accounts, err := s.deps.Accounts.ListActive(ctx, userID)
	if err != nil {
		return UploadResult{}, err
	}
	picked, err := s.deps.Router.Pick(ctx, accounts, size)
	if err != nil {
		return UploadResult{}, err
	}

	// Pastikan token remote masih segar sebelum menyentuh provider.
	if err := s.deps.EnsureRemote(ctx, *picked); err != nil {
		return UploadResult{}, err
	}

	// Path virtual final (untuk index) — dihitung dulu supaya bentrok nama
	// ketahuan sebelum data terlanjur terkirim.
	parentPath := ""
	if folderID != nil {
		folder, err := s.deps.Folders.Get(ctx, userID, *folderID)
		if err != nil {
			return UploadResult{}, err
		}
		parentPath = folder.Path
	}
	finalName, err := s.uniqueName(ctx, userID, parentPath, name)
	if err != nil {
		return UploadResult{}, err
	}
	virtualPath := index.JoinPath(parentPath, finalName)

	// Nama objek fisik diberi prefix unik supaya file dari folder virtual berbeda
	// tak bertabrakan di satu folder flat provider (doc 09 §2).
	objectName := finalName
	baseDir := s.deps.BaseDir

	reader := io.Reader(r)
	if onProgress != nil {
		reader = &progressReader{r: r, total: size, fn: onProgress}
	}

	if err := s.deps.Engine.UploadStream(ctx, reader, picked.FsTarget(), baseDir, objectName); err != nil {
		return UploadResult{}, s.deps.HandleEngineError(ctx, *picked, err)
	}

	// Ambil metadata final dari provider (ukuran & ref sebenarnya).
	objectPath := engine.Join(baseDir, objectName)
	var (
		providerRef string
		realSize    = size
	)
	if stat, err := s.deps.Engine.Stat(ctx, picked.FsTarget(), objectPath); err == nil {
		providerRef = stat.ID
		if stat.Size > 0 {
			realSize = stat.Size
		}
	}
	if providerRef == "" {
		providerRef = objectPath
	}

	mimeType := domain.MimeByName(finalName)
	var mimePtr *string
	if mimeType != "" {
		mimePtr = &mimeType
	}

	entry, err := s.deps.Files.Insert(ctx, index.NewFile{
		UserID: userID, FolderID: folderID, Name: finalName, VirtualPath: virtualPath,
		Mime: mimePtr, SizeBytes: realSize, AccountID: picked.ID, ProviderRef: providerRef,
	})
	if err != nil {
		return UploadResult{}, err
	}

	// Kuota berubah setelah upload — segarkan cache supaya routing berikutnya akurat.
	s.deps.RefreshQuotaAsync(picked.ID, picked.FsTarget())

	return UploadResult{
		File: entry, AccountID: picked.ID, AccountLabel: picked.Label,
		RoutedBy: s.deps.Router.Strategy,
	}, nil
}

// uniqueName menambah suffix " (n)" bila path virtual sudah dipakai (doc 09 §4).
func (s *WholeFileStore) uniqueName(ctx context.Context, userID, parentPath, name string) (string, error) {
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 1; i < 100; i++ {
		exists, err := s.deps.Files.ExistsPath(ctx, userID, index.JoinPath(parentPath, candidate))
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s (%d)%s", stem, i, ext)
	}
	return "", fmt.Errorf("%w: terlalu banyak file bernama sama", domain.ErrPathExists)
}

// Download menstream objek dari account fisiknya (doc 07 §3). Pipe dipakai agar
// byte mengalir langsung ke response writer tanpa buffer penuh.
func (s *WholeFileStore) Download(ctx context.Context, userID, fileID string) (*Download, error) {
	file, err := s.deps.Files.Get(ctx, userID, fileID)
	if err != nil {
		return nil, err
	}
	blocks, err := s.deps.Files.Blocks(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("%w: file %s tak punya block", domain.ErrNotFound, fileID)
	}
	if file.IsChunked || len(blocks) > 1 {
		return nil, fmt.Errorf("%w: file chunked butuh ChunkedStore (Model B)", domain.ErrUnsupported)
	}

	account, err := s.deps.Accounts.Get(ctx, userID, blocks[0].AccountID)
	if err != nil {
		return nil, err
	}
	if err := s.deps.EnsureRemote(ctx, account); err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		err := s.deps.Engine.Download(ctx, account.FsTarget(), s.objectPath(blocks[0]), pw)
		_ = pw.CloseWithError(err)
	}()

	mimeType := "application/octet-stream"
	if file.Mime != nil && *file.Mime != "" {
		mimeType = *file.Mime
	}
	return &Download{Name: file.Name, Mime: mimeType, SizeBytes: file.SizeBytes, Body: pr}, nil
}

// objectPath mengembalikan path objek di dalam remote. provider_ref bisa berupa
// file-id provider (GDrive) atau path — bentuk path dipakai kalau tersedia.
func (s *WholeFileStore) objectPath(b domain.FileBlock) string {
	if strings.Contains(b.ProviderRef, "/") {
		return b.ProviderRef
	}
	return engine.Join(s.deps.BaseDir, b.ProviderRef)
}

// MoveToAccount memindahkan file secara fisik ke account lain (transfer data nyata).
// Beda dari pindah folder VFS yang cuma update DB.
func (s *WholeFileStore) MoveToAccount(ctx context.Context, userID, fileID, destAccountID string) error {
	blocks, err := s.deps.Files.Blocks(ctx, fileID)
	if err != nil {
		return err
	}
	if len(blocks) != 1 {
		return fmt.Errorf("%w: hanya file whole (Model A) yang bisa dipindah", domain.ErrUnsupported)
	}
	if blocks[0].AccountID == destAccountID {
		return nil
	}

	file, err := s.deps.Files.Get(ctx, userID, fileID)
	if err != nil {
		return err
	}
	src, err := s.deps.Accounts.Get(ctx, userID, blocks[0].AccountID)
	if err != nil {
		return err
	}
	dst, err := s.deps.Accounts.Get(ctx, userID, destAccountID)
	if err != nil {
		return err
	}
	if dst.TotalBytes > 0 && dst.FreeBytes < file.SizeBytes {
		return fmt.Errorf("%w: account tujuan tak cukup ruang", domain.ErrNoRoom)
	}
	if err := s.deps.EnsureRemote(ctx, src); err != nil {
		return err
	}
	if err := s.deps.EnsureRemote(ctx, dst); err != nil {
		return err
	}

	srcPath := s.objectPath(blocks[0])
	dstPath := engine.Join(s.deps.BaseDir, file.Name)
	if err := s.deps.Engine.Move(ctx, src.FsTarget(), srcPath, dst.FsTarget(), dstPath); err != nil {
		return s.deps.HandleEngineError(ctx, src, err)
	}

	providerRef := dstPath
	if stat, err := s.deps.Engine.Stat(ctx, dst.FsTarget(), dstPath); err == nil && stat.ID != "" {
		providerRef = stat.ID
	}
	if err := s.deps.Files.SetBlockAccount(ctx, fileID, dst.ID, providerRef); err != nil {
		return err
	}

	s.deps.RefreshQuotaAsync(src.ID, src.FsTarget())
	s.deps.RefreshQuotaAsync(dst.ID, dst.FsTarget())
	return nil
}

// UpdateOrganization = rename / pindah folder. DB murni, provider tak disentuh.
func (s *WholeFileStore) UpdateOrganization(ctx context.Context, userID, fileID string, name *string, folderID **string) (domain.FileEntry, error) {
	return s.deps.Files.UpdateOrganization(ctx, userID, fileID, name, folderID)
}

// Delete menghapus objek fisik lalu barisnya di index.
func (s *WholeFileStore) Delete(ctx context.Context, userID, fileID string) error {
	blocks, err := s.deps.Files.Blocks(ctx, fileID)
	if err != nil {
		return err
	}
	for _, b := range blocks {
		account, err := s.deps.Accounts.Get(ctx, userID, b.AccountID)
		if err != nil {
			// Account sudah hilang → tak ada objek untuk dihapus; index tetap dibersihkan.
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return err
		}
		if err := s.deps.EnsureRemote(ctx, account); err != nil {
			return err
		}
		if err := s.deps.Engine.Delete(ctx, account.FsTarget(), s.objectPath(b)); err != nil {
			return s.deps.HandleEngineError(ctx, account, err)
		}
		s.deps.RefreshQuotaAsync(account.ID, account.FsTarget())
	}
	return s.deps.Files.Delete(ctx, userID, fileID)
}

// Quota membaca kuota dari cache DB (doc 03 §6) — cepat, tak hit provider.
func (s *WholeFileStore) Quota(ctx context.Context, userID string) (domain.QuotaReport, error) {
	accounts, err := s.deps.Accounts.List(ctx, userID)
	if err != nil {
		return domain.QuotaReport{}, err
	}
	var report domain.QuotaReport
	report.Accounts = make([]domain.AccountQuota, 0, len(accounts))
	for _, a := range accounts {
		report.Aggregate.TotalBytes += a.TotalBytes
		report.Aggregate.UsedBytes += a.UsedBytes
		report.Aggregate.FreeBytes += a.FreeBytes
		report.Accounts = append(report.Accounts, domain.AccountQuota{
			AccountID: a.ID, Label: a.Label, Provider: a.Provider,
			TotalBytes: a.TotalBytes, UsedBytes: a.UsedBytes, FreeBytes: a.FreeBytes,
		})
	}
	return report, nil
}

// progressReader melaporkan byte yang sudah lewat ke callback (untuk SSE).
type progressReader struct {
	r     io.Reader
	total int64
	read  int64
	fn    Progress
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.read += int64(n)
		p.fn(p.read, p.total)
	}
	return n, err
}
