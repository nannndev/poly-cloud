package index

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// FolderRepo mengelola VFS (doc 09). Semua operasi di sini transaksi DB murni —
// tak pernah menyentuh engine/provider.
type FolderRepo struct{ pool *pgxpool.Pool }

const folderCols = `id, user_id, parent_id, name, path, created_at`

func scanFolder(row pgx.Row) (domain.Folder, error) {
	var f domain.Folder
	err := row.Scan(&f.ID, &f.UserID, &f.ParentID, &f.Name, &f.Path, &f.CreatedAt)
	return f, err
}

// isUniqueViolation mendeteksi tabrakan constraint unique (path/virtual_path duplikat).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// JoinPath menyusun path materialized dari path induk + nama.
func JoinPath(parent, name string) string {
	name = strings.Trim(name, "/")
	if parent == "" || parent == "/" {
		return "/" + name
	}
	return strings.TrimSuffix(parent, "/") + "/" + name
}

// List mengembalikan subfolder langsung di bawah parentID (nil = root).
func (r *FolderRepo) List(ctx context.Context, userID string, parentID *string) ([]domain.Folder, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if parentID == nil {
		rows, err = r.pool.Query(ctx,
			`select `+folderCols+` from folders where user_id = $1 and parent_id is null order by name`, userID)
	} else {
		rows, err = r.pool.Query(ctx,
			`select `+folderCols+` from folders where user_id = $1 and parent_id = $2 order by name`, userID, *parentID)
	}
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	defer rows.Close()

	folders := make([]domain.Folder, 0)
	for rows.Next() {
		f, err := scanFolder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		folders = append(folders, f)
	}
	return folders, rows.Err()
}

