export type AccountStatus = 'active' | 'needs_reconnect' | 'error'

export interface Account {
  id: string
  provider: string
  label: string
  rclone_remote?: string
  status: AccountStatus
  total_bytes: number | null
  used_bytes: number | null
  free_bytes: number | null
  last_synced: string | null
}

export interface FolderEntry {
  id: string
  name: string
  path: string
  parent_id: string | null
  created_at?: string
}

export interface FileEntry {
  id: string
  name: string
  folder_id?: string | null
  virtual_path?: string
  mime: string | null
  size_bytes: number
  modified_at: string | null
  account_id: string
  account_label: string
  is_chunked: boolean
}

export interface Quota {
  total_bytes: number
  used_bytes: number
  free_bytes: number
}
