package domain

import (
	"mime"
	"path"
	"strings"
)

// extraMimeTypes melengkapi tabel bawaan Go, yang hanya memuat sedikit ekstensi
// web (html/css/js/png/…) dan tak punya .txt sekalipun. Image runtime kita
// (Alpine) juga tak menyertakan /etc/mime.types, jadi tanpa tabel ini sebagian
// besar file akan tercatat tanpa MIME dan UI tak bisa memilih pratinjau.
var extraMimeTypes = map[string]string{
	".txt": "text/plain", ".log": "text/plain", ".md": "text/markdown",
	".csv": "text/csv", ".tsv": "text/tab-separated-values",
	".yaml": "application/yaml", ".yml": "application/yaml",
	".toml": "application/toml", ".ini": "text/plain", ".env": "text/plain",
	".sql": "application/sql", ".sh": "application/x-sh",
	".go": "text/x-go", ".py": "text/x-python", ".rb": "text/x-ruby",
	".java": "text/x-java", ".c": "text/x-c", ".h": "text/x-c",
	".cpp": "text/x-c++", ".rs": "text/x-rust", ".php": "text/x-php",
	".ts": "text/x-typescript", ".tsx": "text/x-typescript",
	".jsx": "text/javascript", ".vue": "text/plain",

	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

	".zip": "application/zip", ".tar": "application/x-tar",
	".gz": "application/gzip", ".tgz": "application/gzip",
	".bz2": "application/x-bzip2", ".7z": "application/x-7z-compressed",
	".rar": "application/vnd.rar",

	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	".gif": "image/gif", ".webp": "image/webp", ".svg": "image/svg+xml",
	".bmp": "image/bmp", ".ico": "image/x-icon", ".heic": "image/heic",
	".avif": "image/avif", ".tiff": "image/tiff",

	".mp4": "video/mp4", ".mov": "video/quicktime", ".mkv": "video/x-matroska",
	".webm": "video/webm", ".avi": "video/x-msvideo",

	".mp3": "audio/mpeg", ".wav": "audio/wav", ".flac": "audio/flac",
	".ogg": "audio/ogg", ".m4a": "audio/mp4", ".aac": "audio/aac",
}

// MimeByName menebak MIME dari nama berkas. Mengembalikan "" bila tak dikenali,
// supaya pemanggil bisa menyimpan NULL alih-alih menebak asal.
func MimeByName(name string) string {
	ext := strings.ToLower(path.Ext(name))
	if ext == "" {
		return ""
	}
	if t, ok := extraMimeTypes[ext]; ok {
		return t
	}
	// Tabel sistem/bawaan Go sebagai cadangan untuk ekstensi di luar daftar.
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return ""
}
