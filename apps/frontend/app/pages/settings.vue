<script setup lang="ts">
interface BackendSettings {
  routing_strategy: string
  remote_base_dir: string
  storage_model: string
  chunking: boolean
  sync_recurse: boolean
  multi_user: boolean
}

const config = useRuntimeConfig()
const accountsStore = useAccountsStore()
const filesStore = useFilesStore()
const api = useApi()
const toast = useToast()
const { formatBytes } = useFormatters()

const apiBase = computed(() => config.public.apiBase)

// Konfigurasi ini berasal dari environment backend saat proses start; mengubahnya
// butuh menyunting .env lalu me-restart, jadi halaman ini menampilkan, bukan menyimpan.
const settings = ref<BackendSettings | null>(null)

// Status kesehatan dibagi dengan indikator di sidebar — satu sumber, supaya
// keduanya tak pernah melaporkan hal yang berbeda.
const {
  status: backendStatus,
  latency: backendLatency,
  detail: healthDetail,
  isChecking: isTesting,
  check: testBackend
} = useBackendHealth()

const isReindexing = ref(false)

/** Sinkronkan ulang seluruh akun: baca kuota & isi terbaru dari tiap provider. */
async function triggerReindex() {
  isReindexing.value = true
  let failed = 0
  let indexed = 0
  for (const acc of accountsStore.accounts) {
    try {
      const res = await accountsStore.syncAccount(acc.id)
      indexed += res.files_indexed
    } catch {
      failed++
    }
  }
  await filesStore.loadAll().catch(() => null)
  isReindexing.value = false

  toast.add({
    title: failed > 0 ? `${failed} account${failed > 1 ? 's' : ''} failed to sync` : 'Index refreshed',
    description: `${indexed} files indexed across ${accountsStore.accounts.length} accounts.`,
    color: failed > 0 ? 'warning' : 'success'
  })
}

const strategyMeta: Record<string, { label: string; desc: string }> = {
  'most-free': {
    label: 'Most Free Space',
    desc: 'Files go to the account with the most free space, spreading usage evenly.'
  },
  'round-robin': {
    label: 'Round Robin',
    desc: 'Destination accounts are picked in turn, without regard to free space.'
  }
}

const activeStrategy = computed(() => {
  const key = settings.value?.routing_strategy || 'most-free'
  return { key, ...(strategyMeta[key] || { label: key, desc: 'Custom strategy.' }) }
})

await useAsyncData('settings-page', async () => {
  await Promise.all([
    accountsStore.loadAll().catch(() => null),
    api.get<BackendSettings>('/settings').then(res => { settings.value = res }).catch(() => null),
    testBackend()
  ])
  return true
}, { server: false, default: () => false })
</script>

