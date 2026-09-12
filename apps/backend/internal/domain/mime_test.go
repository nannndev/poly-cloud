package domain

import "testing"

// Tabel bawaan Go tak memuat .txt maupun ekstensi dokumen/arsip yang umum, dan
// image Alpine tak punya /etc/mime.types — tanpa tabel sendiri, MIME file-file
// ini akan kosong dan UI tak bisa memilih pratinjau.
func TestMimeByName(t *testing.T) {
	cases := map[string]string{
		"catatan.txt":    "text/plain",
		"README.md":      "text/markdown",
		"laporan.pdf":    "application/pdf",
		"data.csv":       "text/csv",
		"arsip.zip":      "application/zip",
		"foto.JPG":       "image/jpeg",
		"klip.mp4":       "video/mp4",
		"lagu.mp3":       "audio/mpeg",
		"compose.yaml":   "application/yaml",
		"tanpa-ekstensi": "",
	}
	for name, want := range cases {
		if got := MimeByName(name); got != want {
			t.Errorf("MimeByName(%q) = %q, mau %q", name, got, want)
		}
	}
}

func TestMimeByNameJatuhKeTabelSistem(t *testing.T) {
	// .html tak ada di tabel tambahan kita, tapi ada di tabel bawaan Go.
	if got := MimeByName("index.html"); got == "" {
		t.Fatal("ekstensi di luar daftar seharusnya jatuh ke tabel bawaan Go")
	}
}
