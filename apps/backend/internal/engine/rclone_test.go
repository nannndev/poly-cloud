package engine

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

// newTestDaemon menjalankan RcloneDaemon terhadap server RC tiruan.
func newTestDaemon(t *testing.T, h http.HandlerFunc) *RcloneDaemon {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewRcloneDaemon(srv.URL, "user", "pass")
}

func TestFsOf(t *testing.T) {
	cases := map[string]string{
		"acc_x":        "acc_x:",
		"acc_x:":       "acc_x:",
		"acc_x:bucket": "acc_x:bucket", // sudah berbentuk fs, jangan ditambah ":"
	}
	for in, want := range cases {
		if got := fsOf(in); got != want {
			t.Errorf("fsOf(%q) = %q, mau %q", in, got, want)
		}
	}
}

func TestJoin(t *testing.T) {
	cases := []struct {
		parts []string
		want  string
	}{
		{[]string{"PolyCloud", "foto.jpg"}, "PolyCloud/foto.jpg"},
		{[]string{"/PolyCloud/", "/foto.jpg"}, "PolyCloud/foto.jpg"},
		{[]string{"", "foto.jpg"}, "foto.jpg"},
		{[]string{"PolyCloud", ""}, "PolyCloud"},
	}
	for _, tc := range cases {
		if got := Join(tc.parts...); got != tc.want {
			t.Errorf("Join(%v) = %q, mau %q", tc.parts, got, tc.want)
		}
	}
}

func TestListMemetakanEntri(t *testing.T) {
	d := newTestDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/operations/list" {
			t.Errorf("endpoint tak terduga: %s", r.URL.Path)
		}
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["fs"] != "acc_1:" {
			t.Errorf("fs = %v, mau acc_1:", req["fs"])
		}
		_, _ = io.WriteString(w, `{"list":[
			{"Path":"laporan.pdf","Name":"laporan.pdf","Size":2048,"MimeType":"application/pdf","ModTime":"2026-01-09T10:00:00Z","IsDir":false,"ID":"gdrive-id-1"},
			{"Path":"sub","Name":"sub","Size":-1,"IsDir":true}
		]}`)
	})

	entries, err := d.List(context.Background(), "acc_1", "PolyCloud", false)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("mau 2 entri, dapat %d", len(entries))
	}
	if entries[0].ID != "gdrive-id-1" || entries[0].Size != 2048 {
		t.Errorf("entri pertama salah: %+v", entries[0])
	}
	if entries[0].ModTime.IsZero() {
		t.Error("ModTime gagal di-parse")
	}
	if !entries[1].IsDir {
		t.Error("entri kedua seharusnya direktori")
	}
}

// Provider tanpa field `free` (mis. S3): nilainya diturunkan dari total-used.
func TestAboutMenurunkanFreeSaatTakDikirim(t *testing.T) {
	d := newTestDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"total":1000,"used":400}`)
	})

	q, err := d.About(context.Background(), "acc_1")
	if err != nil {
		t.Fatalf("About: %v", err)
	}
	if q.Free != 600 {
		t.Fatalf("free = %d, mau 600", q.Free)
	}
}

func TestAboutMemakaiFreeDariProvider(t *testing.T) {
	d := newTestDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"total":1000,"used":400,"free":550}`)
	})

	q, err := d.About(context.Background(), "acc_1")
	if err != nil {
		t.Fatalf("About: %v", err)
	}
	if q.Free != 550 {
		t.Fatalf("free = %d, mau 550 (nilai dari provider)", q.Free)
	}
}

// Error engine harus dipetakan ke sentinel domain supaya HTTP bisa memilih
// error code yang tepat tanpa tahu soal rclone (doc 03 §8).
func TestKlasifikasiError(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		sentinel error
	}{
		{"objek tak ada", 500, `{"error":"object not found"}`, domain.ErrNotFound},
		{"rate limit", 500, `{"error":"userRateLimitExceeded"}`, domain.ErrRateLimited},
		{"token kedaluwarsa", 500, `{"error":"invalid_grant: token expired"}`, domain.ErrNeedsReconnect},
		{"tak didukung", 500, `{"error":"About not supported by this backend"}`, domain.ErrUnsupported},
		{"error umum", 500, `{"error":"boom"}`, domain.ErrProvider},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newTestDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			})
			_, err := d.About(context.Background(), "acc_1")
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("mau %v, dapat %v", tc.sentinel, err)
			}
		})
	}
}

