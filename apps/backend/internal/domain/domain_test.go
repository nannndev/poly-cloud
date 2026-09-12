package domain

import "testing"

// Akar remote S3 adalah daftar bucket, bukan tempat menaruh objek — account
// penyimpanan objek harus menargetkan bucket-nya.
func TestFsTarget(t *testing.T) {
	cases := []struct {
		name    string
		account Account
		want    string
	}{
		{"provider drive tanpa bucket", Account{RcloneRemote: "acc_1"}, "acc_1"},
		{"provider objek dengan bucket", Account{RcloneRemote: "acc_1", RootPath: "polycloud-test"}, "acc_1:polycloud-test"},
		{"bucket dengan slash dibersihkan", Account{RcloneRemote: "acc_1", RootPath: "/bucket/"}, "acc_1:bucket"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.account.FsTarget(); got != tc.want {
				t.Fatalf("FsTarget() = %q, mau %q", got, tc.want)
			}
		})
	}
}

// Filter kategori di UI memetakan satu kategori ke beberapa pola MIME; pola itu
// dikirim sebagai satu parameter dipisah koma.
func TestMimePatterns(t *testing.T) {
	cases := []struct {
		name string
		mime string
		want []string
	}{
		{"kosong", "", nil},
		{"pola tunggal", "pdf", []string{"pdf"}},
		{"beberapa pola", "image/,video/,audio/", []string{"image/", "video/", "audio/"}},
		{"spasi dan koma kosong dibuang", " pdf , ,word ", []string{"pdf", "word"}},
		{"hanya koma", ",,", []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SearchQuery{Mime: tc.mime}.MimePatterns()
			if len(got) != len(tc.want) {
				t.Fatalf("MimePatterns() = %q, mau %q", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("MimePatterns()[%d] = %q, mau %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
