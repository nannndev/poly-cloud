// Kontrak API Poly Cloud — cerminan TS dari Go structs di apps/backend.
// Sumber kebenaran bentuk data = Go structs; file ini diperbarui di PR yang sama
// saat API berubah (doc 02 §4).

export type AccountStatus = 'active' | 'needs_reconnect' | 'error' | 'syncing'

export type StorageProvider = 'gdrive' | 'onedrive' | 'dropbox' | 's3' | 'b2' | 'r2'

/** GET /accounts — kuota berasal dari cache DB, bukan query provider langsung. */
export interface Account {
  id: string
  provider: StorageProvider
  label: string
  status: AccountStatus
  total_bytes: number
  used_bytes: number
  free_bytes: number
  created_at: string
  last_synced: string | null
}

/** GET /providers — dipakai UI untuk tahu provider mana yang siap dipakai. */
export interface ProviderInfo {
  provider: StorageProvider
  /** 'oauth' = lewat redirect consent; 'keys' = form kredensial statis. */
  kind: 'oauth' | 'keys'
  /** false = kredensial OAuth platform belum diisi di env backend. */
  configured: boolean
  fields?: string[]
  required?: string[]
}

/** POST /accounts/connect */
export interface ConnectRequest {
  provider: StorageProvider
  label?: string
  /** Hanya untuk provider 'keys' (S3/B2/R2). */
  fields?: Record<string, string>
}

/** Respons /accounts/connect: OAuth balas auth_url, provider key balas account. */
export interface ConnectOAuthResponse {
  auth_url: string
}

export interface ConnectAccountResponse {
  account_id: string
  status: AccountStatus
  account: Account
}

/** POST /accounts/callback */
export interface CallbackRequest {
  code: string
  state: string
  provider?: StorageProvider
}

/** POST /accounts/{id}/sync */
export interface SyncResult {
  synced: boolean
  files_indexed: number
  total_bytes: number
  used_bytes: number
  free_bytes: number
}

/** Folder virtual — hidup di DB platform, tak ada di provider (doc 09). */
export interface FolderEntry {
  id: string
  parent_id: string | null
  name: string
  path: string
  created_at: string
}

export interface FileEntry {
  id: string
  folder_id: string | null
  name: string
  virtual_path: string
  mime: string | null
  size_bytes: number
  modified_at: string | null
  is_chunked: boolean
  account_id: string
  account_label: string
  provider: StorageProvider
}

/** GET /files dan GET /files/search */
export interface FileListResult {
  path: string
  items: FileEntry[]
  page: number
  total: number
}

/** POST /files/upload */
export interface UploadResponse {
  id: string
  name: string
  account_id: string
  account_label: string
  /** Strategi router yang memilih account tujuan (mis. 'most-free'). */
  routed_by: string
  job_id: string
  file: FileEntry
}

/** PATCH /files/{id} dan PATCH /folders/{id} — field dikirim hanya yang berubah.
 *  folder_id/parent_id bernilai null berarti pindah ke root. */
export interface OrganizationPatch {
  name?: string
  folder_id?: string | null
  parent_id?: string | null
}

export interface Quota {
  total_bytes: number
  used_bytes: number
  free_bytes: number
}

export interface AccountQuota extends Quota {
  account_id: string
  label: string
  provider: StorageProvider
}

/** GET /quota */
export interface QuotaReport {
  aggregate: Quota
  accounts: AccountQuota[]
}

/** GET /settings — konfigurasi runtime backend (read-only).
 *  Nilainya dibaca dari environment saat proses start; mengubahnya butuh restart. */
export interface BackendSettings {
  routing_strategy: string
  remote_base_dir: string
  /** 'A' = whole-file. Model B (chunked) belum diimplementasikan. */
  storage_model: string
  chunking: boolean
  sync_recurse: boolean
  multi_user: boolean
}

/** Event SSE di GET /events/uploads/{jobId}. */
export interface UploadProgressEvent {
  job: string
  bytes: number
  total: number
}

export interface UploadDoneEvent {
  job: string
  file_id: string
  account_label: string
}

export interface UploadErrorEvent {
  job: string
  message: string
}

export type ApiErrorCode =
  | 'NO_ROOM'
  | 'ACCOUNT_NEEDS_RECONNECT'
  | 'PROVIDER_ERROR'
  | 'NOT_FOUND'
  | 'RATE_LIMITED'
  | 'PATH_EXISTS'
  | 'FOLDER_NOT_EMPTY'
  | 'INVALID_ARGUMENT'
  | 'UNSUPPORTED'
  | 'INTERNAL'

export interface ApiError {
  error: {
    code: ApiErrorCode
    message: string
  }
}
