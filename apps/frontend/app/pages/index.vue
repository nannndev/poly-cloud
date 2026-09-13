<script setup lang="ts">
const accountsStore = useAccountsStore()
const filesStore = useFilesStore()
const { formatBytes } = useFormatters()
const { latency, status, check } = useBackendHealth()

// Explorer butuh keduanya: daftar akun untuk badge & filter, folder + file
// untuk isinya. Dimuat di klien karena backend hanya dijangkau dari browser.
await useAsyncData('explorer-page', async () => {
  await Promise.all([
    accountsStore.loadAll().catch(() => null),
    filesStore.loadAll().catch(() => null),
    check()
  ])
  return true
}, { server: false, default: () => false })

// Banner KPI ikut memberi tahu saat backend tak terjangkau; angka nol di kartu
// tanpa penjelasan terbaca seperti "memang belum ada apa-apa".
const loadFailed = computed(() =>
  Boolean(filesStore.loadError) && filesStore.files.length === 0)

// Tujuan yang akan dipilih router untuk unggahan berikutnya. Ini prediksi dari
// kuota yang terakhir disinkronkan — keputusan sebenarnya tetap di backend saat
// unggahan berjalan (ADR-008).
const routingTarget = computed(() => accountsStore.recommendedAccount)

// Menjatuhkan file di mana pun pada halaman ini akan mengunggahnya ke folder
// yang sedang dibuka — tak perlu membuka modal unggah lebih dulu.
const { uploadAndReport } = useUploadReporter()
const { isDraggingFiles } = useDropZone(files => {
  // Modal dibuka supaya antrean dan progresnya terlihat; unggahannya sendiri
  // berjalan lepas dari modal itu.
  filesStore.isUploadModalOpen = true
  void uploadAndReport(files)
})
</script>