// Hapus objek yang sudah tak ada bukan kegagalan — operasi idempotent.
func TestDeleteIdempotenSaatObjekSudahHilang(t *testing.T) {
	d := newTestDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"object not found"}`)
	})

	if err := d.Delete(context.Background(), "acc_1", "PolyCloud/hilang.txt"); err != nil {
		t.Fatalf("Delete seharusnya idempotent, dapat: %v", err)
	}
}

func TestUploadStreamMengirimMultipart(t *testing.T) {
	var (
		gotFS     string
		gotRemote string
		gotBody   string
		gotAuth   bool
	)
	d := newTestDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		gotAuth = ok && user == "user" && pass == "pass"
		gotFS = r.URL.Query().Get("fs")
		gotRemote = r.URL.Query().Get("remote")

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parse multipart: %v", err)
			return
		}
		f, hdr, err := r.FormFile("file")
		if err != nil {
			t.Errorf("ambil file: %v", err)
			return
		}
		defer f.Close()
		if hdr.Filename != "catatan.txt" {
			t.Errorf("filename = %q", hdr.Filename)
		}
		b, _ := io.ReadAll(f)
		gotBody = string(b)
		_, _ = io.WriteString(w, `{}`)
	})

	err := d.UploadStream(context.Background(), strings.NewReader("isi berkas"), "acc_1", "PolyCloud", "catatan.txt")
	if err != nil {
		t.Fatalf("UploadStream: %v", err)
	}
	if !gotAuth {
		t.Error("basic auth tak terkirim ke daemon")
	}
	if gotFS != "acc_1:" || gotRemote != "PolyCloud" {
		t.Errorf("fs=%q remote=%q", gotFS, gotRemote)
	}
	if gotBody != "isi berkas" {
		t.Errorf("isi file = %q", gotBody)
	}
}

// Account penyimpanan objek menargetkan bucket, dan bentuk itu harus sampai
// utuh ke daemon.
func TestDownloadMemakaiFsBerikutBucket(t *testing.T) {
	var gotPath string
	d := newTestDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, "ok")
	})

	var buf strings.Builder
	if err := d.Download(context.Background(), "acc_1:bucket", "PolyCloud/foto.jpg", &buf); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if gotPath != "/[acc_1:bucket]/PolyCloud/foto.jpg" {
		t.Errorf("path = %q, mau /[acc_1:bucket]/PolyCloud/foto.jpg", gotPath)
	}
}

func TestDownloadMenstreamLewatServeHTTP(t *testing.T) {
	var gotPath string
	d := newTestDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, "konten file")
	})

	var buf strings.Builder
	if err := d.Download(context.Background(), "acc_1", "PolyCloud/foto.jpg", &buf); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if buf.String() != "konten file" {
		t.Errorf("isi = %q", buf.String())
	}
	// Nama fs harus dikurung siku, kalau tidak daemon menjawab 404.
	if gotPath != "/[acc_1:]/PolyCloud/foto.jpg" {
		t.Errorf("path = %q, mau /[acc_1:]/PolyCloud/foto.jpg", gotPath)
	}
}

// Nama berkas dengan spasi/karakter khusus harus lolos utuh ke daemon.
func TestDownloadMengescapeNamaBerkas(t *testing.T) {
	var gotPath string
	d := newTestDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, "ok")
	})

	var buf strings.Builder
	if err := d.Download(context.Background(), "acc_1", "PolyCloud/laporan akhir #2.pdf", &buf); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if gotPath != "/[acc_1:]/PolyCloud/laporan akhir #2.pdf" {
		t.Errorf("path ter-decode = %q", gotPath)
	}
}

func TestCreateRemoteMengirimParameter(t *testing.T) {
	var req map[string]any
	d := newTestDaemon(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config/create" {
			t.Errorf("endpoint = %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_, _ = io.WriteString(w, `{}`)
	})

	err := d.CreateRemote(context.Background(), "acc_abc", "drive", map[string]any{"token": "{}"})
	if err != nil {
		t.Fatalf("CreateRemote: %v", err)
	}
	if req["name"] != "acc_abc" || req["type"] != "drive" {
		t.Errorf("payload salah: %+v", req)
	}
	opt, _ := req["opt"].(map[string]any)
	if opt["nonInteractive"] != true {
		t.Error("nonInteractive wajib true agar daemon tak menunggu input wizard")
	}
}

func TestStatMengembalikanNotFoundSaatItemKosong(t *testing.T) {
	d := newTestDaemon(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"item":null}`)
	})

	_, err := d.Stat(context.Background(), "acc_1", "PolyCloud/x.txt")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("mau ErrNotFound, dapat %v", err)
	}
}
