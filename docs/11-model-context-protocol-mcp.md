# 11 — Model Context Protocol (MCP) Integration

Poly Cloud menyediakan implementasi standar **Model Context Protocol (MCP)** (protokol versi `2024-11-05`), memungkinkan asisten AI seperti **Claude Desktop**, **Cursor**, **Antigravity**, **VS Code**, dan AI agent lainnya berinteraksi langsung dengan seluruh cloud drive Anda melalui 1 antarmuka terpadu.

---

## 1. Arsitektur & Transpor

Poly Cloud MCP mendukung dua mode transpor:

```
[Claude Desktop / Cursor / Antigravity]
                  │
                  ▼ (stdio JSON-RPC)
       [cmd/mcp (stdio bridge)]
                  │
                  ▼ (HTTP POST /api/v1/mcp)
     [Poly Cloud Go Backend (api)]
                  │
    ┌─────────────┼─────────────┐
    ▼             ▼             ▼
[Postgres VFS] [Smart Router] [rclone Engine]
                                │
                   ┌────────────┼────────────┐
                   ▼            ▼            ▼
             [Google Drive] [OneDrive]    [AWS S3]
```

1. **Stdio Bridge (`apps/backend/cmd/mcp`)**:
   Digunakan untuk integrasi lokal dengan Claude Desktop, Cursor, dan tools AI desktop yang mengeksekusi sub-proses via stdin/stdout.
2. **HTTP Endpoint (`POST /api/v1/mcp`)**:
   Endpoint HTTP standar yang menerima payload JSON-RPC 2.0 untuk agen AI jarak jauh atau layanan containerized.

---

## 2. Daftar Tools MCP

Poly Cloud mengekspos 8 tools bawaan:

| Tool | Parameter | Fungsi |
|---|---|---|
| `search_files` | `query` (wajib), `account_id`, `limit` | Mencari berkas di seluruh cloud terhubung lewat indeks PostgreSQL lokal. |
| `list_files` | `folder_id`, `account_id` | Menampilkan struktur folder virtual dan berkas di direktori tertentu. |
| `read_file` | `file_id` (wajib), `max_bytes` | Membaca isi dokumen (teks/markdown/kode) atau mengembalikan base64 jika biner. |
| `upload_file` | `name` (wajib), `content` (wajib), `folder_id` | Mengunggah berkas baru dengan penempatan otomatis (*smart routing*) ke akun paling lega. |
| `get_storage_quota` | *tidak ada* | Melihat total kapasitas, sisa ruang, dan rincian per-provider. |
| `list_accounts` | *tidak ada* | Melihat status semua provider cloud yang terhubung. |
| `create_folder` | `name` (wajib), `parent_id` | Membuat virtual folder baru di VFS. |
| `delete_file` | `file_id` (wajib) | Menghapus berkas dari cloud penyimpanan. |

---

## 3. Sumber Daya Live (Resources)

- `polycloud://quota`: Laporan kuota penyimpanan multi-cloud real-time.
- `polycloud://accounts`: Daftar akun cloud yang aktif dan tersambung.

---

## 4. Panduan Menghubungkan ke Client AI

### Claude Desktop

Tambahkan konfigurasi berikut ke berkas konfigurasi Claude Desktop Anda (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "poly-cloud": {
      "command": "go",
      "args": ["run", "./apps/backend/cmd/mcp"],
      "cwd": "/path/to/poly-cloud",
      "env": {
        "POLYCLOUD_API_URL": "http://localhost:8080"
      }
    }
  }
}
```

### Cursor

Masuk ke **Cursor Settings** > **Features** > **MCP**:
- **Type**: `command`
- **Command**: `go run ./apps/backend/cmd/mcp`
- **Environment**: `POLYCLOUD_API_URL=http://localhost:8080`

### Antigravity IDE

Poly Cloud otomatis terkonfigurasi di `.agents/mcp_config.json` dan skill set di `.agents/skills/poly-cloud-mcp/SKILL.md`.
