import { defineStore } from 'pinia'
import type {
  Account,
  ConnectAccountResponse,
  ProviderInfo,
  QuotaReport,
  StorageProvider,
  SyncResult
} from '~/types'
import { ApiError } from '~/composables/useApi'

export const useAccountsStore = defineStore('accounts', () => {
  const api = useApi()

  const accounts = ref<Account[]>([])
  const providers = ref<ProviderInfo[]>([])
  const quota = ref<QuotaReport | null>(null)

  const isLoading = ref(false)
  const isSyncing = ref<Record<string, boolean>>({})
  const loadError = ref<string | null>(null)

  // Kuota agregat dari backend bila sudah dimuat; kalau belum, dihitung dari
  // daftar akun supaya angka di UI tak kosong saat halaman baru terbuka.
  const totalStorage = computed(() =>
    quota.value?.aggregate.total_bytes ?? accounts.value.reduce((s, a) => s + a.total_bytes, 0))

  const usedStorage = computed(() =>
    quota.value?.aggregate.used_bytes ?? accounts.value.reduce((s, a) => s + a.used_bytes, 0))

  const freeStorage = computed(() =>
    quota.value?.aggregate.free_bytes ?? Math.max(0, totalStorage.value - usedStorage.value))

  const usagePercent = computed(() => {
    if (totalStorage.value === 0) return 0
    return Math.round((usedStorage.value / totalStorage.value) * 100)
  })

  const activeAccounts = computed(() => accounts.value.filter(a => a.status === 'active'))

  /** Akun yang dipilih router untuk upload berikutnya (most-free, ADR-008). */
  const recommendedAccount = computed(() => {
    const eligible = activeAccounts.value.filter(a => a.free_bytes > 0 || a.total_bytes === 0)
    if (eligible.length === 0) return null
    return [...eligible].sort((a, b) => b.free_bytes - a.free_bytes)[0]
  })

  function getAccount(id: string) {
    return accounts.value.find(a => a.id === id) || null
  }

  async function fetchAccounts() {
    isLoading.value = true
    loadError.value = null
    try {
      accounts.value = await api.get<Account[]>('/accounts')
    } catch (err) {
      loadError.value = friendlyMessage(err)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function fetchQuota() {
    quota.value = await api.get<QuotaReport>('/quota')
  }

  async function fetchProviders() {
    providers.value = await api.get<ProviderInfo[]>('/providers')
  }

  /** Muat semua data akun sekaligus; dipakai saat halaman dibuka. */
  async function loadAll() {
    isLoading.value = true
    loadError.value = null
    try {
      const [accs, q, provs] = await Promise.all([
        api.get<Account[]>('/accounts'),
        api.get<QuotaReport>('/quota'),
        api.get<ProviderInfo[]>('/providers')
      ])
      accounts.value = accs
      quota.value = q
      providers.value = provs
    } catch (err) {
      loadError.value = friendlyMessage(err)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Mulai OAuth: backend membangun consent URL memakai OAuth app platform.
   * Yang dikembalikan adalah URL tujuan redirect (docs/10 §3).
   */
  async function startOAuthConnect(provider: StorageProvider, label: string) {
    const res = await api.post<{ auth_url: string }>('/accounts/connect', { provider, label })
    return res.auth_url
  }

  /** Provider berbasis key (S3/B2/R2): akun langsung jadi tanpa redirect. */
  async function connectWithKeys(provider: StorageProvider, label: string, fields: Record<string, string>) {
    const res = await api.post<ConnectAccountResponse>('/accounts/connect', { provider, label, fields })
    accounts.value = [res.account, ...accounts.value]
    return res.account
  }

  /** Tukar authorization code jadi akun aktif setelah user kembali dari provider. */
  async function completeOAuthConnect(code: string, state: string) {
    const res = await api.post<ConnectAccountResponse>('/accounts/callback', { code, state })
    accounts.value = [res.account, ...accounts.value]
    return res.account
  }

  async function syncAccount(id: string): Promise<SyncResult> {
    isSyncing.value = { ...isSyncing.value, [id]: true }
    try {
      const res = await api.post<SyncResult>(`/accounts/${id}/sync`)
      await Promise.all([fetchAccounts(), fetchQuota()])
      return res
    } finally {
      isSyncing.value = { ...isSyncing.value, [id]: false }
    }
  }

  async function removeAccount(id: string) {
    await api.del(`/accounts/${id}`)
    accounts.value = accounts.value.filter(a => a.id !== id)
    await fetchQuota().catch(() => null)
  }

  /**
   * Akun yang tokennya tak lagi berlaku harus lewat consent ulang — tak ada
   * jalan pintas dari sisi UI. Kembalikan URL consent baru.
   */
  async function reconnectAccount(id: string) {
    const account = getAccount(id)
    if (!account) {
      throw new ApiError('NOT_FOUND', 'Account not found', 404)
    }
    return startOAuthConnect(account.provider, account.label)
  }

  function providerInfo(provider: StorageProvider) {
    return providers.value.find(p => p.provider === provider) || null
  }

  return {
    accounts,
    providers,
    quota,
    isLoading,
    isSyncing,
    loadError,
    totalStorage,
    usedStorage,
    freeStorage,
    usagePercent,
    activeAccounts,
    recommendedAccount,
    getAccount,
    providerInfo,
    loadAll,
    fetchAccounts,
    fetchQuota,
    fetchProviders,
    startOAuthConnect,
    connectWithKeys,
    completeOAuthConnect,
    syncAccount,
    removeAccount,
    reconnectAccount
  }
})
