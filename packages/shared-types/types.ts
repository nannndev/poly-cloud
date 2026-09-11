export type AccountStatus = 'active' | 'needs_reconnect' | 'error'

export interface Account {
  id: string
  provider: string
  label: string
  status: AccountStatus
  total_bytes: number | null
  used_bytes: number | null
  free_bytes: number | null
  last_synced: string | null
}

export interface FileEntry {
  id: string
  name: string
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
