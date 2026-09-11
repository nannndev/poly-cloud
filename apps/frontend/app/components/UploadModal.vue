<script setup lang="ts">
import type { StorageProvider } from '~/types'

const filesStore = useFilesStore()
const accountsStore = useAccountsStore()
const { formatBytes } = useFormatters()

const isDragging = ref(false)
const selectedTarget = ref<string>('auto')
const selectedFolderId = ref<string>(filesStore.currentFolderId || '__root__')
const fileInput = ref<HTMLInputElement | null>(null)

// Sync destination folder whenever modal opens
watch(() => filesStore.isUploadModalOpen, (isOpen) => {
  if (isOpen) {
    selectedFolderId.value = filesStore.currentFolderId || '__root__'
  }
})

// Hierarchical folder options
const folderOptions = computed(() => {
  return filesStore.allFoldersHierarchical.map(f => ({
    value: f.id === null ? '__root__' : f.id,
    label: f.id === null ? '📁 Root (/)' : `${'  '.repeat(f.depth)}📁 ${f.name} (${f.path})`
  }))
})

// Smart routing recommendation (account with most free bytes)
const recommendedAccount = computed(() => {
  const active = accountsStore.activeAccounts
  if (!active.length) return null
  return [...active].sort((a, b) => (b.free_bytes || 0) - (a.free_bytes || 0))[0]
})

const targetAccountOptions = computed(() => {
  const options = [
    {
      value: 'auto',
      label: recommendedAccount.value
        ? `⚡ Smart Routing (Auto) → ${recommendedAccount.value.label} (${formatBytes(recommendedAccount.value.free_bytes)} free)`
        : '⚡ Smart Routing (Auto)'
    }
  ]

  accountsStore.activeAccounts.forEach(acc => {
    options.push({
      value: acc.id,
      label: `${acc.label} (${formatBytes(acc.free_bytes)} free)`
    })
  })

  return options
})

