export type AccountStatus = 'active' | 'needs_reconnect' | 'error' | 'syncing'

export type StorageProvider = 'gdrive' | 'onedrive' | 'dropbox' | 's3' | 'r2' | 'b2'

export interface Account {
  id: string
  provider: StorageProvider
  label: string
  rclone_remote?: string
  provisioning_type?: 'oauth' | 'credentials'
  status: AccountStatus
  total_bytes: number
  used_bytes: number
  free_bytes: number
  last_synced: string | null
  email?: string
}

export interface FolderEntry {
  id: string
  parent_id: string | null
  name: string
  path: string
  created_at: string
  item_count?: number
}

export interface FileEntry {
  id: string
  name: string
  path: string
  folder_id?: string | null
  virtual_path?: string
  mime: string | null
  size_bytes: number
  modified_at: string
  account_id: string
  account_label: string
  provider: StorageProvider
  is_chunked: boolean
  is_directory?: boolean
  download_url?: string
}

export interface Quota {
  total_bytes: number
  used_bytes: number
  free_bytes: number
}

export interface AggregateQuota {
  aggregate: Quota
  accounts: {
    account_id: string
    label: string
    provider: StorageProvider
    total_bytes: number
    used_bytes: number
    free_bytes: number
  }[]
}

export interface UploadJob {
  id: string
  file_name: string
  size_bytes: number
  bytes_uploaded: number
  progress: number
  status: 'routing' | 'uploading' | 'paused' | 'completed' | 'error'
  target_account_id?: string
  target_account_label?: string
  target_provider?: StorageProvider
  target_folder_id?: string | null
  target_folder_path?: string
  routing_strategy?: string
  speed_mbps?: number
  error_message?: string
}

