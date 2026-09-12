package index

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// FileRepo mengelola files_index + file_blocks. Model A: tepat 1 block per file
// (seq=0); struktur query di sini sudah siap untuk N block (Model B).
type FileRepo struct{ pool *pgxpool.Pool }

// fileSelect menggabungkan file dengan block pertamanya supaya explorer bisa
// menampilkan badge "file ini ada di akun mana".
const fileSelect = `
	select f.id, f.user_id, f.folder_id, f.name, f.virtual_path, f.mime,
	       f.size_bytes, f.modified_at, f.is_chunked,
	       coalesce(a.id::text,''), coalesce(a.label,''), coalesce(a.provider,'')
	from files_index f
	left join lateral (
		select account_id from file_blocks where file_id = f.id order by seq limit 1
	) b on true
	left join accounts a on a.id = b.account_id`

func scanFile(row pgx.Row) (domain.FileEntry, error) {
	var f domain.FileEntry
	err := row.Scan(&f.ID, &f.UserID, &f.FolderID, &f.Name, &f.VirtualPath, &f.Mime,
		&f.SizeBytes, &f.ModifiedAt, &f.IsChunked, &f.AccountID, &f.AccountLabel, &f.Provider)
	return f, err
}

func collectFiles(rows pgx.Rows) ([]domain.FileEntry, error) {
	defer rows.Close()
	items := make([]domain.FileEntry, 0)
	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}
		items = append(items, f)
	}
	return items, rows.Err()
}