function getFileBadge(fileName: string) {
  const name = (fileName || '').toLowerCase()
  const parts = name.split('.')
  const rawExt = parts.length > 1 ? parts.pop() || 'FILE' : 'FILE'

  if (name.endsWith('.xlsx') || name.endsWith('.xls') || name.endsWith('.csv')) {
    return {
      ext: 'XLSX',
      color: 'bg-gradient-to-br from-emerald-500 to-teal-600 text-white shadow-md shadow-emerald-500/30',
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
}

function processFiles(fileList: File[]) {
  let targetAcc = undefined
  if (selectedTarget.value !== 'auto') {
    const found = accountsStore.accounts.find(a => a.id === selectedTarget.value)
    if (found) {
      targetAcc = { id: found.id, label: found.label, provider: found.provider }
    }
  } else if (recommendedAccount.value) {
    targetAcc = {
      id: recommendedAccount.value.id,
      label: recommendedAccount.value.label,
      provider: recommendedAccount.value.provider
    }
  }

  const folderId = selectedFolderId.value === '__root__' ? null : selectedFolderId.value
  const folderObj = folderId ? filesStore.folders.find(f => f.id === folderId) : null
  const folderPath = folderObj ? folderObj.path : '/'

  for (const f of fileList) {
    filesStore.simulateUpload({
      name: f.name,
      size: f.size || 2400000,
      targetAccount: targetAcc,
      targetFolderId: folderId,
      targetFolderPath: folderPath
    })
  }
}

function triggerFileInput() {
  fileInput.value?.click()
}

function triggerDemoUpload(name: string, sizeBytes: number) {
  let targetAcc = undefined
  if (selectedTarget.value !== 'auto') {
    const found = accountsStore.accounts.find(a => a.id === selectedTarget.value)
    if (found) targetAcc = { id: found.id, label: found.label, provider: found.provider }
  } else if (recommendedAccount.value) {
    targetAcc = {
      id: recommendedAccount.value.id,
      label: recommendedAccount.value.label,
      provider: recommendedAccount.value.provider
    }
  }

  const folderId = selectedFolderId.value === '__root__' ? null : selectedFolderId.value
  const folderObj = folderId ? filesStore.folders.find(f => f.id === folderId) : null
  const folderPath = folderObj ? folderObj.path : '/'

  filesStore.simulateUpload({
    name,
    size: sizeBytes,
    targetAccount: targetAcc,
    targetFolderId: folderId,
    targetFolderPath: folderPath
  })
}
</script>

<template>
  <UModal
    v-model:open="filesStore.isUploadModalOpen"
    :ui="{
      content: 'sm:max-w-2xl bg-[#0c0c0e] dark:bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
      body: 'p-6 space-y-5',
      footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
    }"
  >
    <!-- Modal Header -->
    <template #header>
      <div class="px-6 pt-5 pb-4 bg-[#111114] flex items-start justify-between gap-4 border-b border-white/[0.06]">
        <div class="flex items-start gap-3.5">
          <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-emerald-500/15 text-emerald-400 border border-emerald-500/25">
            <UIcon name="i-lucide-upload-cloud" class="size-5" />
          </div>
          <div>
            <h2 class="text-base font-bold tracking-tight text-white flex items-center gap-2">
              Smart Multi-Cloud File Upload
              <UBadge label="Zero Vendor Lock-in" color="emerald" variant="subtle" size="xs" class="font-medium text-[9px] rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" />
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
          <div class="p-3.5 rounded-2xl bg-[#121215] border border-white/[0.07] space-y-2">
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
              class="w-full rounded-xl bg-[#16161a] border border-white/[0.08] text-white"
              icon="i-lucide-folder-open"
              size="md"
            />
          </div>

          <!-- Placement 2: Physical Cloud Storage Target -->
          <div class="p-3.5 rounded-2xl bg-[#121215] border border-white/[0.07] space-y-2">
            <div class="flex items-center justify-between text-xs font-semibold text-zinc-300">
              <span class="flex items-center gap-2">
                <UIcon name="i-lucide-route" class="size-4 text-emerald-400" />
                Cloud Target (Smart Routing)
              </span>
              <span class="text-[10px] text-zinc-400 font-mono">Physical</span>
            </div>
            <USelect
              v-model="selectedTarget"
              :items="targetAccountOptions"
              class="w-full rounded-xl bg-[#16161a] border border-white/[0.08] text-white"
              icon="i-lucide-layers-3"
              size="md"
            />
          </div>
        </div>

        <!-- Animated Drag & Drop Zone -->
        <div
          class="relative flex flex-col items-center justify-center p-8 rounded-3xl border-2 border-dashed transition-all duration-200 cursor-pointer overflow-hidden group"
          :class="[
            isDragging
              ? 'border-emerald-500/60 bg-emerald-500/10 scale-[0.99]'
              : 'border-white/[0.12] bg-[#121215] hover:border-emerald-500/40 hover:bg-[#141418]'
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
            <div class="flex size-13 items-center justify-center rounded-2xl bg-emerald-500/15 text-emerald-400 border border-emerald-500/25 transition-transform group-hover:scale-105">
              <UIcon name="i-lucide-cloud-arrow-up" class="size-6" />
            </div>
          </div>

          <p class="text-sm font-semibold text-white text-center">
            Drag files here or <span class="text-emerald-400 underline underline-offset-4 decoration-emerald-400/40 hover:decoration-emerald-400">browse from computer</span>
          </p>
          <p class="text-xs text-zinc-400 text-center mt-1 max-w-sm">
            Automatic distributed chunking activates for files over 100 MB.
          </p>

          <!-- Quick Test Demo Pills -->
          <div class="mt-4 pt-3.5 border-t border-white/[0.06] w-full flex flex-col sm:flex-row items-center justify-center gap-2 text-xs" @click.stop>
            <span class="text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Demo Animation:</span>
            <div class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 border border-emerald-500/25 transition-all flex items-center gap-1.5 cursor-pointer shadow-xs"
                @click="triggerDemoUpload('create-ui.xlsx', 2411724)"
              >
                <UIcon name="i-lucide-file-spreadsheet" class="size-3.5" />
                <span>create-ui.xlsx (2.3 MB)</span>
              </button>

              <button
                type="button"
                class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-zinc-800/70 text-zinc-300 hover:bg-zinc-700 hover:text-white border border-white/[0.08] transition-all flex items-center gap-1.5 cursor-pointer shadow-xs"
                @click="triggerDemoUpload('enterprise-spec.pdf', 14885000)"
              >
                <UIcon name="i-lucide-file-text" class="size-3.5" />
                <span>spec.pdf (14.2 MB)</span>
              </button>

              <button
                type="button"
                class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-zinc-800/70 text-zinc-300 hover:bg-zinc-700 hover:text-white border border-white/[0.08] transition-all flex items-center gap-1.5 cursor-pointer shadow-xs"
                @click="triggerDemoUpload('docker-compose.yaml', 85000)"
              >
                <UIcon name="i-lucide-file-code" class="size-3.5" />
                <span>docker.yaml (85 KB)</span>
              </button>
            </div>
          </div>
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
              @click="filesStore.uploadJobs = filesStore.uploadJobs.filter(j => j.status !== 'completed')"
            />
          </div>

          <div class="space-y-3 max-h-72 overflow-y-auto pr-1">
            <div
              v-for="job in filesStore.uploadJobs"
              :key="job.id"
              class="relative flex flex-col p-4 rounded-2xl border transition-all duration-200 shadow-xs group overflow-hidden"
              :class="[
                job.status === 'completed'
                  ? 'border-emerald-500/30 bg-[#121614]'
                  : job.status === 'paused'
                  ? 'border-white/[0.08] bg-[#151518]'
                  : 'border-white/[0.08] bg-[#131317] hover:border-emerald-500/30'
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
                    <h4 class="font-bold text-sm text-white truncate group-hover:text-emerald-300 transition-colors" :title="job.file_name">
                      {{ job.file_name }}
                    </h4>
                    
                    <div class="flex flex-wrap items-center gap-2 text-xs text-zinc-400 mt-1 font-medium">
                      <!-- Status State -->
                      <span
                        v-if="job.status === 'uploading'"
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-400 font-medium text-[10px] border border-emerald-500/20"
                      >
                        <span class="size-1.5 rounded-full bg-emerald-400 animate-ping" />
                        Uploading {{ job.progress }}%
                      </span>

                      <span
                        v-else-if="job.status === 'paused'"
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-zinc-800 text-zinc-300 font-medium text-[10px] border border-white/[0.08]"
                      >
                        <span class="size-1.5 rounded-full bg-zinc-400" />
                        Paused
                      </span>

                      <span
                        v-else-if="job.status === 'completed'"
                        class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-emerald-500/15 text-emerald-400 font-medium text-[10px] border border-emerald-500/25"
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

                      <!-- Speed indicator -->
                      <template v-if="job.status === 'uploading' && job.speed_mbps">
                        <span class="text-zinc-600">•</span>
                        <span class="font-mono text-[10px] text-emerald-400">{{ job.speed_mbps }} MB/s</span>
                      </template>

                      <!-- VFS Folder Destination -->
                      <span class="text-zinc-600 hidden sm:inline">•</span>
                      <span class="text-[11px] text-zinc-400 hidden sm:inline-flex items-center gap-1 font-mono">
                        <UIcon name="i-lucide-folder" class="size-3 text-amber-400" />
                        <span class="text-zinc-300">{{ job.target_folder_path || '/' }}</span>
                      </span>

                      <!-- Cloud Target Pill -->
                      <span class="text-zinc-600 hidden sm:inline">•</span>
                      <span class="text-[11px] text-zinc-400 hidden sm:inline-flex items-center gap-1 font-mono">
                        → <strong class="text-zinc-300 font-normal">{{ job.target_account_label }}</strong>
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Right Action Buttons (Pause, Resume, Cancel) -->
                <div class="flex items-center gap-1.5 shrink-0">
                  <!-- Pause Button -->
                  <button
                    v-if="job.status === 'uploading'"
                    type="button"
                    title="Pause Upload"
                    class="flex size-7 items-center justify-center rounded-lg bg-zinc-800/80 hover:bg-zinc-700 text-zinc-300 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
                    @click="filesStore.pauseUpload(job.id)"
                  >
                    <UIcon name="i-lucide-pause" class="size-3.5 fill-current" />
                  </button>

                  <!-- Resume Button -->
                  <button
                    v-else-if="job.status === 'paused'"
                    type="button"
                    title="Resume Upload"
                    class="flex size-7 items-center justify-center rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white border border-emerald-500/40 transition-all cursor-pointer shadow-xs"
                    @click="filesStore.resumeUpload(job.id)"
                  >
                    <UIcon name="i-lucide-play" class="size-3.5 fill-current ml-0.5" />
                  </button>

                  <!-- Completed Success Checkmark Badge -->
                  <div
                    v-else-if="job.status === 'completed'"
                    class="flex size-7 items-center justify-center rounded-lg bg-emerald-500/15 text-emerald-400 border border-emerald-500/25"
                  >
                    <UIcon name="i-lucide-check" class="size-3.5 stroke-[2.5]" />
                  </div>

                  <!-- Cancel / Dismiss Button -->
                  <button
                    type="button"
                    :title="job.status === 'completed' ? 'Dismiss' : 'Cancel Upload'"
                    class="flex size-7 items-center justify-center rounded-lg bg-zinc-800/60 hover:bg-rose-500/20 hover:text-rose-400 text-zinc-400 border border-white/[0.06] transition-all cursor-pointer"
                    @click="filesStore.cancelUpload(job.id)"
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
                    job.status === 'completed'
                      ? 'bg-emerald-500'
                      : job.status === 'paused'
                      ? 'bg-zinc-500'
                      : 'bg-emerald-500'
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
          <UIcon name="i-lucide-shield-check" class="size-4 text-emerald-400" />
          Zero-Spool direct multi-cloud streaming with SHA-256 validation.
        </span>
        <button
          type="button"
          class="px-5 py-2.5 rounded-xl text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white transition-all cursor-pointer shadow-xs"
          @click="filesStore.isUploadModalOpen = false"
        >
          Done / Close
        </button>
      </div>
    </template>
  </UModal>
</template>
