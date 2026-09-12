package routing

import (
	"context"
	"errors"
	"testing"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

func acct(id string, free int64, status string) domain.Account {
	return domain.Account{
		ID: id, Label: id, Status: status,
		TotalBytes: free * 2, UsedBytes: free, FreeBytes: free,
	}
}

func TestPickMostFreeMemilihAkunTerlowong(t *testing.T) {
	r := New(StrategyMostFree)
	accounts := []domain.Account{
		acct("a", 1<<30, domain.StatusActive),
		acct("b", 8<<30, domain.StatusActive),
		acct("c", 4<<30, domain.StatusActive),
	}
	got, err := r.Pick(context.Background(), accounts, 100<<20)
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if got.ID != "b" {
		t.Fatalf("mau akun b (free terbesar), dapat %s", got.ID)
	}
}

func TestPickMelewatiAkunYangTakMuat(t *testing.T) {
	r := New(StrategyMostFree)
	accounts := []domain.Account{
		acct("besar-tapi-penuh", 1<<20, domain.StatusActive),
		acct("cukup", 3<<30, domain.StatusActive),
	}
	got, err := r.Pick(context.Background(), accounts, 2<<30)
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if got.ID != "cukup" {
		t.Fatalf("mau akun cukup, dapat %s", got.ID)
	}
}

func TestPickMengabaikanAkunTakAktif(t *testing.T) {
	r := New(StrategyMostFree)
	accounts := []domain.Account{
		acct("perlu-reconnect", 100<<30, domain.StatusNeedsReconnect),
		acct("sehat", 5<<30, domain.StatusActive),
	}
	got, err := r.Pick(context.Background(), accounts, 1<<30)
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if got.ID != "sehat" {
		t.Fatalf("akun needs_reconnect tak boleh dipilih, dapat %s", got.ID)
	}
}

// ErrNoRoom adalah titik masuk Model B — pastikan sentinel-nya benar (ADR-001).
func TestPickTanpaAkunMuatMengembalikanErrNoRoom(t *testing.T) {
	r := New(StrategyMostFree)
	accounts := []domain.Account{acct("kecil", 1<<20, domain.StatusActive)}

	_, err := r.Pick(context.Background(), accounts, 20<<30)
	if !errors.Is(err, domain.ErrNoRoom) {
		t.Fatalf("mau ErrNoRoom, dapat %v", err)
	}
}

func TestPickTanpaAkunSamaSekali(t *testing.T) {
	r := New(StrategyMostFree)
	if _, err := r.Pick(context.Background(), nil, 1); !errors.Is(err, domain.ErrNoRoom) {
		t.Fatalf("mau ErrNoRoom, dapat %v", err)
	}
}

// Provider tanpa dukungan `about` (total=0) tetap layak jadi kandidat (doc 01 §9).
func TestPickMenerimaAkunTanpaInfoKuota(t *testing.T) {
	r := New(StrategyMostFree)
	accounts := []domain.Account{{ID: "s3", Status: domain.StatusActive}}

	got, err := r.Pick(context.Background(), accounts, 50<<30)
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if got.ID != "s3" {
		t.Fatalf("mau akun s3, dapat %s", got.ID)
	}
}

func TestPickRoundRobinBerputar(t *testing.T) {
	r := New(StrategyRoundRobin)
	accounts := []domain.Account{
		acct("a", 5<<30, domain.StatusActive),
		acct("b", 5<<30, domain.StatusActive),
	}
	seen := map[string]int{}
	for i := 0; i < 4; i++ {
		got, err := r.Pick(context.Background(), accounts, 1<<20)
		if err != nil {
			t.Fatalf("Pick: %v", err)
		}
		seen[got.ID]++
	}
	if seen["a"] != 2 || seen["b"] != 2 {
		t.Fatalf("round-robin tak merata: %v", seen)
	}
}