func (r *FolderRepo) Get(ctx context.Context, userID, id string) (domain.Folder, error) {
	f, err := scanFolder(r.pool.QueryRow(ctx,
		`select `+folderCols+` from folders where user_id = $1 and id = $2`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return f, fmt.Errorf("%w: folder %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return f, fmt.Errorf("get folder: %w", err)
	}
	return f, nil
}

// GetByPath mencari folder lewat path materialized (dipakai untuk ?path= di API).
func (r *FolderRepo) GetByPath(ctx context.Context, userID, path string) (domain.Folder, error) {
	f, err := scanFolder(r.pool.QueryRow(ctx,
		`select `+folderCols+` from folders where user_id = $1 and path = $2`, userID, path))
	if errors.Is(err, pgx.ErrNoRows) {
		return f, fmt.Errorf("%w: folder %s", domain.ErrNotFound, path)
	}
	if err != nil {
		return f, fmt.Errorf("get folder by path: %w", err)
	}
	return f, nil
}

func (r *FolderRepo) Create(ctx context.Context, userID, name string, parentID *string) (domain.Folder, error) {
	parentPath := ""
	if parentID != nil {
		parent, err := r.Get(ctx, userID, *parentID)
		if err != nil {
			return domain.Folder{}, err
		}
		parentPath = parent.Path
	}

	path := JoinPath(parentPath, name)
	f, err := scanFolder(r.pool.QueryRow(ctx, `
		insert into folders (user_id, parent_id, name, path) values ($1,$2,$3,$4)
		returning `+folderCols, userID, parentID, name, path))
	if isUniqueViolation(err) {
		return f, fmt.Errorf("%w: %s", domain.ErrPathExists, path)
	}
	if err != nil {
		return f, fmt.Errorf("create folder: %w", err)
	}
	return f, nil
}

// Update melakukan rename dan/atau pindah induk, lalu me-recompute path
// folder ini beserta seluruh turunannya (doc 09 §5).
func (r *FolderRepo) Update(ctx context.Context, userID, id string, name *string, parentID **string) (domain.Folder, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cur, err := scanFolder(tx.QueryRow(ctx,
		`select `+folderCols+` from folders where user_id = $1 and id = $2 for update`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Folder{}, fmt.Errorf("%w: folder %s", domain.ErrNotFound, id)
	}
	if err != nil {
		return domain.Folder{}, fmt.Errorf("lock folder: %w", err)
	}

	newName := cur.Name
	if name != nil && *name != "" {
		newName = *name
	}
	newParent := cur.ParentID
	if parentID != nil {
		newParent = *parentID
	}

	// Cegah folder dipindah ke dalam dirinya sendiri (siklus).
	if newParent != nil {
		if *newParent == id {
			return domain.Folder{}, fmt.Errorf("%w: folder tak bisa jadi induk dirinya sendiri", domain.ErrInvalidArgument)
		}
		var descendant bool
		err = tx.QueryRow(ctx, `
			with recursive sub as (
				select id from folders where id = $1
				union all
				select f.id from folders f join sub on f.parent_id = sub.id
			)
			select exists(select 1 from sub where id = $2)`, id, *newParent).Scan(&descendant)
		if err != nil {
			return domain.Folder{}, fmt.Errorf("cek siklus folder: %w", err)
		}
		if descendant {
			return domain.Folder{}, fmt.Errorf("%w: folder tak bisa dipindah ke dalam turunannya", domain.ErrInvalidArgument)
		}
	}

	parentPath := ""
	if newParent != nil {
		p, err := scanFolder(tx.QueryRow(ctx,
			`select `+folderCols+` from folders where user_id = $1 and id = $2`, userID, *newParent))
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Folder{}, fmt.Errorf("%w: folder induk %s", domain.ErrNotFound, *newParent)
		}
		if err != nil {
			return domain.Folder{}, fmt.Errorf("get parent: %w", err)
		}
		parentPath = p.Path
	}
	newPath := JoinPath(parentPath, newName)

	updated, err := scanFolder(tx.QueryRow(ctx, `
		update folders set name = $3, parent_id = $4, path = $5
		where user_id = $1 and id = $2 returning `+folderCols,
		userID, id, newName, newParent, newPath))
	if isUniqueViolation(err) {
		return domain.Folder{}, fmt.Errorf("%w: %s", domain.ErrPathExists, newPath)
	}
	if err != nil {
		return domain.Folder{}, fmt.Errorf("update folder: %w", err)
	}

	if cur.Path != newPath {
		if err := recomputeSubtree(ctx, tx, userID, cur.Path, newPath); err != nil {
			return domain.Folder{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Folder{}, fmt.Errorf("commit: %w", err)
	}
	return updated, nil
}

// recomputeSubtree mengganti prefix path pada seluruh turunan folder & file
// setelah induknya pindah/rename. Adjacency tetap sumber kebenaran; path cuma cache.
//
// Prefix lama diganti lewat overlay posisi-tetap, bukan replace() global, supaya
// nama folder yang kebetulan muncul lagi di tengah path tak ikut tertukar.
// Panjang prefix di-cast eksplisit ke int: tanpa itu Postgres tak bisa meng-infer
// tipe parameter di posisi substring dan menolak nilainya sebagai text.
func recomputeSubtree(ctx context.Context, tx pgx.Tx, userID, oldPath, newPath string) error {
	oldPrefix := strings.TrimSuffix(oldPath, "/") + "/"
	newPrefix := strings.TrimSuffix(newPath, "/") + "/"
	tailFrom := len(oldPrefix) + 1

	_, err := tx.Exec(ctx, `
		update folders set path = $3 || substring(path from $4::int)
		where user_id = $1 and path like $2`,
		userID, oldPrefix+"%", newPrefix, tailFrom)
	if err != nil {
		return fmt.Errorf("recompute path folder turunan: %w", err)
	}

	_, err = tx.Exec(ctx, `
		update files_index set virtual_path = $3 || substring(virtual_path from $4::int)
		where user_id = $1 and virtual_path like $2`,
		userID, oldPrefix+"%", newPrefix, tailFrom)
	if err != nil {
		return fmt.Errorf("recompute path file turunan: %w", err)
	}
	return nil
}

// Counts melaporkan jumlah subfolder & file langsung di dalam folder.
func (r *FolderRepo) Counts(ctx context.Context, userID, id string) (subfolders, files int, err error) {
	err = r.pool.QueryRow(ctx, `
		select (select count(*) from folders where user_id = $1 and parent_id = $2),
		       (select count(*) from files_index where user_id = $1 and folder_id = $2)`,
		userID, id).Scan(&subfolders, &files)
	if err != nil {
		return 0, 0, fmt.Errorf("hitung isi folder: %w", err)
	}
	return subfolders, files, nil
}

// DescendantFileIDs mengembalikan semua file di dalam folder dan turunannya —
// dipakai hapus rekursif untuk tahu objek fisik mana yang harus dihapus.
func (r *FolderRepo) DescendantFileIDs(ctx context.Context, userID, id string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		with recursive sub as (
			select id from folders where user_id = $1 and id = $2
			union all
			select f.id from folders f join sub on f.parent_id = sub.id
		)
		select fi.id from files_index fi join sub on fi.folder_id = sub.id`, userID, id)
	if err != nil {
		return nil, fmt.Errorf("list file turunan: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var fileID string
		if err := rows.Scan(&fileID); err != nil {
			return nil, fmt.Errorf("scan file id: %w", err)
		}
		ids = append(ids, fileID)
	}
	return ids, rows.Err()
}

func (r *FolderRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `delete from folders where user_id = $1 and id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("hapus folder: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: folder %s", domain.ErrNotFound, id)
	}
	return nil
}
