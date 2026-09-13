<script setup lang="ts">
const filesStore = useFilesStore()
const accountsStore = useAccountsStore()
const { formatBytes } = useFormatters()
const { uploadAndReport } = useUploadReporter()

const isDragging = ref(false)
const selectedFolderId = ref<string>(filesStore.currentFolderId || '__root__')
const fileInput = ref<HTMLInputElement | null>(null)

// Selaraskan folder tujuan dengan lokasi yang sedang dibuka tiap modal terbuka.
watch(() => filesStore.isUploadModalOpen, (isOpen) => {
  if (isOpen) {
    selectedFolderId.value = filesStore.currentFolderId || '__root__'
  }
})

const folderOptions = computed(() => {
  return filesStore.allFoldersHierarchical.map(f => ({
    value: f.id === null ? '__root__' : f.id,
    label: f.id === null ? '📁 Root (/)' : `${'  '.repeat(f.depth)}📁 ${f.name} (${f.path})`
  }))
})

// Akun yang akan dipilih router untuk unggahan berikutnya. Ini prediksi untuk
// ditampilkan saja — keputusan sebenarnya ada di backend (ADR-008).
const recommendedAccount = computed(() => accountsStore.recommendedAccount)

function getFileBadge(fileName: string) {
  const name = (fileName || '').toLowerCase()
  const parts = name.split('.')
  const rawExt = parts.length > 1 ? parts.pop() || 'FILE' : 'FILE'

  if (name.endsWith('.xlsx') || name.endsWith('.xls') || name.endsWith('.csv')) {
    return {
      ext: 'XLSX',
      color: 'bg-gradient-to-br from-primary-500 to-teal-600 text-white shadow-md shadow-primary-500/30',
      icon: 'i-lucide-file-spreadsheet'
    }
  }
  if (name.endsWith('.pdf')) {
    return {
      ext: 'PDF',
      color: 'bg-gradient-to-br from-rose-500 to-red-600 text-white shadow-md shadow-rose-500/30',
      icon: 'i-lucide-file-text'
    }
  }
  if (name.endsWith('.doc') || name.endsWith('.docx')) {
    return {
      ext: 'DOCX',
      color: 'bg-gradient-to-br from-sky-500 to-blue-600 text-white shadow-md shadow-sky-500/30',
      icon: 'i-lucide-file-type'
    }
  }
  if (name.endsWith('.zip') || name.endsWith('.tar') || name.endsWith('.gz') || name.endsWith('.rar') || name.endsWith('.7z')) {
    return {
      ext: 'ZIP',
      color: 'bg-gradient-to-br from-amber-500 to-orange-600 text-white shadow-md shadow-amber-500/30',
      icon: 'i-lucide-archive'
    }
  }
  if (name.endsWith('.yaml') || name.endsWith('.yml')) {
    return {
      ext: 'YAML',
      color: 'bg-gradient-to-br from-violet-500 to-purple-600 text-white shadow-md shadow-violet-500/30',
      icon: 'i-lucide-file-code'
    }
  }
  if (name.endsWith('.json')) {
    return {
      ext: 'JSON',
      color: 'bg-gradient-to-br from-amber-500 to-yellow-600 text-white shadow-md shadow-amber-500/30',
      icon: 'i-lucide-file-code'
    }
  }
  if (name.endsWith('.sql') || name.endsWith('.db')) {
    return {
      ext: 'SQL',
      color: 'bg-gradient-to-br from-blue-600 to-indigo-700 text-white shadow-md shadow-blue-500/30',
      icon: 'i-lucide-database'
    }
  }
  if (name.endsWith('.mp4') || name.endsWith('.mov') || name.endsWith('.mkv')) {
    return {
      ext: 'MP4',
      color: 'bg-gradient-to-br from-fuchsia-500 to-pink-600 text-white shadow-md shadow-fuchsia-500/30',
      icon: 'i-lucide-film'
    }
  }
  if (name.endsWith('.png') || name.endsWith('.jpg') || name.endsWith('.jpeg') || name.endsWith('.webp') || name.endsWith('.svg')) {
    return {
      ext: 'IMG',
      color: 'bg-gradient-to-br from-purple-500 to-indigo-600 text-white shadow-md shadow-purple-500/30',
      icon: 'i-lucide-image'
    }
  }
  if (name.endsWith('.ts') || name.endsWith('.js') || name.endsWith('.go') || name.endsWith('.py')) {
    return {
      ext: 'CODE',
      color: 'bg-gradient-to-br from-cyan-500 to-teal-600 text-white shadow-md shadow-cyan-500/30',
      icon: 'i-lucide-code'
    }
  }
  return {
    ext: rawExt.slice(0, 4).toUpperCase(),
    color: 'bg-gradient-to-br from-slate-600 to-zinc-700 text-white shadow-md shadow-zinc-700/30',
    icon: 'i-lucide-file'
  }
}

