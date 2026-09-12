// Kontrak API — cerminan Go structs di apps/backend.
// Sumber kebenaran: packages/shared-types/types.ts (lihat docs/02 §4).

export type AccountStatus = 'active' | 'needs_reconnect' | 'error' | 'syncing'

export type StorageProvider = 'gdrive' | 'onedrive' | 'dropbox' | 's3' | 'b2' | 'r2'

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

/** GET /providers — menandai provider mana yang kredensialnya siap dipakai. */
export interface ProviderInfo {
  provider: StorageProvider
  kind: 'oauth' | 'keys'
  configured: boolean
  fields?: string[]
  required?: string[]
}

export interface ConnectAccountResponse {
  account_id: string
  status: AccountStatus
  account: Account
}

export interface SyncResult {
  synced: boolean
  files_indexed: number
  total_bytes: number
  used_bytes: number
  free_bytes: number
}

/** Folder virtual — hidup di DB platform, tak ada di provider (docs/09). */
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

export interface FileListResult {
  path: string
  items: FileEntry[]
  page: number
  total: number
}

export interface UploadResponse {
  id: string
  name: string
  account_id: string
  account_label: string
  routed_by: string
  job_id: string
  file: FileEntry
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

export interface QuotaReport {
  aggregate: Quota
  accounts: AccountQuota[]
}

/** Job upload di UI. Progres nyata datang dari SSE /events/uploads/{jobId}. */
export interface UploadJob {
  id: string
  file_name: string
  size_bytes: number
  bytes_uploaded: number
  progress: number
  status: 'routing' | 'uploading' | 'completed' | 'error' | 'cancelled'
  target_account_id?: string
  target_account_label?: string
  target_provider?: StorageProvider
  target_folder_id?: string | null
  target_folder_path?: string
  routing_strategy?: string
  error_message?: string
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

export interface ApiErrorBody {
  error: {
    code: ApiErrorCode
    message: string
  }
}
