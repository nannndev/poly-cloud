import { defineStore } from 'pinia'
import type { Account, AggregateQuota, StorageProvider } from '~/types'

export const useAccountsStore = defineStore('accounts', () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase || 'http://localhost:8080/api/v1'

  const accounts = ref<Account[]>([
    {
      id: 'acc-1',
      provider: 'gdrive',
      label: 'Google Drive Primary',
      email: 'yubidev.work@gmail.com',
      status: 'active',
      total_bytes: 15 * 1024 * 1024 * 1024, // 15 GB
      used_bytes: 12.4 * 1024 * 1024 * 1024, // 12.4 GB
      free_bytes: 2.6 * 1024 * 1024 * 1024,
      last_synced: new Date(Date.now() - 1000 * 60 * 12).toISOString()
    },
    {
      id: 'acc-2',
      provider: 'onedrive',
      label: 'OneDrive Business',
      email: 'alex@enterprise-cloud.io',
      status: 'active',
      total_bytes: 100 * 1024 * 1024 * 1024, // 100 GB
      used_bytes: 34.2 * 1024 * 1024 * 1024, // 34.2 GB
      free_bytes: 65.8 * 1024 * 1024 * 1024,
      last_synced: new Date(Date.now() - 1000 * 60 * 35).toISOString()
    },
    {
      id: 'acc-3',
      provider: 'dropbox',
      label: 'Dropbox Team Share',
      email: 'admin@polycloud.dev',
      status: 'active',
      total_bytes: 20 * 1024 * 1024 * 1024, // 20 GB
      used_bytes: 16.8 * 1024 * 1024 * 1024,
      free_bytes: 3.2 * 1024 * 1024 * 1024,
      last_synced: new Date(Date.now() - 1000 * 60 * 90).toISOString()
    },
    {
      id: 'acc-4',
      provider: 's3',
      label: 'AWS S3 Cold Bucket (us-east-1)',
      email: 'arn:aws:s3:::poly-archive-vault',
      status: 'active',
      total_bytes: 500 * 1024 * 1024 * 1024, // 500 GB
      used_bytes: 142 * 1024 * 1024 * 1024,
      free_bytes: 358 * 1024 * 1024 * 1024,
      last_synced: new Date(Date.now() - 1000 * 60 * 5).toISOString()
    },
    {
      id: 'acc-5',
      provider: 'r2',
      label: 'Cloudflare R2 Media Hot',
      email: 'assets-bucket@cf-edge.net',
      status: 'needs_reconnect',
      total_bytes: 50 * 1024 * 1024 * 1024, // 50 GB
      used_bytes: 41.5 * 1024 * 1024 * 1024,
      free_bytes: 8.5 * 1024 * 1024 * 1024,
      last_synced: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString()
    }
  ])

  const isLoading = ref(false)
  const isSyncing = ref<Record<string, boolean>>({})

  const totalStorage = computed(() => {
    return accounts.value.reduce((sum, a) => sum + (a.total_bytes || 0), 0)
  })

  const usedStorage = computed(() => {
    return accounts.value.reduce((sum, a) => sum + (a.used_bytes || 0), 0)
  })

  const freeStorage = computed(() => {
    return Math.max(0, totalStorage.value - usedStorage.value)
  })

  const usagePercent = computed(() => {
    if (totalStorage.value === 0) return 0
    return Math.round((usedStorage.value / totalStorage.value) * 100)
  })

  const activeAccounts = computed(() => {
    return accounts.value.filter(a => a.status === 'active')
  })

  async function fetchAccounts() {
    isLoading.value = true
    try {
      const res = await $fetch<Account[]>(`${apiBase}/accounts`, {
        timeout: 2000
      }).catch(() => null)

      if (res && Array.isArray(res) && res.length > 0) {
        accounts.value = res
      }
    } finally {
      isLoading.value = false
    }
  }

  async function syncAccount(id: string) {
    isSyncing.value[id] = true
    try {
      await $fetch(`${apiBase}/accounts/${id}/sync`, { method: 'POST' }).catch(() => null)
      const target = accounts.value.find(a => a.id === id)
      if (target) {
        target.last_synced = new Date().toISOString()
        target.status = 'active'
      }
    } finally {
      setTimeout(() => {
        isSyncing.value[id] = false
      }, 1000)
    }
  }

  function addAccount(data: { provider: StorageProvider; label: string; email?: string; total_bytes?: number }) {
    const newAcc: Account = {
      id: 'acc-' + Math.random().toString(36).substring(2, 9),
      provider: data.provider,
      label: data.label,
      email: data.email || 'connected-user@storage.cloud',
      status: 'active',
      total_bytes: data.total_bytes || 25 * 1024 * 1024 * 1024,
      used_bytes: 0,
      free_bytes: data.total_bytes || 25 * 1024 * 1024 * 1024,
      last_synced: new Date().toISOString()
    }
    accounts.value.unshift(newAcc)
    return newAcc
  }

  function removeAccount(id: string) {
    accounts.value = accounts.value.filter(a => a.id !== id)
  }

  function reconnectAccount(id: string) {
    const target = accounts.value.find(a => a.id === id)
    if (target) {
      target.status = 'active'
      target.last_synced = new Date().toISOString()
    }
  }

  return {
    accounts,
    isLoading,
    isSyncing,
    totalStorage,
    usedStorage,
    freeStorage,
    usagePercent,
    activeAccounts,
    fetchAccounts,
    syncAccount,
    addAccount,
    removeAccount,
    reconnectAccount
  }
})
