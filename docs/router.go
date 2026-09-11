package main

import (
	"context"
	"errors"
	"sort"
)

// Account = akun cloud terhubung + kapasitasnya (nanti dari DB; di spike dari About()).
type Account struct {
	Remote string // nama rclone remote, mis. "gdrive1:"
	Label  string
	Total  int64
	Used   int64
	Free   int64
}

// ErrNoRoom = tak ada satu akun pun yang muat file ini.
// Di Model A: tolak / minta user pilih. Di Model B: inilah trigger chunking.
var ErrNoRoom = errors.New("no single account has enough free space")

// Router memilih akun tujuan untuk sebuah upload.
type Router struct {
	Strategy string // "most-free" (default) | "round-robin" | ...
}

// Pick mengembalikan akun terbaik untuk menampung file berukuran size.
// Model A: harus muat utuh di satu akun.
func (r *Router) Pick(_ context.Context, accounts []Account, size int64) (*Account, error) {
	// Filter yang muat.
	var candidates []Account
	for _, a := range accounts {
		if a.Free >= size {
			candidates = append(candidates, a)
		}
	}
	if len(candidates) == 0 {
		return nil, ErrNoRoom
	}

	switch r.Strategy {
	case "round-robin":
		// Placeholder: butuh state persist; sementara jatuh ke most-free.
		fallthrough
	default: // "most-free": isi akun paling lowong dulu → distribusi merata.
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Free > candidates[j].Free
		})
	}
	return &candidates[0], nil
}
