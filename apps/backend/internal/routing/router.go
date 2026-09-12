// Package routing memilih account tujuan untuk sebuah upload (ADR-008).
package routing

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"

	"github.com/polycloud/platform/apps/backend/internal/domain"
)

const (
	StrategyMostFree   = "most-free"
	StrategyRoundRobin = "round-robin"
)

// Router menyimpan strategi aktif. Aman dipakai bersamaan antar-goroutine.
type Router struct {
	Strategy string
	rr       atomic.Uint64 // kursor round-robin
}

func New(strategy string) *Router {
	if strategy == "" {
		strategy = StrategyMostFree
	}
	return &Router{Strategy: strategy}
}

// Pick memilih account yang muat menampung file berukuran size.
// Model A: file harus muat utuh di satu account; kalau tak ada → ErrNoRoom,
// yang nanti jadi titik masuk chunking Model B.
func (r *Router) Pick(_ context.Context, accounts []domain.Account, size int64) (*domain.Account, error) {
	candidates := make([]domain.Account, 0, len(accounts))
	for _, a := range accounts {
		if a.Status != domain.StatusActive {
			continue
		}
		// Kuota belum diketahui (provider tanpa `about`) → tetap dianggap kandidat.
		if a.TotalBytes == 0 || a.FreeBytes >= size {
			candidates = append(candidates, a)
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w (butuh %d byte)", domain.ErrNoRoom, size)
	}

	switch r.Strategy {
	case StrategyRoundRobin:
		idx := r.rr.Add(1) - 1
		picked := candidates[idx%uint64(len(candidates))]
		return &picked, nil
	default: // most-free: isi account terlowong dulu → distribusi merata.
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].FreeBytes > candidates[j].FreeBytes
		})
		picked := candidates[0]
		return &picked, nil
	}
}