function handleDrop(e: DragEvent) {
  isDragging.value = false
  if (e.dataTransfer?.files) {
    processFiles(Array.from(e.dataTransfer.files))
  }
}

function handleFileInput(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files) {
    processFiles(Array.from(target.files))
  }
  // Kosongkan agar memilih file yang sama dua kali tetap memicu change.
  target.value = ''
}

async function processFiles(fileList: File[]) {
  // Modal punya pemilih folder sendiri, jadi tujuannya dikirim eksplisit —
  // beda dari drop di halaman, yang selalu memakai folder yang sedang dibuka.
  const folderId = selectedFolderId.value === '__root__' ? null : selectedFolderId.value
  await uploadAndReport(fileList, folderId)
}

function triggerFileInput() {
  fileInput.value?.click()
}
</script>

<template>
  <UModal
    v-model:open="filesStore.isUploadModalOpen"
    :ui="{
      content: 'sm:max-w-2xl bg-[#0d111a] dark:bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
      body: 'p-6 space-y-5',
      footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
    }"
  >
    <!-- Modal Header -->
    <template #header>
      <div class="px-6 pt-5 pb-4 bg-[#141925] flex items-start justify-between gap-4 border-b border-white/[0.06]">
        <div class="flex items-start gap-3.5">
          <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-primary-500/15 text-primary-400 border border-primary-500/25">
            <UIcon name="i-lucide-upload-cloud" class="size-5" />
          </div>
          <div>
            <h2 class="text-base font-bold tracking-tight text-white flex items-center gap-2">
              Smart Multi-Cloud File Upload
              <UBadge label="Zero Vendor Lock-in" color="primary" variant="subtle" size="xs" class="font-medium text-[9px] rounded-md bg-primary-500/10 text-primary-400 border border-primary-500/20" />
            </h2>
            <p class="text-xs text-zinc-400 mt-0.5 leading-relaxed max-w-lg">
              Dynamic animated tracker with real-time transfer telemetry, smart routing, and instant file synchronization.
            </p>
          </div>
        </div>

        <button
          type="button"
          class="flex size-8 items-center justify-center rounded-xl bg-zinc-800/60 hover:bg-zinc-700 text-zinc-400 hover:text-white border border-white/[0.06] transition-colors cursor-pointer"
          @click="filesStore.isUploadModalOpen = false"
        >
          <UIcon name="i-lucide-x" class="size-4" />
        </button>
      </div>
    </template>

    <template #body>
      <div class="space-y-5">
        <!-- Dual Target Placement Config: Virtual Folder vs Cloud Provider -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <!-- Placement 1: Virtual Folder Destination -->
          <div class="p-3.5 rounded-2xl bg-[#151a27] border border-white/[0.07] space-y-2">
            <div class="flex items-center justify-between text-xs font-semibold text-zinc-300">
              <span class="flex items-center gap-2">
                <UIcon name="i-lucide-folder" class="size-4 text-amber-400" />
                Virtual Folder (VFS)
              </span>
              <span class="text-[10px] text-zinc-400 font-mono">Organization</span>
            </div>
            <USelect
              v-model="selectedFolderId"
              :items="folderOptions"
              class="w-full rounded-xl bg-[#1c2231] border border-white/[0.08] text-white"
              icon="i-lucide-folder-open"
              size="md"
            />
          </div>

          <!-- Tujuan fisik ditentukan backend, bukan dipilih user (ADR-008) -->
          <div class="p-3.5 rounded-2xl bg-[#151a27] border border-white/[0.07] space-y-2">
            <div class="flex items-center justify-between text-xs font-semibold text-zinc-300">
              <span class="flex items-center gap-2">
                <UIcon name="i-lucide-route" class="size-4 text-primary-400" />
                Cloud Target (Smart Routing)
              </span>
              <span class="text-[10px] text-zinc-400 font-mono">Physical</span>
            </div>

            <div
              v-if="recommendedAccount"
              class="flex items-center gap-2.5 px-3 py-2 rounded-xl bg-[#1c2231] border border-white/[0.08]"
            >
              <UIcon name="i-lucide-zap" class="size-4 text-primary-400 shrink-0" />
              <div class="min-w-0">
                <p class="text-xs font-semibold text-white truncate">{{ recommendedAccount.label }}</p>
                <p class="text-[10px] text-zinc-500">
                  Most free space ({{ formatBytes(recommendedAccount.free_bytes) }}) — the backend
                  decides the final destination as the upload runs.
                </p>
              </div>
            </div>

            <div
              v-else
              class="flex items-center gap-2.5 px-3 py-2 rounded-xl bg-amber-500/[0.06] border border-amber-500/20"
            >
              <UIcon name="i-lucide-triangle-alert" class="size-4 text-amber-400 shrink-0" />
              <p class="text-[11px] text-zinc-300">
                No active account yet. Connect one before uploading.
              </p>
            </div>
          </div>
        </div>

        <!-- Animated Drag & Drop Zone -->
        <div
          class="relative flex flex-col items-center justify-center p-8 rounded-3xl border-2 border-dashed transition-all duration-200 cursor-pointer overflow-hidden group"
          :class="[
            isDragging
              ? 'border-primary-500/60 bg-primary-500/10 scale-[0.99]'
              : 'border-white/[0.12] bg-[#151a27] hover:border-primary-500/40 hover:bg-[#1a1f2d]'
          ]"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="handleDrop"
          @click="triggerFileInput"
        >
          <input
            ref="fileInput"
            type="file"
            multiple
            class="hidden"
            @change="handleFileInput"
          >

          <!-- Cloud Upload Glow Icon -->
          <div class="relative mb-3">
            <div class="flex size-13 items-center justify-center rounded-2xl bg-primary-500/15 text-primary-400 border border-primary-500/25 transition-transform group-hover:scale-105">
              <UIcon name="i-lucide-cloud-arrow-up" class="size-6" />
            </div>
          </div>

          <p class="text-sm font-semibold text-white text-center">
            Drag files here or <span class="text-primary-400 underline underline-offset-4 decoration-primary-400/40 hover:decoration-primary-400">browse from computer</span>
          </p>
          <p class="text-xs text-zinc-400 text-center mt-1 max-w-sm">
            Files stream straight to the provider — no copy is ever spooled on the server.
          </p>
        </div>

        <!-- Active Upload Progress Queue (Create UI Animated Card Rows) -->
        <div v-if="filesStore.uploadJobs.length > 0" class="space-y-3 pt-1">
          <div class="flex items-center justify-between text-xs font-bold px-1">
            <span class="text-white flex items-center gap-2">
              <UIcon name="i-lucide-activity" class="size-4 text-indigo-400" />
              Upload Queue ({{ filesStore.uploadJobs.length }})
            </span>
            <UButton
              variant="link"
              color="neutral"
              size="xs"
              label="Clear Completed"
              class="text-zinc-400 hover:text-white"
              @click="filesStore.clearFinishedJobs()"
            />
          </div>

          <div class="space-y-3 max-h-72 overflow-y-auto pr-1">
            <div
              v-for="job in filesStore.uploadJobs"
              :key="job.id"
              class="relative flex flex-col p-4 rounded-2xl border transition-all duration-200 shadow-xs group overflow-hidden"
              :class="[
                job.status === 'completed'
                  ? 'border-primary-500/30 bg-[#141d1c]'
                  : job.status === 'error'
                  ? 'border-red-500/30 bg-[#1e1719]'
                  : 'border-white/[0.08] bg-[#171c2a] hover:border-primary-500/30'
              ]"
            >
              <!-- Card Top Content: File Badge + Title + Metrics + Action Buttons -->
              <div class="flex items-center justify-between gap-3.5 mb-3">
                <!-- Left File Type Badge + File Info -->
                <div class="flex items-center gap-3.5 min-w-0">
                  <!-- Distinct File Badge Squircle -->
                  <div
                    class="relative flex flex-col items-center justify-center size-11 rounded-xl shrink-0 select-none shadow-xs transition-transform group-hover:scale-105"
                    :class="getFileBadge(job.file_name).color"
                  >
                    <UIcon :name="getFileBadge(job.file_name).icon" class="size-4 mb-0.5" />
                    <span class="text-[9px] font-bold leading-none tracking-wider font-mono">
                      {{ getFileBadge(job.file_name).ext }}
                    </span>
                  </div>

                  <!-- File Name & Dynamic Status Tracker -->
                  <div class="min-w-0 leading-tight">
                    <h4 class="font-bold text-sm text-white truncate group-hover:text-primary-300 transition-colors" :title="job.file_name">
                      {{ job.file_name }}
                    </h4>
                    
                    <div class="flex flex-wrap items-center gap-2 text-xs text-zinc-400 mt-1 font-medium">
                      <!-- Status State -->
                      <span
                        v-if="job.status === 'uploading'"
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-primary-500/10 text-primary-400 font-medium text-[10px] border border-primary-500/20"
                      >
                        <span class="size-1.5 rounded-full bg-primary-400 animate-ping" />
                        Uploading {{ job.progress }}%
                      </span>

                      <span
                        v-else-if="job.status === 'error'"
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-red-500/15 text-red-400 font-medium text-[10px] border border-red-500/25"
                      >
                        <UIcon name="i-lucide-circle-alert" class="size-3" />
                        Failed
                      </span>

                      <span
                        v-else-if="job.status === 'cancelled'"
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-zinc-800 text-zinc-300 font-medium text-[10px] border border-white/[0.08]"
                      >
                        <UIcon name="i-lucide-ban" class="size-3" />
                        Cancelled
                      </span>

                      <span
                        v-else-if="job.status === 'completed'"
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-primary-500/15 text-primary-400 font-medium text-[10px] border border-primary-500/25"
                      >
                        <UIcon name="i-lucide-check" class="size-3 stroke-[2.5]" />
                        Completed
                      </span>

                      <span
                        v-else
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-zinc-800 text-zinc-300 font-medium text-[10px] border border-white/[0.08]"
                      >
                        <UIcon name="i-lucide-route" class="size-3" />
                        Routing
                      </span>

                      <span class="text-zinc-600">•</span>

                      <!-- Byte Progress -->
                      <span v-if="job.status !== 'completed'" class="font-mono text-zinc-300 text-[11px]">
                        {{ formatBytes(job.bytes_uploaded) }} of {{ formatBytes(job.size_bytes) }}
                      </span>
                      <span v-else class="font-mono text-zinc-300 text-[11px]">
                        {{ formatBytes(job.size_bytes) }}
                      </span>

                      <!-- VFS Folder Destination -->
                      <span class="text-zinc-600 hidden sm:inline">•</span>
                      <span class="text-[11px] text-zinc-400 hidden sm:inline-flex items-center gap-1 font-mono">
                        <UIcon name="i-lucide-folder" class="size-3 text-amber-400" />
                        <span class="text-zinc-300">{{ job.target_folder_path || '/' }}</span>
                      </span>

                      <!-- Akun tujuan baru diketahui setelah router memutuskan -->
                      <template v-if="job.target_account_label">
                        <span class="text-zinc-600 hidden sm:inline">•</span>
                        <span class="text-[11px] text-zinc-400 hidden sm:inline-flex items-center gap-1 font-mono">
                          → <strong class="text-zinc-300 font-normal">{{ job.target_account_label }}</strong>
                        </span>
                      </template>
                    </div>

                    <p v-if="job.error_message" class="text-[11px] text-red-400 mt-1.5 leading-snug">
                      {{ job.error_message }}
                    </p>
                  </div>
                </div>

                <!-- Aksi: batalkan yang berjalan, singkirkan yang selesai -->
                <div class="flex items-center gap-1.5 shrink-0">
                  <div
                    v-if="job.status === 'completed'"
                    class="flex size-7 items-center justify-center rounded-lg bg-primary-500/15 text-primary-400 border border-primary-500/25"
                  >
                    <UIcon name="i-lucide-check" class="size-3.5 stroke-[2.5]" />
                  </div>

                  <!-- Upload berjalan bisa dibatalkan; request diputus dan
                       objek tak pernah tercatat di index. -->
                  <button
                    v-if="job.status === 'uploading' || job.status === 'routing'"
                    type="button"
                    title="Cancel upload"
                    class="flex size-7 items-center justify-center rounded-lg bg-zinc-800/60 hover:bg-rose-500/20 hover:text-rose-400 text-zinc-400 border border-white/[0.06] transition-all cursor-pointer"
                    @click="filesStore.cancelUpload(job.id)"
                  >
                    <UIcon name="i-lucide-x" class="size-3.5" />
                  </button>

                  <button
                    v-else
                    type="button"
                    title="Dismiss from list"
                    class="flex size-7 items-center justify-center rounded-lg bg-zinc-800/60 hover:bg-zinc-700 text-zinc-400 hover:text-white border border-white/[0.06] transition-all cursor-pointer"
                    @click="filesStore.dismissJob(job.id)"
                  >
                    <UIcon name="i-lucide-x" class="size-3.5" />
                  </button>
                </div>
              </div>

              <!-- Animated Progress Bar with Soft Emerald Fill -->
              <div class="h-1.5 w-full bg-zinc-900 border border-white/[0.06] rounded-full overflow-hidden relative mt-1">
                <div
                  class="h-full rounded-full transition-all duration-300 ease-out relative overflow-hidden"
                  :class="[
                    job.status === 'error'
                      ? 'bg-red-500'
                      : job.status === 'cancelled'
                      ? 'bg-zinc-600'
                      : 'bg-primary-500'
                  ]"
                  :style="{ width: `${job.progress}%` }"
                >
                  <!-- Shimmer wave line while uploading -->
                  <div
                    v-if="job.status === 'uploading'"
                    class="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent animate-shimmer"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex items-center justify-between w-full text-zinc-400">
        <span class="text-[11px] text-zinc-500 flex items-center gap-2">
          <UIcon name="i-lucide-shield-check" class="size-4 text-primary-400" />
          Streamed straight to the provider — nothing is spooled on the server.
        </span>
        <button
          type="button"
          class="px-5 py-2.5 rounded-xl text-xs font-semibold bg-primary-600 hover:bg-primary-500 text-white transition-all cursor-pointer shadow-xs"
          @click="filesStore.isUploadModalOpen = false"
        >
          Done / Close
        </button>
      </div>
    </template>
  </UModal>
</template>
