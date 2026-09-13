<script setup lang="ts">
import type { FileEntry } from '~/types'

const accountsStore = useAccountsStore()
const filesStore = useFilesStore()
const { formatBytes, getProviderMeta } = useFormatters()

// Statistik harus mencakup seluruh file, bukan hanya folder yang sedang dibuka.
const allFiles = ref<FileEntry[]>([])

const { refresh: reloadQuota } = await useAsyncData('quota-page', async () => {
  const [, files] = await Promise.all([
    accountsStore.loadAll().catch(() => null),
    filesStore.fetchAllFiles().catch(() => [] as FileEntry[])
  ])
  allFiles.value = files || []
  return true
}, { server: false, default: () => false })

// Halaman ini seluruhnya angka; tanpa pembeda, backend mati tampak seperti
// akun yang memang kosong.
const isEmpty = computed(() => accountsStore.accounts.length === 0)

const categoryStats = computed(() => {
  let docsBytes = 0
  let mediaBytes = 0
  let archivesBytes = 0
  let codeBytes = 0

  allFiles.value.forEach(f => {
    const mime = (f.mime || '').toLowerCase()
    const name = f.name.toLowerCase()
    if (mime.includes('image') || mime.includes('video') || mime.includes('audio') || name.endsWith('.mp4') || name.endsWith('.png')) {
      mediaBytes += f.size_bytes
    } else if (mime.includes('zip') || mime.includes('tar') || mime.includes('gz') || name.endsWith('.zip') || name.endsWith('.gz')) {
      archivesBytes += f.size_bytes
    } else if (mime.includes('pdf') || mime.includes('sheet') || mime.includes('word') || mime.includes('doc')) {
      docsBytes += f.size_bytes
    } else {
      codeBytes += f.size_bytes
    }
  })

  const total = docsBytes + mediaBytes + archivesBytes + codeBytes || 1
  return [
    { label: 'Media & 4K Video', bytes: mediaBytes, percent: Math.round((mediaBytes / total) * 100), color: 'bg-rose-500', icon: 'i-lucide-film' },
    { label: 'Databases & Archives', bytes: archivesBytes, percent: Math.round((archivesBytes / total) * 100), color: 'bg-amber-500', icon: 'i-lucide-archive' },
    { label: 'Documents & Spreadsheets', bytes: docsBytes, percent: Math.round((docsBytes / total) * 100), color: 'bg-primary-500', icon: 'i-lucide-file-text' },
    { label: 'Source Code & Configs', bytes: codeBytes, percent: Math.round((codeBytes / total) * 100), color: 'bg-sky-500', icon: 'i-lucide-code' }
  ]
})
</script>

