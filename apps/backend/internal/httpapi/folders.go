package httpapi

import (
	"net/http"
	"strings"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// Folder handler: seluruh operasi di sini transaksi DB murni, tak menyentuh
// provider (doc 09 §3). Satu pengecualian: hapus rekursif ikut menghapus
// objek fisik file di dalamnya.

func (a *API) ListFolders(w http.ResponseWriter, r *http.Request) {
	folders, err := a.folders.List(r.Context(), a.userID(r), optionalID(r, "parent_id"))
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, folders)
}

type createFolderRequest struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

func (a *API) CreateFolder(w http.ResponseWriter, r *http.Request) {
	var req createFolderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, a.log, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || strings.Contains(req.Name, "/") {
		writeError(w, a.log, errInvalid("nama folder wajib diisi dan tak boleh memuat '/'"))
		return
	}
	folder, err := a.folders.Create(r.Context(), a.userID(r), req.Name, req.ParentID)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, folder)
}

func (a *API) PatchFolder(w http.ResponseWriter, r *http.Request) {
	var raw map[string]any
	if err := decodeJSON(r, &raw); err != nil {
		writeError(w, a.log, err)
		return
	}
	name, parentID, err := parseOrganizationPatch(raw)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	if name != nil && strings.Contains(*name, "/") {
		writeError(w, a.log, errInvalid("nama folder tak boleh memuat '/'"))
		return
	}
	folder, err := a.folders.Update(r.Context(), a.userID(r), r.PathValue("id"), name, parentID)
	if err != nil {
		writeError(w, a.log, err)
		return
	}
	writeJSON(w, http.StatusOK, folder)
}

// DeleteFolder menolak folder berisi kecuali ?recursive=true. Mode rekursif
// menghapus objek fisik tiap file di dalamnya, lalu barisnya (cascade).
func (a *API) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		userID = a.userID(r)
		id     = r.PathValue("id")
	)
	if _, err := a.folders.Get(ctx, userID, id); err != nil {
		writeError(w, a.log, err)
		return
	}

	if !queryBool(r, "recursive") {
		subfolders, files, err := a.folders.Counts(ctx, userID, id)
		if err != nil {
			writeError(w, a.log, err)
			return
		}
		if subfolders > 0 || files > 0 {
			writeError(w, a.log, wrap(domain.ErrFolderNotEmpty, "folder tak kosong; pakai ?recursive=true"))
			return
		}
	} else {
		fileIDs, err := a.folders.DescendantFileIDs(ctx, userID, id)
		if err != nil {
			writeError(w, a.log, err)
			return
		}
		for _, fileID := range fileIDs {
			if err := a.files.Delete(ctx, userID, fileID); err != nil {
				writeError(w, a.log, err)
				return
			}
		}
	}

	if err := a.folders.Delete(ctx, userID, id); err != nil {
		writeError(w, a.log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func wrap(sentinel error, msg string) error {
	return &wrappedErr{sentinel: sentinel, msg: msg}
}
