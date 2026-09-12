package httpapi

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// Body PATCH harus membedakan tiga hal: field tak dikirim, field bernilai null
// (pindah ke root), dan field berisi UUID.
func TestParseOrganizationPatch(t *testing.T) {
	t.Run("rename saja", func(t *testing.T) {
		name, folder, err := parseOrganizationPatch(map[string]any{"name": "baru.pdf"})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if name == nil || *name != "baru.pdf" {
			t.Fatalf("name = %v", name)
		}
		if folder != nil {
			t.Fatal("folder_id tak dikirim, seharusnya nil")
		}
	})

	t.Run("pindah ke folder", func(t *testing.T) {
		_, folder, err := parseOrganizationPatch(map[string]any{"folder_id": "fol-1"})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if folder == nil || *folder == nil || **folder != "fol-1" {
			t.Fatalf("folder = %v", folder)
		}
	})

	t.Run("null berarti pindah ke root", func(t *testing.T) {
		_, folder, err := parseOrganizationPatch(map[string]any{"folder_id": nil})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if folder == nil {
			t.Fatal("folder_id dikirim null, seharusnya pointer non-nil")
		}
		if *folder != nil {
			t.Fatalf("nilai dalam seharusnya nil (root), dapat %v", *folder)
		}
	})

	t.Run("body kosong ditolak", func(t *testing.T) {
		_, _, err := parseOrganizationPatch(map[string]any{})
		if !errors.Is(err, domain.ErrInvalidArgument) {
			t.Fatalf("mau ErrInvalidArgument, dapat %v", err)
		}
	})

	t.Run("field tak dikenal ditolak", func(t *testing.T) {
		_, _, err := parseOrganizationPatch(map[string]any{"size_bytes": 10})
		if !errors.Is(err, domain.ErrInvalidArgument) {
			t.Fatalf("mau ErrInvalidArgument, dapat %v", err)
		}
	})

	t.Run("name kosong ditolak", func(t *testing.T) {
		_, _, err := parseOrganizationPatch(map[string]any{"name": "   "})
		if !errors.Is(err, domain.ErrInvalidArgument) {
			t.Fatalf("mau ErrInvalidArgument, dapat %v", err)
		}
	})
}

func TestParseQueryDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/files", nil)
	q := parseQuery(r)
	if q.Page != 1 || q.PerPage != 100 {
		t.Fatalf("default page/per_page salah: %d/%d", q.Page, q.PerPage)
	}
}

func TestParseQueryMembatasiPerPage(t *testing.T) {
	for _, raw := range []string{"0", "-5", "9999", "bukan-angka"} {
		r := httptest.NewRequest("GET", "/api/v1/files?per_page="+raw, nil)
		if got := parseQuery(r).PerPage; got != 100 {
			t.Errorf("per_page=%s → %d, mau dibatasi ke 100", raw, got)
		}
	}
}

func TestParseQueryMembacaFilter(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/v1/files/search?q=laporan&type=pdf&min_size=100&account_id=acc-1&sort=size_desc&page=3", nil)
	q := parseQuery(r)

	if q.Q != "laporan" || q.Mime != "pdf" || q.MinSize != 100 || q.AccountID != "acc-1" {
		t.Fatalf("filter tak terbaca: %+v", q)
	}
	if q.Sort != "size_desc" || q.Page != 3 {
		t.Fatalf("sort/page tak terbaca: %+v", q)
	}
}

func TestWriteErrorMemetakanKeCodeAPI(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrNoRoom, 409, "NO_ROOM"},
		{domain.ErrNeedsReconnect, 401, "ACCOUNT_NEEDS_RECONNECT"},
		{domain.ErrNotFound, 404, "NOT_FOUND"},
		{domain.ErrPathExists, 409, "PATH_EXISTS"},
		{domain.ErrFolderNotEmpty, 409, "FOLDER_NOT_EMPTY"},
		{domain.ErrRateLimited, 429, "RATE_LIMITED"},
		{domain.ErrProvider, 502, "PROVIDER_ERROR"},
		{domain.ErrInvalidArgument, 400, "INVALID_ARGUMENT"},
		{errors.New("kaget"), 500, "INTERNAL"},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Error dibungkus untuk meniru pemakaian nyata (%w berlapis).
			writeError(w, discardLogger(), errors.Join(tc.err, errors.New("konteks tambahan")))

			if w.Code != tc.status {
				t.Errorf("status = %d, mau %d", w.Code, tc.status)
			}
			if !strings.Contains(w.Body.String(), `"code":"`+tc.code+`"`) {
				t.Errorf("body = %s, mau code %s", w.Body.String(), tc.code)
			}
		})
	}
}

func TestQueryBool(t *testing.T) {
	cases := map[string]bool{"?inline=1": true, "?inline=true": true, "?inline=0": false, "": false, "?inline=abc": false}
	for raw, want := range cases {
		r := httptest.NewRequest("GET", "/api/v1/files/x/download"+raw, nil)
		if got := queryBool(r, "inline"); got != want {
			t.Errorf("queryBool(%q) = %v, mau %v", raw, got, want)
		}
	}
}
