package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Spike Fase 0: buktikan rclone bisa di-drive dari Go.
// Alur:
//   1. Deteksi semua remote yang sudah dikonfigurasi di rclone.
//   2. Untuk tiap remote: ambil quota (About) + list root (List).
//   3. Demonstrasikan smart routing: "kalau upload X bytes, ke akun mana?"
//   4. (Opsional) upload file argumen ke akun terpilih.
//
// Jalankan:
//   go run .                      -> hanya inspeksi (list + quota + simulasi routing)
//   go run . /path/ke/file.jpg    -> plus upload beneran ke akun terpilih
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	engine := NewRcloneEngine()

	// 1. Remote apa saja yang ada?
	remotes, err := engine.ListRemotes(ctx)
	if err != nil {
		fmt.Println("Gagal baca remotes. Sudah `rclone config` belum?")
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if len(remotes) == 0 {
		fmt.Println("Belum ada remote. Jalankan `rclone config` untuk connect akun dulu.")
		os.Exit(0)
	}
	fmt.Printf("Ditemukan %d remote: %v\n\n", len(remotes), remotes)

	// 2. Kumpulkan quota tiap akun (jadi input router).
	var accounts []Account
	var totalFree, totalUsed, totalCap int64
	for _, rem := range remotes {
		about, err := engine.About(ctx, rem)
		if err != nil {
			fmt.Printf("[%s] quota tidak tersedia (%v)\n", rem, err)
			// tetap masukkan dengan Free=0 supaya router tak memilihnya
			accounts = append(accounts, Account{Remote: rem, Label: rem})
			continue
		}
		accounts = append(accounts, Account{
			Remote: rem, Label: rem,
			Total: about.Total, Used: about.Used, Free: about.Free,
		})
		totalFree += about.Free
		totalUsed += about.Used
		totalCap += about.Total

		fmt.Printf("[%s] total=%s used=%s free=%s\n",
			rem, human(about.Total), human(about.Used), human(about.Free))

		// list beberapa item root sebagai bukti akses.
		entries, err := engine.List(ctx, rem, "")
		if err != nil {
			fmt.Printf("    list root gagal: %v\n", err)
			continue
		}
		fmt.Printf("    %d item di root. Contoh:\n", len(entries))
		for i, e := range entries {
			if i >= 3 {
				break
			}
			kind := "file"
			if e.IsDir {
				kind = "dir"
			}
			fmt.Printf("      - [%s] %s (%s)\n", kind, e.Name, human(e.Size))
		}
	}

	// Agregat — inti value proposition: banyak akun jadi satu angka.
	fmt.Printf("\n=== AGREGAT SEMUA AKUN ===\n")
	fmt.Printf("Kapasitas total: %s | Terpakai: %s | Bebas: %s\n\n",
		human(totalCap), human(totalUsed), human(totalFree))

	// 3. Simulasi smart routing.
	router := &Router{Strategy: "most-free"}
	demoSizes := []int64{50 << 20, 2 << 30, 20 << 30} // 50MB, 2GB, 20GB
	fmt.Println("=== SIMULASI SMART ROUTING (most-free) ===")
	for _, sz := range demoSizes {
		acc, err := router.Pick(ctx, accounts, sz)
		if err != nil {
			fmt.Printf("Upload %s -> %v (Model B akan pecah & sebar)\n", human(sz), err)
			continue
		}
		fmt.Printf("Upload %s -> pilih %s (free %s)\n", human(sz), acc.Remote, human(acc.Free))
	}

	// 4. Upload beneran kalau ada argumen file.
	if len(os.Args) > 1 {
		localPath := os.Args[1]
		info, err := os.Stat(localPath)
		if err != nil {
			fmt.Printf("\nFile '%s' tidak ditemukan: %v\n", localPath, err)
			return
		}
		acc, err := router.Pick(ctx, accounts, info.Size())
		if err != nil {
			fmt.Printf("\nTidak bisa upload '%s': %v\n", localPath, err)
			return
		}
		dest := "/" + info.Name()
		fmt.Printf("\nUpload '%s' (%s) -> %s%s ...\n",
			localPath, human(info.Size()), acc.Remote, dest)
		if err := engine.Upload(ctx, localPath, acc.Remote, dest); err != nil {
			fmt.Printf("Upload gagal: %v\n", err)
			return
		}
		fmt.Println("Upload sukses. Cek di akun tujuan.")
	}
}

// human memformat bytes jadi string enak dibaca.
func human(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