<template>
  <div class="flex flex-1 min-w-0 min-h-0 w-full">
    <UDashboardPanel id="quota-panel">
      <template #header>
        <UDashboardNavbar title="Storage & Quota Analytics" :ui="{ right: 'gap-2.5' }">
          <template #leading>
            <UDashboardSidebarCollapse />
          </template>

          <template #title>
            <div class="flex items-center gap-2">
              <h1 class="font-bold text-base text-zinc-100">Storage & Quota Analytics</h1>
              <UBadge
                label="Realtime Aggregate"
                color="primary"
                variant="subtle"
                size="xs"
                class="bg-primary-500/10 text-primary-400 border border-primary-500/20 font-medium"
              />
            </div>
          </template>
        </UDashboardNavbar>
      </template>

      <template #body>
        <StateNotice
          v-if="accountsStore.isLoading && isEmpty"
          variant="loading"
          title="Loading storage report..."
        />

        <StateNotice
          v-else-if="accountsStore.loadError && isEmpty"
          variant="error"
          title="Could not load the storage report"
          :description="accountsStore.loadError"
          @retry="reloadQuota"
        />

        <StateNotice
          v-else-if="isEmpty"
          title="No accounts connected yet"
          description="Connect a cloud account to see how your storage is distributed."
          icon="i-lucide-cloud-off"
        >
          <template #actions>
            <UButton
              label="Connect an account"
              icon="i-lucide-plus"
              color="primary"
              size="xs"
              class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white"
              to="/accounts"
            />
          </template>
        </StateNotice>

        <div v-else class="space-y-6 p-1">
          <!-- Big Total Aggregate Banner -->
          <div class="p-6 rounded-3xl border border-white/[0.08] bg-[#141925] shadow-xs space-y-4">
            <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div>
                <span class="text-xs font-semibold text-zinc-400 uppercase tracking-wider">Unified Total Capacity</span>
                <div class="flex items-baseline gap-2 mt-1">
                  <h2 class="text-3xl sm:text-4xl font-bold text-white font-mono">
                    {{ formatBytes(accountsStore.usedStorage) }}
                  </h2>
                  <span class="text-base text-zinc-400 font-mono">
                    used of {{ formatBytes(accountsStore.totalStorage) }} total
                  </span>
                </div>
              </div>

              <div class="flex items-center gap-3">
                <div class="text-right">
                  <span class="text-[11px] text-zinc-500 block">Available Free Space</span>
                  <span class="text-lg font-bold font-mono text-primary-400">
                    {{ formatBytes(accountsStore.freeStorage) }}
                  </span>
                </div>
                <div class="p-3 rounded-2xl bg-primary-500/10 text-primary-400 border border-primary-500/20 font-bold font-mono text-lg">
                  {{ accountsStore.usagePercent }}%
                </div>
              </div>
            </div>

            <!-- Segmented Multi-Provider Bar -->
            <div class="space-y-2">
              <div class="h-4 w-full rounded-full bg-elevated/80 flex overflow-hidden p-0.5 gap-0.5">
                <div
                  v-for="acc in accountsStore.accounts"
                  :key="acc.id"
                  class="h-full rounded-full transition-all duration-300 relative group cursor-pointer"
                  :style="{
                    width: `${Math.max(2, Math.round(((acc.used_bytes || 0) / (accountsStore.totalStorage || 1)) * 100))}%`,
                    backgroundColor: acc.provider === 'gdrive' ? '#10b981' : acc.provider === 'onedrive' ? '#0ea5e9' : acc.provider === 's3' ? '#f59e0b' : acc.provider === 'r2' ? '#f97316' : '#3b82f6'
                  }"
                  :title="`${acc.label}: ${formatBytes(acc.used_bytes || 0)}`"
                />
              </div>

              <div class="flex flex-wrap items-center justify-between text-xs text-muted pt-1 gap-2">
                <div class="flex flex-wrap items-center gap-4">
                  <div
                    v-for="acc in accountsStore.accounts"
                    :key="acc.id"
                    class="flex items-center gap-1.5"
                  >
                    <span
                      class="size-2.5 rounded-full"
                      :style="{
                        backgroundColor: acc.provider === 'gdrive' ? '#10b981' : acc.provider === 'onedrive' ? '#0ea5e9' : acc.provider === 's3' ? '#f59e0b' : acc.provider === 'r2' ? '#f97316' : '#3b82f6'
                      }"
                    />
                    <span class="text-highlighted font-medium">{{ acc.label }}</span>
                    <span class="font-mono text-muted">({{ formatBytes(acc.used_bytes || 0) }})</span>
                  </div>
                </div>

                <span class="font-mono text-primary-500 font-semibold">
                  {{ formatBytes(accountsStore.freeStorage) }} Free
                </span>
              </div>
            </div>
          </div>

          <!-- Distribution Breakdown Grid -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
            <!-- Breakdown by Cloud Provider -->
            <div class="p-5 rounded-2xl border border-white/[0.08] bg-[#141925] shadow-xs space-y-4">
              <div class="flex items-center justify-between">
                <h3 class="font-semibold text-sm text-zinc-100 flex items-center gap-2">
                  <UIcon name="i-lucide-cloud" class="size-4 text-primary-400" />
                  Allocation per Cloud Provider
                </h3>
                <span class="text-xs text-zinc-500">{{ accountsStore.accounts.length }} Providers</span>
              </div>

              <div class="space-y-3">
                <div
                  v-for="acc in accountsStore.accounts"
                  :key="acc.id"
                  class="p-3 rounded-xl border border-white/[0.06] bg-zinc-900/60 space-y-1.5"
                >
                  <div class="flex items-center justify-between text-xs">
                    <div class="flex items-center gap-2">
                      <UIcon :name="getProviderMeta(acc.provider).icon" class="size-4" />
                      <span class="font-medium text-zinc-200">{{ acc.label }}</span>
                    </div>
                    <span class="font-mono font-medium text-zinc-300">
                      {{ formatBytes(acc.used_bytes || 0) }} / {{ formatBytes(acc.total_bytes) }}
                    </span>
                  </div>

                  <UProgress
                    :model-value="Math.round(((acc.used_bytes || 0) / (acc.total_bytes || 1)) * 100)"
                    color="primary"
                    size="xs"
                  />

                  <div class="flex items-center justify-between text-[10px] text-zinc-500">
                    <span>Remaining capacity: {{ formatBytes(acc.free_bytes) }}</span>
                    <span>{{ Math.round(((acc.used_bytes || 0) / (acc.total_bytes || 1)) * 100) }}% used</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Breakdown by File Category -->
            <div class="p-5 rounded-2xl border border-white/[0.08] bg-[#141925] shadow-xs space-y-4">
              <div class="flex items-center justify-between">
                <h3 class="font-semibold text-sm text-zinc-100 flex items-center gap-2">
                  <UIcon name="i-lucide-pie-chart" class="size-4 text-primary-400" />
                  Indexed File Categories
                </h3>
                <span class="text-xs text-zinc-500">{{ allFiles.length }} Total Files</span>
              </div>

              <div class="space-y-3">
                <div
                  v-for="cat in categoryStats"
                  :key="cat.label"
                  class="p-3 rounded-xl border border-white/[0.06] bg-zinc-900/60 space-y-1.5"
                >
                  <div class="flex items-center justify-between text-xs">
                    <div class="flex items-center gap-2">
                      <UIcon :name="cat.icon" class="size-4 text-primary-400" />
                      <span class="font-medium text-zinc-200">{{ cat.label }}</span>
                    </div>
                    <span class="font-mono font-medium text-zinc-300">
                      {{ formatBytes(cat.bytes) }}
                    </span>
                  </div>

                  <div class="h-1.5 w-full rounded-full bg-zinc-900 overflow-hidden">
                    <div class="h-full rounded-full bg-primary-500" :style="{ width: `${cat.percent}%` }" />
                  </div>

                  <div class="flex items-center justify-between text-[10px] text-zinc-500">
                    <span>Storage share</span>
                    <span>{{ cat.percent }}% of total</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </UDashboardPanel>
  </div>
</template>
