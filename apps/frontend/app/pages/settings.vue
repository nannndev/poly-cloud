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
const healthUrl = computed(() => apiBase.value.replace(/\/api\/v1$/, '') + '/healthz')

// Konfigurasi ini berasal dari environment backend saat proses start; mengubahnya
// butuh menyunting .env lalu me-restart, jadi halaman ini menampilkan, bukan menyimpan.
const settings = ref<BackendSettings | null>(null)

type HealthState = 'checking' | 'online' | 'degraded' | 'offline'
const backendStatus = ref<HealthState>('checking')
const backendLatency = ref<number | null>(null)
const healthDetail = ref<Record<string, string>>({})
const isTesting = ref(false)

const isReindexing = ref(false)

async function testBackend() {
  isTesting.value = true
  backendStatus.value = 'checking'
  const start = performance.now()
  try {
    const res = await $fetch<Record<string, string>>(healthUrl.value, { timeout: 5000 })
    backendLatency.value = Math.round(performance.now() - start)
    healthDetail.value = res
    backendStatus.value = res.status === 'ok' ? 'online' : 'degraded'
  } catch (err: any) {
    // Backend menjawab 503 saat DB tak terjangkau — itu tetap informasi berguna.
    const body = err?.data as Record<string, string> | undefined
    backendLatency.value = Math.round(performance.now() - start)
    if (body?.status) {
      healthDetail.value = body
      backendStatus.value = 'degraded'
    } else {
      healthDetail.value = {}
      backendStatus.value = 'offline'
    }
  } finally {
    isTesting.value = false
  }
}

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
    title: failed > 0 ? `${failed} akun gagal disinkronkan` : 'Index diperbarui',
    description: `${indexed} file terindeks dari ${accountsStore.accounts.length} akun.`,
    color: failed > 0 ? 'warning' : 'success'
  })
}

const strategyMeta: Record<string, { label: string; desc: string }> = {
  'most-free': {
    label: 'Most Free Space',
    desc: 'File diarahkan ke akun dengan sisa ruang terbanyak, sehingga pemakaian tersebar merata.'
  },
  'round-robin': {
    label: 'Round Robin',
    desc: 'Akun tujuan dipilih bergantian secara berurutan tanpa melihat sisa ruang.'
  }
}