<template>
  <div class="flex flex-1 min-w-0 min-h-0 w-full">
    <UDashboardPanel id="settings-panel">
      <template #header>
        <UDashboardNavbar title="Smart Routing & Engine Settings" :ui="{ right: 'gap-2.5' }">
          <template #leading>
            <UDashboardSidebarCollapse />
          </template>

          <template #title>
            <div class="flex items-center gap-2.5">
              <h1 class="font-bold text-base text-highlighted">
                Smart Routing & Engine Settings
              </h1>
              <UBadge
                label="System Architecture"
                color="primary"
                variant="subtle"
                size="xs"
                class="rounded-lg font-medium bg-primary-500/10 text-primary-400 border border-primary-500/20"
              />
            </div>
          </template>

          <template #right>
            <UButton
              label="Re-check"
              icon="i-lucide-refresh-cw"
              color="neutral"
              variant="outline"
              class="rounded-xl font-semibold shadow-xs px-4 transition-all cursor-pointer"
              :loading="isTesting"
              @click="testBackend"
            />
          </template>
        </UDashboardNavbar>
      </template>

      <template #body>
        <!-- Centered responsive container that gracefully fills wide monitors -->
        <div class="w-full max-w-6xl mx-auto space-y-6 py-2 pb-16">
          <!-- Row 1: Backend Connection & Maintenance (2 Columns) -->
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
            <!-- Backend Connection (2 Cols) -->
            <div class="lg:col-span-2 p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-5">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-default/60">
                <div class="flex items-start gap-3.5">
                  <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-sky-500/10 text-sky-500 border border-sky-500/20">
                    <UIcon name="i-lucide-server" class="size-6" />
                  </div>
                  <div>
                    <h3 class="font-bold text-sm text-highlighted">
                      Go Backend API & Rclone Engine Connectivity
                    </h3>
                    <p class="text-xs text-muted mt-0.5 leading-relaxed">
                      REST communication between Nuxt 4 frontend, rclone backend engine, and PostgreSQL database.
                    </p>
                  </div>
                </div>

                <div class="flex items-center gap-2 shrink-0">
                  <div
                    class="flex items-center gap-2 px-3 py-1 rounded-xl border text-xs font-semibold"
                    :class="{
                      'border-primary-500/30 bg-primary-500/10 text-primary-500': backendStatus === 'online',
                      'border-amber-500/30 bg-amber-500/10 text-amber-500': backendStatus === 'degraded',
                      'border-rose-500/30 bg-rose-500/10 text-rose-500': backendStatus === 'offline',
                      'border-default/70 bg-elevated/40 text-muted': backendStatus === 'checking'
                    }"
                  >
                    <span
                      class="size-2 rounded-full"
                      :class="{
                        'bg-primary-500 animate-pulse': backendStatus === 'online',
                        'bg-amber-500': backendStatus === 'degraded',
                        'bg-rose-500': backendStatus === 'offline',
                        'bg-zinc-500 animate-pulse': backendStatus === 'checking'
                      }"
                    />
                    <span>
                      {{ backendStatus === 'online' ? 'Connected'
                        : backendStatus === 'degraded' ? 'Partially degraded'
                        : backendStatus === 'offline' ? 'Unreachable'
                        : 'Checking...' }}
                    </span>
                  </div>

                  <UButton
                    icon="i-lucide-refresh-cw"
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    square
                    class="rounded-xl"
                    :loading="isTesting"
                    @click="testBackend"
                  />
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="sm:col-span-2 space-y-1.5">
                  <label class="block text-xs font-bold text-highlighted">API Base Endpoint</label>
                  <div class="h-10 px-3.5 rounded-xl border border-default/70 bg-elevated/40 flex items-center text-xs font-mono text-muted truncate">
                    {{ apiBase }}
                  </div>
                  <span class="text-[10px] text-muted block">
                    Set via <code class="font-mono">NUXT_PUBLIC_API_BASE</code> at build time.
                  </span>
                </div>

                <div class="space-y-1.5">
                  <label class="block text-xs font-bold text-highlighted">Latency</label>
                  <div class="h-10 px-3.5 rounded-xl border border-default/70 bg-elevated/40 flex items-center justify-between text-xs">
                    <span class="text-muted flex items-center gap-1.5 font-medium">
                      <UIcon name="i-lucide-activity" class="size-3.5 text-sky-500" />
                      /healthz
                    </span>
                    <span class="font-mono font-bold text-sm" :class="backendStatus === 'offline' ? 'text-rose-500' : 'text-primary-500'">
                      {{ backendLatency === null ? '—' : `${backendLatency} ms` }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Per-dependency health detail -->
              <div v-if="Object.keys(healthDetail).length > 0" class="flex flex-wrap gap-2 pt-1">
                <span
                  v-for="(value, key) in healthDetail"
                  :key="key"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg border text-[11px] font-mono"
                  :class="value === 'ok'
                    ? 'border-primary-500/25 bg-primary-500/[0.07] text-primary-400'
                    : 'border-amber-500/25 bg-amber-500/[0.07] text-amber-400'"
                >
                  <UIcon :name="value === 'ok' ? 'i-lucide-check' : 'i-lucide-triangle-alert'" class="size-3" />
                  {{ key }}: {{ value }}
                </span>
              </div>
            </div>

            <!-- Maintenance & Cache Actions (1 Col) -->
            <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-4 flex flex-col justify-between">
              <div>
                <div class="flex items-center gap-2.5 mb-1">
                  <UIcon name="i-lucide-wrench" class="size-5 text-rose-500" />
                  <h3 class="font-bold text-sm text-highlighted">Catalog & Cache</h3>
                </div>
                <p class="text-xs text-muted leading-relaxed">
                  Maintenance operations for multi-cloud metadata catalog and database indices.
                </p>
              </div>

              <div class="space-y-2.5">
                <UButton
                  label="Resync All Accounts"
                  icon="i-lucide-refresh-cw"
                  color="info"
                  variant="subtle"
                  block
                  class="rounded-xl text-xs font-bold justify-start"
                  :loading="isReindexing"
                  :disabled="accountsStore.accounts.length === 0"
                  @click="triggerReindex"
                />
                <p class="text-[10px] text-muted leading-relaxed">
                  Re-reads each account's quota and contents from the provider, then refreshes
                  the database index. Your virtual folder layout is left untouched.
                </p>
              </div>
            </div>
          </div>

          <!-- Backend runtime config: displayed, not editable -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-5">
            <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-indigo-500/10 text-indigo-500 border border-indigo-500/20">
                  <UIcon name="i-lucide-route" class="size-6" />
                </div>
                <div>
                  <h3 class="font-bold text-sm text-highlighted">Engine Configuration</h3>
                  <p class="text-xs text-muted mt-0.5 leading-relaxed max-w-xl">
                    The backend reads these values from the environment at startup. Changing them
                    means editing <code class="font-mono">.env</code> and restarting the service —
                    not from this page.
                  </p>
                </div>
              </div>
            </div>

            <div v-if="settings" class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Active routing strategy -->
              <div class="p-5 rounded-2xl border border-sky-500/30 bg-sky-500/[0.06] space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Smart Routing</span>
                  <code class="text-[10px] font-mono text-muted">ROUTING_STRATEGY</code>
                </div>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-check-circle-2" class="size-4 text-sky-500 shrink-0" />
                  <span class="font-bold text-sm text-highlighted">{{ activeStrategy.label }}</span>
                </div>
                <p class="text-[11px] text-muted leading-relaxed">{{ activeStrategy.desc }}</p>
              </div>

              <!-- Storage model -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Storage Model</span>
                  <UBadge :label="`Model ${settings.storage_model}`" color="primary" variant="subtle" size="xs" class="rounded-lg font-mono text-[10px]" />
                </div>
                <p class="text-[11px] text-muted leading-relaxed">
                  Each file is stored whole on a single account (whole-file).
                  <template v-if="!settings.chunking">
                    Splitting a file across accounts (Model B) is not available yet, so a file
                    larger than the roomiest account's free space is rejected rather than split.
                  </template>
                </p>
              </div>

              <!-- Physical folder -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Object Folder</span>
                  <code class="text-[10px] font-mono text-muted">RCLONE_BASE_DIR</code>
                </div>
                <code class="block font-mono text-sm text-primary-400">{{ settings.remote_base_dir }}/</code>
                <p class="text-[11px] text-muted leading-relaxed">
                  Every object lands flat in this folder on each account. The folder structure
                  you see lives in the database, not at the provider.
                </p>
              </div>

              <!-- Transfer -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <span class="text-[11px] font-bold uppercase tracking-wider text-muted block">Transfer</span>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-shield-check" class="size-4 text-primary-500 shrink-0" />
                  <span class="font-bold text-sm text-highlighted">Stream-through</span>
                </div>
                <p class="text-[11px] text-muted leading-relaxed">
                  Uploads and downloads stream directly between the browser and the provider;
                  the server keeps no copy of the file.
                </p>
              </div>
            </div>

            <div v-else class="p-5 rounded-2xl border border-amber-500/25 bg-amber-500/[0.06] text-xs text-muted">
              Backend configuration could not be read. Check the API connection.
            </div>
          </div>

          <!-- Capacity summary -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm">
            <div class="flex flex-wrap items-center justify-between gap-4">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-primary-500/10 text-primary-500 border border-primary-500/20">
                  <UIcon name="i-lucide-database" class="size-6" />
                </div>
                <div>
                  <h3 class="font-bold text-sm text-highlighted">Aggregate Capacity</h3>
                  <p class="text-xs text-muted mt-0.5">
                    {{ accountsStore.accounts.length }} accounts connected,
                    {{ accountsStore.activeAccounts.length }} active.
                  </p>
                </div>
              </div>

              <div class="flex items-center gap-6 font-mono text-xs">
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Used</span>
                  <span class="text-highlighted font-bold text-sm">{{ formatBytes(accountsStore.usedStorage) }}</span>
                </div>
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Free</span>
                  <span class="text-primary-500 font-bold text-sm">{{ formatBytes(accountsStore.freeStorage) }}</span>
                </div>
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Total</span>
                  <span class="text-highlighted font-bold text-sm">{{ formatBytes(accountsStore.totalStorage) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </UDashboardPanel>
  </div>
</template>
