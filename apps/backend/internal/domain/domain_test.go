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
