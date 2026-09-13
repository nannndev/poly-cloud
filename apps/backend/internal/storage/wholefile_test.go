package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/polycloud/platform/apps/backend/internal/domain"
	"github.com/polycloud/platform/apps/backend/internal/engine"
)

// fakeEngine mencatat panggilan supaya perilaku store bisa diperiksa tanpa rclone.
type fakeEngine struct {
	objects    map[string][]byte // "remote:path" -> isi
	uploads    []uploadCall
	deletes    []string
	moves      []moveCall
	quota      map[string]domain.Quota
	failUpload error
}

type uploadCall struct {
	Remote, Dir, Name string
	Body              string
}

type moveCall struct {
	SrcRemote, Src, DstRemote, Dst string
}

func newFakeEngine() *fakeEngine {
	return &fakeEngine{objects: map[string][]byte{}, quota: map[string]domain.Quota{}}
}

func key(remote, path string) string { return remote + ":" + path }

func (f *fakeEngine) CreateRemote(context.Context, string, string, map[string]any) error { return nil }
func (f *fakeEngine) UpdateRemote(context.Context, string, map[string]any) error         { return nil }
func (f *fakeEngine) DeleteRemote(context.Context, string) error                         { return nil }
func (f *fakeEngine) ListRemotes(context.Context) ([]string, error)                      { return nil, nil }
func (f *fakeEngine) Mkdir(context.Context, string, string) error                        { return nil }
func (f *fakeEngine) Ping(context.Context) error                                         { return nil }

func (f *fakeEngine) List(context.Context, string, string, bool) ([]domain.RemoteEntry, error) {
	return nil, nil
}

func (f *fakeEngine) Stat(_ context.Context, remote, path string) (*domain.RemoteEntry, error) {
	body, ok := f.objects[key(remote, path)]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &domain.RemoteEntry{Path: path, Name: path, Size: int64(len(body)), ID: "ref-" + path}, nil
}

func (f *fakeEngine) About(_ context.Context, remote string) (domain.Quota, error) {
	return f.quota[remote], nil
}