<template>
  <div class="flex flex-1 min-w-0 min-h-0 w-full">
    <UDashboardPanel id="files-panel">
      <template #header>
        <UDashboardNavbar title="Unified File Explorer" :ui="{ right: 'gap-2.5' }">
          <template #leading>
            <UDashboardSidebarCollapse />
          </template>

          <template #title>
            <div class="flex items-center gap-2">
              <h1 class="font-bold text-base text-highlighted">Unified File Explorer</h1>
              <UBadge
                label="Virtual Aggregator"
                color="primary"
                variant="subtle"
                size="xs"
                class="hidden sm:inline-flex rounded-lg font-semibold bg-primary-500/10 text-primary-400 border border-primary-500/20"
              />
            </div>
          </template>

          <template #right>
            <div class="flex items-center gap-2">
              <UButton
                label="New Folder"
                icon="i-lucide-folder-plus"
                color="neutral"
                variant="outline"
                class="rounded-xl font-medium shadow-xs text-zinc-300 hover:text-white border-white/[0.08] hover:bg-white/[0.05] px-3 transition-all cursor-pointer"
                @click="filesStore.isNewFolderModalOpen = true"
              />
              <UButton
                label="Upload File"
                icon="i-lucide-upload-cloud"
                color="primary"
                variant="solid"
                class="rounded-xl font-bold shadow-xs bg-primary-600 hover:bg-primary-500 text-white px-3.5 transition-all cursor-pointer"
                @click="filesStore.isUploadModalOpen = true"
              />
            </div>
          </template>
        </UDashboardNavbar>
      </template>

      <template #body>
        <div class="space-y-6 p-1">
          <div
            v-if="loadFailed"
            class="flex flex-wrap items-center justify-between gap-3 p-4 rounded-2xl border border-rose-500/25 bg-rose-500/[0.06]"
          >
            <span class="flex items-center gap-2.5 text-xs text-rose-200">
              <UIcon name="i-lucide-circle-alert" class="size-4 text-rose-400 shrink-0" />
              <span>{{ filesStore.loadError }} — the figures below may be out of date.</span>
            </span>
            <UButton
              label="Retry"
              icon="i-lucide-refresh-cw"
              color="error"
              variant="subtle"
              size="xs"
              class="rounded-xl font-semibold"
              @click="filesStore.loadAll().catch(() => null)"
            />
          </div>

          <!-- Top KPI Metrics Banner -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <!-- Metric 1: Total Unified Storage -->
            <div class="p-5 rounded-3xl border border-white/[0.08] bg-[#151a27] hover:border-primary-500/30 hover:bg-[#1a1f2d] transition-all duration-200 space-y-3.5 relative overflow-hidden group shadow-xs">
              <div class="flex items-center justify-between">
                <span class="text-[11px] font-bold uppercase tracking-wider text-zinc-400">Unified Capacity</span>
                <div class="p-2 rounded-xl bg-primary-500/10 text-primary-400 border border-primary-500/20 group-hover:scale-105 transition-transform">
                  <UIcon name="i-lucide-hard-drive" class="size-4" />
                </div>
              </div>
              <div>
                <div class="flex items-baseline gap-1.5">
                  <span class="text-2xl font-black text-white font-mono tracking-tight">
                    {{ formatBytes(accountsStore.usedStorage) }}
                  </span>
                  <span class="text-xs text-zinc-400 font-mono">/ {{ formatBytes(accountsStore.totalStorage) }}</span>
                </div>
                <p class="text-[11px] text-zinc-400 mt-0.5">Aggregated multi-cloud storage pool</p>
              </div>
              <div class="space-y-1.5 pt-1">
                <div class="h-1.5 w-full bg-zinc-900 border border-white/[0.06] rounded-full overflow-hidden">
                  <div class="h-full bg-primary-500 rounded-full transition-all duration-300" :style="{ width: `${accountsStore.usagePercent}%` }" />
                </div>
                <div class="flex justify-between text-[10px] text-zinc-400 font-mono font-medium">
                  <span class="text-primary-400 font-semibold">{{ accountsStore.usagePercent }}% used</span>
                  <span class="text-zinc-400 font-medium">{{ formatBytes(accountsStore.freeStorage) }} free</span>
                </div>
              </div>
            </div>

            <!-- Metric 2: Total Items Indexed (VFS) -->
            <div class="p-5 rounded-3xl border border-white/[0.08] bg-[#151a27] hover:border-primary-500/30 hover:bg-[#1a1f2d] transition-all duration-200 space-y-3.5 relative overflow-hidden group shadow-xs">
              <div class="flex items-center justify-between">
                <span class="text-[11px] font-bold uppercase tracking-wider text-zinc-400">VFS Indexed Items</span>
                <div class="p-2 rounded-xl bg-zinc-800/80 text-zinc-300 border border-white/[0.06] group-hover:scale-105 transition-transform">
                  <UIcon name="i-lucide-folder-tree" class="size-4 text-primary-400" />
                </div>
              </div>
              <div>
                <div class="flex items-baseline gap-2">
                  <span class="text-2xl font-black text-white font-mono tracking-tight">
                    {{ filesStore.totalFiles }}
                  </span>
                  <span class="text-xs text-zinc-400 font-mono">files</span>
                  <span class="text-zinc-600">•</span>
                  <span class="text-lg font-bold text-zinc-300 font-mono">
                    {{ filesStore.folders.length }}
                  </span>
                  <span class="text-xs text-zinc-400 font-mono">folders</span>
                </div>
                <p class="text-[11px] text-zinc-400 mt-0.5">Unified virtual hierarchy in platform DB</p>
              </div>
              <!-- Latensi nyata dari /healthz, bukan angka tetap. -->
              <div class="p-2 rounded-xl bg-zinc-900/80 border border-white/[0.06] flex items-center justify-between text-[11px]">
                <span class="text-zinc-400 font-medium">API Latency</span>
                <span
                  class="font-mono font-medium"
                  :class="status === 'offline' ? 'text-rose-400' : 'text-primary-400'"
                >
                  {{ status === 'offline' ? 'unreachable' : latency === null ? '—' : `${latency} ms` }}
                </span>
              </div>
            </div>

            <!-- Metric 3: Connected Cloud Accounts -->
            <div class="p-5 rounded-3xl border border-white/[0.08] bg-[#151a27] hover:border-primary-500/30 hover:bg-[#1a1f2d] transition-all duration-200 space-y-3.5 relative overflow-hidden group shadow-xs">
              <div class="flex items-center justify-between">
                <span class="text-[11px] font-bold uppercase tracking-wider text-zinc-400">Active Providers</span>
                <div class="p-2 rounded-xl bg-zinc-800/80 text-zinc-300 border border-white/[0.06] group-hover:scale-105 transition-transform">
                  <UIcon name="i-lucide-cloud-check" class="size-4" />
                </div>
              </div>
              <div>
                <div class="flex items-baseline gap-2">
                  <span class="text-2xl font-black text-white font-mono tracking-tight">
                    {{ accountsStore.activeAccounts.length }}
                  </span>
                  <span class="text-xs text-zinc-400">of {{ accountsStore.accounts.length }} accounts</span>
                </div>
                <p class="text-[11px] text-zinc-400 mt-0.5">GDrive, OneDrive, Dropbox, S3</p>
              </div>
              <NuxtLink
                to="/accounts"
                class="inline-flex items-center justify-between w-full p-2 rounded-xl bg-zinc-900/80 border border-white/[0.06] text-[11px] font-medium text-zinc-300 hover:text-primary-400 hover:border-primary-500/30 transition-colors"
              >
                <span>Manage Accounts</span>
                <UIcon name="i-lucide-arrow-right" class="size-3.5 text-zinc-400" />
              </NuxtLink>
            </div>

            <!-- Metric 4: Smart Routing Engine -->
            <div class="p-5 rounded-3xl border border-white/[0.08] bg-[#151a27] hover:border-primary-500/30 hover:bg-[#1a1f2d] transition-all duration-200 space-y-3.5 relative overflow-hidden group shadow-xs">
              <div class="flex items-center justify-between">
                <span class="text-[11px] font-bold uppercase tracking-wider text-zinc-400">Smart Routing</span>
                <div class="p-2 rounded-xl bg-zinc-800/80 text-zinc-300 border border-white/[0.06] group-hover:scale-105 transition-transform">
                  <UIcon name="i-lucide-cpu" class="size-4" />
                </div>
              </div>
              <div>
                <div class="flex items-center gap-1.5">
                  <span class="text-sm font-bold text-white">
                    Most-Free Capacity
                  </span>
                </div>
                <p class="text-[11px] text-zinc-400 mt-0.5">Automated storage placement</p>
              </div>
              <!-- Tujuan nyata menurut kuota terakhir, bukan contoh tetap. -->
              <div class="p-2 rounded-xl bg-zinc-900/80 border border-white/[0.06] flex items-center justify-between gap-2 text-[11px]">
                <span class="text-zinc-400 font-medium shrink-0">Next Target</span>
                <span v-if="routingTarget" class="font-medium text-primary-400 truncate">
                  {{ routingTarget.label }} ({{ formatBytes(routingTarget.free_bytes) }})
                </span>
                <span v-else class="font-medium text-amber-400">No active account</span>
              </div>
            </div>
          </div>

          <!-- Main File Explorer View -->
          <FileExplorer />
        </div>
      </template>
    </UDashboardPanel>

    <DropOverlay :show="isDraggingFiles" :destination="filesStore.currentPath" />
  </div>
</template>