func (r *FileRepo) Get(ctx context.Context, userID, id string) (domain.FileEntry, error) {
	f, err := scanFile(r.pool.QueryRow(ctx,
		fileSelect+` where f.user_id = $1 and f.id = $2`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return f, fmt.Errorf("%w: file %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return f, fmt.Errorf("get file: %w", err)
	}
	return f, nil
}

// orderClause memetakan parameter sort API ke ORDER BY (whitelist, bukan interpolasi bebas).
func orderClause(sort string) string {
	switch sort {
	case "name_desc":
		return " order by f.name desc"
	case "size":
		return " order by f.size_bytes asc"
	case "size_desc":
		return " order by f.size_bytes desc"
	case "modified":
		return " order by f.modified_at asc nulls last"
	case "modified_desc":
		return " order by f.modified_at desc nulls last"
	default:
		return " order by f.name asc"
	}
}

// ListByFolder mengambil isi satu folder virtual (folderID nil = root).
func (r *FileRepo) ListByFolder(ctx context.Context, userID string, folderID *string, q domain.SearchQuery) ([]domain.FileEntry, int, error) {
	where := ` where f.user_id = $1 and f.folder_id is null`
	args := []any{userID}
	if folderID != nil {
		where = ` where f.user_id = $1 and f.folder_id = $2`
		args = append(args, *folderID)
	}

	var total int
	countQ := `select count(*) from files_index f` + where
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hitung file: %w", err)
	}

	args = append(args, q.PerPage, (q.Page-1)*q.PerPage)
	sql := fileSelect + where + orderClause(q.Sort) +
		fmt.Sprintf(" limit $%d offset $%d", len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list file: %w", err)
	}
	items, err := collectFiles(rows)
	return items, total, err
}

// Search mencari lintas account dari cache DB (tak menyentuh provider).
func (r *FileRepo) Search(ctx context.Context, userID string, q domain.SearchQuery) ([]domain.FileEntry, int, error) {
	var (
		conds = []string{"f.user_id = $1"}
		args  = []any{userID}
	)
	add := func(cond string, val any) {
		args = append(args, val)
		conds = append(conds, strings.Replace(cond, "?", "$"+strconv.Itoa(len(args)), 1))
	}

	if q.Q != "" {
		add("f.name ilike ?", "%"+q.Q+"%")
	}
	if q.Mime != "" {
		add("f.mime ilike ?", "%"+q.Mime+"%")
	}
	if q.MinSize > 0 {
		add("f.size_bytes >= ?", q.MinSize)
	}
	if q.MaxSize > 0 {
		add("f.size_bytes <= ?", q.MaxSize)
	}
	if q.AccountID != "" {
		add("exists (select 1 from file_blocks fb where fb.file_id = f.id and fb.account_id = ?)", q.AccountID)
	}
	where := " where " + strings.Join(conds, " and ")

	var total int
	if err := r.pool.QueryRow(ctx, `select count(*) from files_index f`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hitung hasil search: %w", err)
	}

	args = append(args, q.PerPage, (q.Page-1)*q.PerPage)
	sql := fileSelect + where + orderClause(q.Sort) +
		fmt.Sprintf(" limit $%d offset $%d", len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search file: %w", err)
	}
	items, err := collectFiles(rows)
	return items, total, err
}

// NewFile = payload pembuatan baris index setelah upload/sync berhasil.
type NewFile struct {
	UserID      string
	FolderID    *string
	Name        string
	VirtualPath string
	Mime        *string
	SizeBytes   int64
	ModifiedAt  *time.Time
	AccountID   string
	ProviderRef string
	Checksum    *string
}

// Insert menulis files_index + file_blocks(seq=0) dalam satu transaksi.
// Upsert by (user_id, virtual_path) supaya sync berulang idempotent (doc 03 §8).
func (r *FileRepo) Insert(ctx context.Context, nf NewFile) (domain.FileEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.FileEntry{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var fileID string
	err = tx.QueryRow(ctx, `
		insert into files_index (user_id, folder_id, name, virtual_path, mime, size_bytes, modified_at, is_chunked)
		values ($1,$2,$3,$4,$5,$6,$7,false)
		on conflict (user_id, virtual_path) do update
		  -- name/folder_id sengaja tak ikut di-update: organisasi virtual milik user,
		  -- sync hanya menyegarkan metadata fisik.
		  set mime = excluded.mime, size_bytes = excluded.size_bytes,
		      modified_at = excluded.modified_at
		returning id`,
		nf.UserID, nf.FolderID, nf.Name, nf.VirtualPath, nf.Mime, nf.SizeBytes, nf.ModifiedAt).Scan(&fileID)
	if err != nil {
		return domain.FileEntry{}, fmt.Errorf("insert files_index: %w", err)
	}

	// Model A: satu file tepat satu block. Replace agar re-upload path yang sama
	// tak meninggalkan block yatim.
	if _, err := tx.Exec(ctx, `delete from file_blocks where file_id = $1`, fileID); err != nil {
		return domain.FileEntry{}, fmt.Errorf("bersihkan block lama: %w", err)
	}
	_, err = tx.Exec(ctx, `
		insert into file_blocks (file_id, account_id, provider_ref, seq, size_bytes, checksum)
		values ($1,$2,$3,0,$4,$5)`,
		fileID, nf.AccountID, nf.ProviderRef, nf.SizeBytes, nf.Checksum)
	if err != nil {
		return domain.FileEntry{}, fmt.Errorf("insert file_blocks: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.FileEntry{}, fmt.Errorf("commit: %w", err)
	}
	return r.Get(ctx, nf.UserID, fileID)
}

// Blocks mengembalikan lokasi fisik file, urut seq (Model A: 1 baris).
func (r *FileRepo) Blocks(ctx context.Context, fileID string) ([]domain.FileBlock, error) {
	rows, err := r.pool.Query(ctx, `
		select id, file_id, account_id, provider_ref, seq, size_bytes, checksum
		from file_blocks where file_id = $1 order by seq`, fileID)
	if err != nil {
		return nil, fmt.Errorf("list block: %w", err)
	}
	defer rows.Close()

	blocks := make([]domain.FileBlock, 0)
	for rows.Next() {
		var b domain.FileBlock
		if err := rows.Scan(&b.ID, &b.FileID, &b.AccountID, &b.ProviderRef, &b.Seq, &b.SizeBytes, &b.Checksum); err != nil {
			return nil, fmt.Errorf("scan block: %w", err)
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

// UpdateOrganization mengubah nama/folder file — DB murni, provider tak disentuh (doc 09 §3).
func (r *FileRepo) UpdateOrganization(ctx context.Context, userID, id string, name *string, folderID **string) (domain.FileEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.FileEntry{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		curName     string
		curFolderID *string
	)
	err = tx.QueryRow(ctx,
		`select name, folder_id from files_index where user_id = $1 and id = $2 for update`,
		userID, id).Scan(&curName, &curFolderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.FileEntry{}, fmt.Errorf("%w: file %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return domain.FileEntry{}, fmt.Errorf("lock file: %w", err)
	}

	newName := curName
	if name != nil && *name != "" {
		newName = *name
	}
	newFolder := curFolderID
	if folderID != nil {
		newFolder = *folderID
	}

	parentPath := ""
	if newFolder != nil {
		err = tx.QueryRow(ctx, `select path from folders where user_id = $1 and id = $2`,
			userID, *newFolder).Scan(&parentPath)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FileEntry{}, fmt.Errorf("%w: folder %s", domain.ErrNotFound, *newFolder)
		}
		if err != nil {
			return domain.FileEntry{}, fmt.Errorf("get folder: %w", err)
		}
	}
	newPath := JoinPath(parentPath, newName)

	_, err = tx.Exec(ctx,
		`update files_index set name = $3, folder_id = $4, virtual_path = $5 where user_id = $1 and id = $2`,
		userID, id, newName, newFolder, newPath)
	if isUniqueViolation(err) {
		return domain.FileEntry{}, fmt.Errorf("%w: %s", domain.ErrPathExists, newPath)
	}
	if err != nil {
		return domain.FileEntry{}, fmt.Errorf("update organisasi file: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.FileEntry{}, fmt.Errorf("commit: %w", err)
	}
	return r.Get(ctx, userID, id)
}

// SetBlockAccount memindahkan pencatatan block ke account lain setelah
// transfer fisik antar akun berhasil.
func (r *FileRepo) SetBlockAccount(ctx context.Context, fileID, accountID, providerRef string) error {
	_, err := r.pool.Exec(ctx,
		`update file_blocks set account_id = $2, provider_ref = $3 where file_id = $1 and seq = 0`,
		fileID, accountID, providerRef)
	if err != nil {
		return fmt.Errorf("update block account: %w", err)
	}
	return nil
}

func (r *FileRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `delete from files_index where user_id = $1 and id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("hapus file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: file %s", domain.ErrNotFound, id)
	}
	return nil
}

// DeleteByAccount membersihkan index milik satu account (dipakai saat resync & cabut akun).
func (r *FileRepo) DeleteByAccount(ctx context.Context, userID, accountID string) error {
	_, err := r.pool.Exec(ctx, `
		delete from files_index f
		where f.user_id = $1
		  and exists (select 1 from file_blocks b where b.file_id = f.id and b.account_id = $2)`,
		userID, accountID)
	if err != nil {
		return fmt.Errorf("hapus index account: %w", err)
	}
	return nil
}

// ExistsPath dipakai untuk auto-suffix nama file duplikat (doc 09 §4).
func (r *FileRepo) ExistsPath(ctx context.Context, userID, virtualPath string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`select exists(select 1 from files_index where user_id = $1 and virtual_path = $2)`,
		userID, virtualPath).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("cek path: %w", err)
	}
	return exists, nil
}
