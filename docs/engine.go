package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// RcloneEngine membungkus binary rclone sebagai subprocess.
// Ini implementasi Fase 0 (CLI mode). Nanti bisa diganti library mode
// tanpa mengubah pemanggil, karena dibungkus interface Engine.
type RcloneEngine struct {
	// Binary path rclone (default: "rclone" dari PATH).
	Bin string
	// ConfigPath opsional; kalau kosong, pakai default rclone (~/.config/rclone/rclone.conf).
	ConfigPath string
}

func NewRcloneEngine() *RcloneEngine {
	return &RcloneEngine{Bin: "rclone"}
}

// baseArgs menambahkan flag global (config path) bila diset.
func (e *RcloneEngine) baseArgs(args ...string) []string {
	var out []string
	if e.ConfigPath != "" {
		out = append(out, "--config", e.ConfigPath)
	}
	return append(out, args...)
}

// run mengeksekusi rclone dan mengembalikan stdout.
func (e *RcloneEngine) run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, e.Bin, e.baseArgs(args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("rclone %s: %w | stderr: %s",
			strings.Join(args, " "), err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// ---- Tipe hasil ----

// FileEntry = 1 baris di explorer. Bentuknya provider-agnostic.
type FileEntry struct {
	Name    string    `json:"Path"`
	Size    int64     `json:"Size"`
	MimeType string   `json:"MimeType"`
	ModTime time.Time `json:"ModTime"`
	IsDir   bool      `json:"IsDir"`
	ID      string    `json:"ID"` // provider ref (mis. GDrive fileId)
}

// AboutResult = quota akun (total/used/free) dalam bytes.
type AboutResult struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
	Free  int64 `json:"free"`
}

// ---- Operasi inti (dipakai StorageService nanti) ----

// ListRemotes mengembalikan nama remote yang sudah terkonfigurasi (mis. "gdrive1:").
func (e *RcloneEngine) ListRemotes(ctx context.Context) ([]string, error) {
	out, err := e.run(ctx, "listremotes")
	if err != nil {
		return nil, err
	}
	var remotes []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			remotes = append(remotes, line)
		}
	}
	return remotes, nil
}

// List membaca isi folder di remote (JSON), path relatif dari root remote.
// remote contoh: "gdrive1:", path contoh: "" (root) atau "Documents".
func (e *RcloneEngine) List(ctx context.Context, remote, path string) ([]FileEntry, error) {
	target := remote + path
	out, err := e.run(ctx, "lsjson", target)
	if err != nil {
		return nil, err
	}
	var entries []FileEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("parse lsjson: %w", err)
	}
	return entries, nil
}

// About mengembalikan quota akun. Butuh provider yang support (GDrive/Dropbox/OneDrive support).
func (e *RcloneEngine) About(ctx context.Context, remote string) (*AboutResult, error) {
	out, err := e.run(ctx, "about", remote, "--json")
	if err != nil {
		return nil, err
	}
	var res AboutResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("parse about: %w", err)
	}
	return &res, nil
}

// Upload menaruh file lokal ke remote:destPath.
// Fase 0 pakai path lokal (rclone copyto). Untuk stream io.Reader,
// nanti pakai rcat (lihat UploadStream).
func (e *RcloneEngine) Upload(ctx context.Context, localPath, remote, destPath string) error {
	target := remote + destPath
	_, err := e.run(ctx, "copyto", localPath, target, "--progress")
	return err
}

// UploadStream menaruh isi reader ke remote:destPath tanpa file lokal (rcat).
// Ini yang dipakai backend saat menerima upload dari browser (stream-through).
func (e *RcloneEngine) UploadStream(ctx context.Context, r io.Reader, remote, destPath string) error {
	target := remote + destPath
	cmd := exec.CommandContext(ctx, e.Bin, e.baseArgs("rcat", target)...)
	cmd.Stdin = r
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rcat %s: %w | %s", target, err, stderr.String())
	}
	return nil
}

// Download menstream file remote ke writer (rclone cat). Tanpa simpan permanen.
func (e *RcloneEngine) Download(ctx context.Context, remote, srcPath string, w io.Writer) error {
	target := remote + srcPath
	cmd := exec.CommandContext(ctx, e.Bin, e.baseArgs("cat", target)...)
	cmd.Stdout = w
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cat %s: %w | %s", target, err, stderr.String())
	}
	return nil
}
