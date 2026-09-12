package auth

import (
	"crypto/sha256"
	"testing"
	"time"
)

func testCipher(t *testing.T) *Cipher {
	t.Helper()
	c, err := NewCipher(sha256.Sum256([]byte("kunci-uji-yang-cukup-panjang")))
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	return c
}

func TestSealOpenPulangPergi(t *testing.T) {
	c := testCipher(t)
	want := TokenSet{
		AccessToken: "ya29.access", TokenType: "Bearer",
		RefreshToken: "1//refresh", Expiry: time.Now().Add(time.Hour).UTC().Truncate(time.Second),
	}

	blob, err := c.Seal(want)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	got, err := c.Open(blob)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken {
		t.Fatalf("token tak utuh: %+v", got)
	}
	if !got.Expiry.Equal(want.Expiry) {
		t.Fatalf("expiry berubah: %v != %v", got.Expiry, want.Expiry)
	}
}

// Nonce acak per-Seal: ciphertext yang sama tak boleh berulang.
func TestSealMenghasilkanCiphertextBerbeda(t *testing.T) {
	c := testCipher(t)
	ts := TokenSet{AccessToken: "sama"}

	a, err := c.Seal(ts)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	b, err := c.Seal(ts)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if string(a) == string(b) {
		t.Fatal("dua Seal menghasilkan ciphertext identik — nonce tak acak")
	}
}

func TestOpenMenolakPayloadRusak(t *testing.T) {
	c := testCipher(t)
	blob, err := c.Seal(TokenSet{AccessToken: "rahasia"})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	blob[len(blob)-1] ^= 0xff // balik satu bit di ciphertext

	if _, err := c.Open(blob); err == nil {
		t.Fatal("payload yang diubah seharusnya ditolak GCM")
	}
}

func TestOpenMenolakPayloadTerlaluPendek(t *testing.T) {
	c := testCipher(t)
	if _, err := c.Open([]byte{1, 2, 3}); err == nil {
		t.Fatal("payload pendek seharusnya ditolak")
	}
}

func TestKunciBerbedaTakBisaMembuka(t *testing.T) {
	a := testCipher(t)
	b, err := NewCipher(sha256.Sum256([]byte("kunci-lain-yang-berbeda")))
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	blob, err := a.Seal(TokenSet{AccessToken: "x"})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if _, err := b.Open(blob); err == nil {
		t.Fatal("kunci berbeda seharusnya gagal membuka")
	}
}

func TestExpired(t *testing.T) {
	cases := []struct {
		name   string
		expiry time.Time
		want   bool
	}{
		{"tanpa expiry dianggap tak kedaluwarsa", time.Time{}, false},
		{"masih lama", time.Now().Add(time.Hour), false},
		{"sudah lewat", time.Now().Add(-time.Minute), true},
		{"dalam margin aman 60 detik", time.Now().Add(30 * time.Second), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := (TokenSet{Expiry: tc.expiry}).Expired(); got != tc.want {
				t.Fatalf("Expired() = %v, mau %v", got, tc.want)
			}
		})
	}
}