const activeStrategy = computed(() => {
  const key = settings.value?.routing_strategy || 'most-free'
  return { key, ...(strategyMeta[key] || { label: key, desc: 'Strategi kustom.' }) }
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
                class="rounded-lg font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
              />
            </div>
          </template>

          <template #right>
            <UButton
              label="Cek Ulang"
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
                      'border-emerald-500/30 bg-emerald-500/10 text-emerald-500': backendStatus === 'online',
                      'border-amber-500/30 bg-amber-500/10 text-amber-500': backendStatus === 'degraded',
                      'border-rose-500/30 bg-rose-500/10 text-rose-500': backendStatus === 'offline',
                      'border-default/70 bg-elevated/40 text-muted': backendStatus === 'checking'
                    }"
                  >
                    <span
                      class="size-2 rounded-full"
                      :class="{
                        'bg-emerald-500 animate-pulse': backendStatus === 'online',
                        'bg-amber-500': backendStatus === 'degraded',
                        'bg-rose-500': backendStatus === 'offline',
                        'bg-zinc-500 animate-pulse': backendStatus === 'checking'
                      }"
                    />
                    <span>
                      {{ backendStatus === 'online' ? 'Terhubung'
                        : backendStatus === 'degraded' ? 'Sebagian bermasalah'
                        : backendStatus === 'offline' ? 'Tidak terjangkau'
                        : 'Memeriksa...' }}
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
                    Diatur lewat <code class="font-mono">NUXT_PUBLIC_API_BASE</code> saat build.
                  </span>
                </div>

                <div class="space-y-1.5">
                  <label class="block text-xs font-bold text-highlighted">Latensi</label>
                  <div class="h-10 px-3.5 rounded-xl border border-default/70 bg-elevated/40 flex items-center justify-between text-xs">
                    <span class="text-muted flex items-center gap-1.5 font-medium">
                      <UIcon name="i-lucide-activity" class="size-3.5 text-sky-500" />
                      /healthz
                    </span>
                    <span class="font-mono font-bold text-sm" :class="backendStatus === 'offline' ? 'text-rose-500' : 'text-emerald-500'">
                      {{ backendLatency === null ? '—' : `${backendLatency} ms` }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Rincian kesehatan per dependensi -->
              <div v-if="Object.keys(healthDetail).length > 0" class="flex flex-wrap gap-2 pt-1">
                <span
                  v-for="(value, key) in healthDetail"
                  :key="key"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg border text-[11px] font-mono"
                  :class="value === 'ok'
                    ? 'border-emerald-500/25 bg-emerald-500/[0.07] text-emerald-400'
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
                  label="Sinkronkan Ulang Semua Akun"
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
                  Membaca ulang kuota dan isi tiap akun dari provider, lalu memperbarui
                  index di database. Organisasi folder virtual tidak terpengaruh.
                </p>
              </div>
            </div>
          </div>

          <!-- Konfigurasi runtime backend: ditampilkan, bukan disunting -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-5">
            <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-indigo-500/10 text-indigo-500 border border-indigo-500/20">
                  <UIcon name="i-lucide-route" class="size-6" />
                </div>
                <div>
                  <h3 class="font-bold text-sm text-highlighted">Konfigurasi Engine</h3>
                  <p class="text-xs text-muted mt-0.5 leading-relaxed max-w-xl">
                    Nilai berikut dibaca backend dari environment saat proses start.
                    Mengubahnya berarti menyunting <code class="font-mono">.env</code> lalu
                    me-restart layanan — bukan lewat halaman ini.
                  </p>
                </div>
              </div>
            </div>

            <div v-if="settings" class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Strategi routing aktif -->
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

              <!-- Model penyimpanan -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Model Penyimpanan</span>
                  <UBadge :label="`Model ${settings.storage_model}`" color="primary" variant="subtle" size="xs" class="rounded-lg font-mono text-[10px]" />
                </div>
                <p class="text-[11px] text-muted leading-relaxed">
                  Satu file disimpan utuh di satu akun (whole-file).
                  <template v-if="!settings.chunking">
                    Pemecahan file lintas akun (Model B) belum tersedia, jadi file yang lebih
                    besar dari sisa ruang akun terlowong akan ditolak.
                  </template>
                </p>
              </div>

              <!-- Folder fisik -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Folder Objek</span>
                  <code class="text-[10px] font-mono text-muted">RCLONE_BASE_DIR</code>
                </div>
                <code class="block font-mono text-sm text-emerald-400">{{ settings.remote_base_dir }}/</code>
                <p class="text-[11px] text-muted leading-relaxed">
                  Semua objek ditaruh rata di folder ini pada tiap akun. Struktur folder
                  yang Anda lihat hidup di database, bukan di provider.
                </p>
              </div>

              <!-- Transfer -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <span class="text-[11px] font-bold uppercase tracking-wider text-muted block">Transfer</span>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-shield-check" class="size-4 text-emerald-500 shrink-0" />
                  <span class="font-bold text-sm text-highlighted">Stream-through</span>
                </div>
                <p class="text-[11px] text-muted leading-relaxed">
                  Unggahan dan unduhan dialirkan langsung antara browser dan provider;
                  server tak menyimpan salinan file.
                </p>
              </div>
            </div>

            <div v-else class="p-5 rounded-2xl border border-amber-500/25 bg-amber-500/[0.06] text-xs text-muted">
              Konfigurasi backend tidak bisa dibaca. Periksa koneksi ke API.
            </div>
          </div>

          <!-- Ringkasan kapasitas -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm">
            <div class="flex flex-wrap items-center justify-between gap-4">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                  <UIcon name="i-lucide-database" class="size-6" />
                </div>
                <div>
                  <h3 class="font-bold text-sm text-highlighted">Kapasitas Agregat</h3>
                  <p class="text-xs text-muted mt-0.5">
                    {{ accountsStore.accounts.length }} akun terhubung,
                    {{ accountsStore.activeAccounts.length }} aktif.
                  </p>
                </div>
              </div>

              <div class="flex items-center gap-6 font-mono text-xs">
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Terpakai</span>
                  <span class="text-highlighted font-bold text-sm">{{ formatBytes(accountsStore.usedStorage) }}</span>
                </div>
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Bebas</span>
                  <span class="text-emerald-500 font-bold text-sm">{{ formatBytes(accountsStore.freeStorage) }}</span>
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