func (f *fakeEngine) UploadStream(_ context.Context, r io.Reader, remote, dir, name string) error {
	if f.failUpload != nil {
		return f.failUpload
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	f.uploads = append(f.uploads, uploadCall{Remote: remote, Dir: dir, Name: name, Body: string(body)})
	f.objects[key(remote, engine.Join(dir, name))] = body
	return nil
}

func (f *fakeEngine) Download(_ context.Context, remote, src string, w io.Writer) error {
	body, ok := f.objects[key(remote, src)]
	if !ok {
		return domain.ErrNotFound
	}
	_, err := w.Write(body)
	return err
}

func (f *fakeEngine) Move(_ context.Context, srcRemote, src, dstRemote, dst string) error {
	body, ok := f.objects[key(srcRemote, src)]
	if !ok {
		return domain.ErrNotFound
	}
	delete(f.objects, key(srcRemote, src))
	f.objects[key(dstRemote, dst)] = body
	f.moves = append(f.moves, moveCall{srcRemote, src, dstRemote, dst})
	return nil
}

func (f *fakeEngine) Delete(_ context.Context, remote, path string) error {
	f.deletes = append(f.deletes, key(remote, path))
	delete(f.objects, key(remote, path))
	return nil
}

// Upload harus mengalirkan byte apa adanya ke engine, tanpa menyentuh disk.
func TestUploadStreamMeneruskanIsiUtuh(t *testing.T) {
	f := newFakeEngine()
	content := strings.Repeat("poly", 1000)

	err := f.UploadStream(context.Background(), strings.NewReader(content), "acc_1", "PolyCloud", "besar.txt")
	if err != nil {
		t.Fatalf("UploadStream: %v", err)
	}
	if len(f.uploads) != 1 || f.uploads[0].Body != content {
		t.Fatal("isi file tak sampai utuh ke engine")
	}
}

func TestProgressReaderMelaporkanKemajuan(t *testing.T) {
	var reports []int64
	pr := &progressReader{
		r:     strings.NewReader(strings.Repeat("x", 1000)),
		total: 1000,
		fn:    func(sent, _ int64) { reports = append(reports, sent) },
	}

	n, err := io.Copy(io.Discard, pr)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if n != 1000 {
		t.Fatalf("byte terbaca = %d", n)
	}
	if len(reports) == 0 {
		t.Fatal("tak ada laporan progres")
	}
	if last := reports[len(reports)-1]; last != 1000 {
		t.Fatalf("laporan terakhir = %d, mau 1000", last)
	}
}

// Laporan progres harus monoton naik — UI memakainya untuk progress bar.
func TestProgressReaderMonotonNaik(t *testing.T) {
	var reports []int64
	pr := &progressReader{
		r:     bytes.NewReader(make([]byte, 100_000)),
		total: 100_000,
		fn:    func(sent, _ int64) { reports = append(reports, sent) },
	}
	if _, err := io.Copy(io.Discard, pr); err != nil {
		t.Fatalf("copy: %v", err)
	}
	for i := 1; i < len(reports); i++ {
		if reports[i] <= reports[i-1] {
			t.Fatalf("progres tak naik di indeks %d: %d setelah %d", i, reports[i], reports[i-1])
		}
	}
}

// objectPath memilih antara path (provider tanpa file-id) dan file-id provider.
func TestObjectPath(t *testing.T) {
	s := &WholeFileStore{deps: &Deps{BaseDir: "PolyCloud"}}

	cases := []struct {
		name, providerRef, want string
	}{
		{"ref lama tanpa path dianggap nama di base dir", "gdrive-abc123", "PolyCloud/gdrive-abc123"},
		{"ref berbentuk path dipakai apa adanya", "PolyCloud/foto.jpg", "PolyCloud/foto.jpg"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s.objectPath(domain.FileBlock{ProviderRef: tc.providerRef})
			if got != tc.want {
				t.Fatalf("objectPath = %q, mau %q", got, tc.want)
			}
		})
	}
}

// provider_ref yang dicatat harus berupa path objek, bukan file-id provider.
// Engine menjangkau objek lewat rclone serve yang hanya mengerti path, jadi
// menyimpan file-id membuat download menjawab 404 — dan karena header sudah
// terkirim lebih dulu, klien menerima 200 dengan badan kosong: pratinjau
// tampak blank tanpa satu pun pesan galat.
func TestObjectPathMenemukanObjekDariRefYangDicatat(t *testing.T) {
	const (
		remote = "acc_1:"
		dir    = "PolyCloud"
		name   = "foto.png"
	)

	eng := newFakeEngine()
	ctx := context.Background()
	if err := eng.UploadStream(ctx, strings.NewReader("isi"), remote, dir, name); err != nil {
		t.Fatalf("UploadStream: %v", err)
	}

	objectPath := engine.Join(dir, name)
	stat, err := eng.Stat(ctx, remote, objectPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	// Bentuk yang dulu dicatat Upload (stat.ID) vs bentuk yang dicatat sekarang.
	store := &WholeFileStore{deps: &Deps{BaseDir: dir}}
	for _, tc := range []struct {
		name, ref string
		wantFound bool
	}{
		{"path objek — bentuk yang dicatat sekarang", objectPath, true},
		{"file-id provider — bentuk lama yang bikin 404", stat.ID, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolved := store.objectPath(domain.FileBlock{ProviderRef: tc.ref})
			var buf bytes.Buffer
			err := eng.Download(ctx, remote, resolved, &buf)
			if tc.wantFound {
				if err != nil {
					t.Fatalf("objek tak terjangkau lewat %q: %v", resolved, err)
				}
				if buf.String() != "isi" {
					t.Fatalf("isi = %q, mau %q", buf.String(), "isi")
				}
				return
			}
			if err == nil {
				t.Fatalf("ref %q seharusnya tak menemukan objek, tapi berhasil", tc.ref)
			}
		})
	}
}
