<script setup lang="ts">
import type { FileEntry } from '~/types'

const props = defineProps<{
  open: boolean
  file: FileEntry | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'download': [file: FileEntry]
  'moveToFolder': [file: FileEntry]
  'migrateProvider': [file: FileEntry]
  'delete': [file: FileEntry]
  'rename': [file: FileEntry]
}>()

const { formatBytes, formatDate, getProviderMeta, getFileIcon } = useFormatters()

const activeTab = ref<'preview' | 'metadata'>('preview')
const zoomLevel = ref(100)
const isCopied = ref(false)
const copiedField = ref<string | null>(null)

/**
 * Gambar, video, dan PDF diambil langsung oleh elemennya sendiri, jadi satu-
 * satunya kabar tentang kemajuannya adalah event load/error miliknya. Tanpa
 * dilacak, modal menampilkan kotak kosong selama byte masih mengalir dari
 * provider — dan kotak kosong itu tak bisa dibedakan dari pratinjau yang gagal.
 */
const isMediaLoading = ref(true)
const mediaError = ref(false)

/**
 * Lebar baris placeholder saat teks masih dimuat untuk efek skeleton menyerupai kode.
 */
const SKELETON_LINES = ['82%', '64%', '91%', '47%', '73%', '88%', '35%', '69%', '79%', '52%']

/** Dimensi asli gambar, baru diketahui setelah elemennya selesai memuat. */
const imageDimensions = ref('')

function onMediaLoad(event?: Event) {
  isMediaLoading.value = false
  mediaError.value = false

  const el = event?.target as HTMLImageElement | undefined
  if (el?.naturalWidth) {
    imageDimensions.value = `${el.naturalWidth} × ${el.naturalHeight}`
  }
}

/**
 * Gambar dirender dengan v-show, jadi elemennya sudah ada di DOM dan mulai
 * mengambil byte sebelum listener @load terpasang. Bila byte-nya datang lebih
 * cepat daripada pemasangan listener — cache browser, atau berkas kecil —
 * event `load` sudah lewat dan tak akan terulang: skeleton membeku selamanya
 * padahal gambarnya sebenarnya sudah siap.
 *
 * Karena itu status `complete` diperiksa langsung saat elemen terpasang,
 * alih-alih hanya menunggu event.
 */
function onImageMounted(el: Element | ComponentPublicInstance | null) {
  const img = el as HTMLImageElement | null
  if (!img?.complete) return
  if (img.naturalWidth > 0) {
    onMediaLoad({ target: img } as unknown as Event)
  } else {
    onMediaError()
  }
}

function onMediaError() {
  isMediaLoading.value = false
  mediaError.value = true
}

// Ekstensi berkas teks yang aman ditampilkan sebagai teks bernomor baris.
const TEXT_EXTENSIONS = [
  'txt', 'log', 'md', 'markdown', 'csv', 'tsv', 'json', 'yaml', 'yml', 'toml',
  'ini', 'env', 'conf', 'sql', 'sh', 'bash', 'zsh', 'go', 'py', 'rb', 'java',
  'c', 'h', 'cpp', 'rs', 'php', 'ts', 'tsx', 'js', 'jsx', 'vue', 'css', 'scss',
  'html', 'xml', 'svg', 'gitignore', 'dockerfile'
]

const fileType = computed(() => {
  if (!props.file) return 'unknown'
  const name = props.file.name.toLowerCase()
  const mime = (props.file.mime || '').toLowerCase()
  const ext = name.includes('.') ? name.split('.').pop()! : ''

  // MIME dari provider dipercaya lebih dulu; ekstensi jadi cadangan untuk file
  // yang providernya tak melaporkan tipe.
  if (mime.startsWith('image/') && !mime.includes('svg')) return 'image'
  if (mime.startsWith('video/')) return 'video'
  if (mime.startsWith('audio/')) return 'audio'
  if (mime.includes('pdf')) return 'pdf'

  if (['png', 'jpg', 'jpeg', 'webp', 'gif', 'bmp', 'avif', 'ico'].includes(ext)) return 'image'
  if (['mp4', 'webm', 'mov', 'mkv', 'avi'].includes(ext)) return 'video'
  if (['mp3', 'wav', 'ogg', 'm4a', 'flac', 'aac'].includes(ext)) return 'audio'
  if (ext === 'pdf') return 'pdf'

  if (mime.startsWith('text/') || mime.includes('json') || mime.includes('yaml')
    || mime.includes('xml') || mime.includes('sql') || mime.includes('javascript')) {
    return 'text'
  }
  if (TEXT_EXTENSIONS.includes(ext)) return 'text'

  // Sisanya (dokumen Office, arsip, biner) tak bisa dirender browser.
  return 'other'
})

const filesStore = useFilesStore()

// Semua pratinjau menunjuk endpoint download dengan ?inline=1, sehingga browser
// merender isinya alih-alih mengunduh. Byte-nya dialirkan backend langsung dari
// provider — tak ada salinan di server maupun di memori JS.
const inlineUrl = computed(() =>
  props.file ? `${filesStore.downloadUrl(props.file.id)}?inline=1` : '')

// Teks (kode/markdown/json) perlu diambil isinya untuk ditampilkan bernomor baris.
const textContent = ref('')
const isLoadingText = ref(false)
const textError = ref('')

// Batas aman: file teks raksasa tak perlu dimuat utuh ke memori browser.
const TEXT_PREVIEW_LIMIT = 512 * 1024
const isTextTooLarge = computed(() => (props.file?.size_bytes ?? 0) > TEXT_PREVIEW_LIMIT)

async function loadTextPreview() {
  if (!props.file || fileType.value !== 'text' || isTextTooLarge.value) return
  isLoadingText.value = true
  textError.value = ''
  try {
    const res = await fetch(inlineUrl.value)
    if (!res.ok) throw new Error(`Could not load file contents (${res.status})`)
    textContent.value = await res.text()
  } catch (err) {
    textError.value = err instanceof Error ? err.message : 'Could not load file contents'
  } finally {
    isLoadingText.value = false
  }
}

/**
 * Satu-satunya tempat status pratinjau disetel ulang saat file berganti.
 */
function resetPreviewState() {
  activeTab.value = 'preview'
  zoomLevel.value = 100
  isMediaLoading.value = true
  mediaError.value = false
  imageDimensions.value = ''
  textContent.value = ''
  textError.value = ''
}

watch(() => props.file?.id, () => {
  resetPreviewState()
  if (props.open) void loadTextPreview()
}, { immediate: true })

watch(() => props.open, (open) => {
  if (open) {
    resetPreviewState()
    void loadTextPreview()
  }
})

const codeLines = computed(() => textContent.value.split('\n'))

function copySnippet() {
  navigator.clipboard?.writeText(textContent.value)
  isCopied.value = true
  setTimeout(() => {
    isCopied.value = false
  }, 2000)
}

function copyMetadata(value: string, field: string) {
  navigator.clipboard?.writeText(value)
  copiedField.value = field
  setTimeout(() => {
    if (copiedField.value === field) copiedField.value = null
  }, 2000)
}

function zoomIn() {
  zoomLevel.value = Math.min(200, zoomLevel.value + 25)
}

function zoomOut() {
  zoomLevel.value = Math.max(50, zoomLevel.value - 25)
}

function resetZoom() {
  zoomLevel.value = 100
}

function toggleZoom() {
  zoomLevel.value = zoomLevel.value === 100 ? 150 : 100
}
</script>

<template>
  <UModal
    :open="open"
    :ui="{
      content: 'sm:max-w-4xl w-full h-[88vh] max-h-[820px] min-h-[580px] bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200 flex flex-col',
      header: 'p-0 block shrink-0',
      body: 'p-0 flex-1 overflow-hidden flex flex-col min-h-0',
      footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06] shrink-0'
    }"
    @update:open="emit('update:open', $event)"
  >
    <!-- Modal Header -->
    <template #header>
      <div class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-white/[0.06] bg-[#141925] px-6 py-4">
        <div class="flex min-w-0 flex-1 items-center gap-3.5">
          <div class="flex size-11 items-center justify-center rounded-2xl bg-[#1c2231] border border-white/[0.08] shrink-0 shadow-xs shadow-black/20">
            <UIcon
              v-if="file"
              :name="getFileIcon(file.mime, file.name).icon"
              class="size-6"
              :class="getFileIcon(file.mime, file.name).color"
            />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <h2 class="truncate text-sm font-bold text-white tracking-tight" :title="file?.name">
                {{ file?.name }}
              </h2>
              <UBadge
                v-if="file?.is_chunked"
                label="Chunked Store"
                color="primary"
                variant="subtle"
                size="xs"
                class="rounded-md bg-primary-500/10 text-primary-400 border border-primary-500/20 text-[9px] px-1.5 py-0 shrink-0"
              />
            </div>
            <div class="mt-0.5 flex items-center gap-2 font-mono text-[11px] text-zinc-400">
              <span class="truncate max-w-[280px] sm:max-w-sm">{{ file?.virtual_path }}</span>
              <span class="shrink-0 text-zinc-600">•</span>
              <span class="shrink-0 font-medium text-primary-400">{{ file ? formatBytes(file.size_bytes) : '' }}</span>
            </div>
          </div>
        </div>

        <!-- Right Header: Tabs & Close -->
        <div class="flex items-center gap-3 shrink-0">
          <!-- View Tab Switcher -->
          <div class="flex items-center rounded-xl border border-white/[0.08] bg-[#1c2231] p-1 shadow-xs">
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer"
              :class="[
                activeTab === 'preview'
                  ? 'bg-primary-600 text-white shadow-xs'
                  : 'text-zinc-400 hover:text-white'
              ]"
              @click="activeTab = 'preview'"
            >
              <UIcon name="i-lucide-eye" class="size-3.5" />
              <span>Preview</span>
            </button>

            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer"
              :class="[
                activeTab === 'metadata'
                  ? 'bg-primary-600 text-white shadow-xs'
                  : 'text-zinc-400 hover:text-white'
              ]"
              @click="activeTab = 'metadata'"
            >
              <UIcon name="i-lucide-info" class="size-3.5" />
              <span>Details</span>
            </button>
          </div>

          <button
            type="button"
            class="flex size-8 items-center justify-center rounded-xl bg-zinc-800/60 hover:bg-zinc-700 text-zinc-400 hover:text-white border border-white/[0.06] transition-colors cursor-pointer"
            aria-label="Close modal"
            @click="emit('update:open', false)"
          >
            <UIcon name="i-lucide-x" class="size-4" />
          </button>
        </div>
      </div>
    </template>

    <!-- Modal Body -->
    <template #body>
      <div v-if="file" class="flex-1 overflow-hidden flex flex-col min-h-0 bg-[#0d111a]">
        <!-- TAB 1: INTERACTIVE CONTENT PREVIEW -->
        <div v-if="activeTab === 'preview'" class="flex-1 min-h-0 h-full flex flex-col p-5">
          <!-- 1. IMAGE VIEWER -->
          <div v-if="fileType === 'image'" class="flex flex-1 flex-col gap-3 min-h-0 h-full">
            <!-- Toolbar -->
            <div class="flex shrink-0 items-center justify-between gap-3 rounded-2xl border border-white/[0.07] bg-[#151a27] px-4 py-2 text-xs">
              <div class="flex items-center gap-2 min-w-0">
                <span class="inline-flex items-center gap-1.5 rounded-lg bg-primary-500/10 px-2 py-0.5 font-mono text-[11px] font-medium text-primary-400 border border-primary-500/20">
                  {{ file.mime || 'image/png' }}
                </span>
                <span v-if="imageDimensions" class="truncate font-mono text-[11px] text-zinc-400">
                  {{ imageDimensions }}
                </span>
                <span v-else-if="isMediaLoading" class="font-mono text-[11px] text-zinc-500 animate-pulse">
                  Detecting dimensions...
                </span>
              </div>

              <div class="flex shrink-0 items-center gap-1">
                <button
                  type="button"
                  title="Zoom out"
                  class="flex size-7 cursor-pointer items-center justify-center rounded-lg text-zinc-400 hover:bg-white/[0.08] hover:text-white transition-colors disabled:cursor-not-allowed disabled:opacity-40"
                  :disabled="zoomLevel <= 50"
                  @click="zoomOut"
                >
                  <UIcon name="i-lucide-zoom-out" class="size-3.5" />
                </button>
                <span class="min-w-12 px-1 text-center font-mono text-[11px] text-primary-400 font-semibold select-none">{{ zoomLevel }}%</span>
                <button
                  type="button"
                  title="Zoom in"
                  class="flex size-7 cursor-pointer items-center justify-center rounded-lg text-zinc-400 hover:bg-white/[0.08] hover:text-white transition-colors disabled:cursor-not-allowed disabled:opacity-40"
                  :disabled="zoomLevel >= 200"
                  @click="zoomIn"
                >
                  <UIcon name="i-lucide-zoom-in" class="size-3.5" />
                </button>
                <button
                  type="button"
                  class="ml-1 cursor-pointer rounded-lg px-2 py-1 font-mono text-[10px] text-zinc-400 hover:bg-white/[0.08] hover:text-white transition-colors disabled:opacity-40"
                  :disabled="zoomLevel === 100"
                  @click="resetZoom"
                >
                  Reset
                </button>
                <div class="w-px h-3.5 bg-white/10 mx-1" />
                <a
                  :href="inlineUrl"
                  target="_blank"
                  title="Open original image in new tab"
                  class="flex size-7 cursor-pointer items-center justify-center rounded-lg text-zinc-400 hover:bg-white/[0.08] hover:text-white transition-colors"
                >
                  <UIcon name="i-lucide-external-link" class="size-3.5" />
                </a>
              </div>
            </div>

            <!-- Image Stage Canvas -->
            <div class="relative flex-1 min-h-0 h-full rounded-2xl border border-white/[0.08] bg-[#080b11] overflow-hidden flex items-center justify-center">
              <!-- Shimmer Loading Overlay -->
              <div
                v-if="isMediaLoading"
                class="absolute inset-0 z-10 flex flex-col items-center justify-center p-6 bg-[#080b11]"
              >
                <!-- Sweep shimmer effect across background -->
                <div class="absolute inset-0 overflow-hidden pointer-events-none">
                  <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.08] to-transparent animate-shimmer" />
                </div>

                <!-- Centered Preview Skeleton Card -->
                <div class="relative flex flex-col items-center gap-4 max-w-sm w-full p-8 rounded-3xl border border-white/[0.07] bg-[#121724]/90 shadow-2xl backdrop-blur-md">
                  <div class="relative flex size-16 items-center justify-center rounded-2xl bg-primary-500/15 border border-primary-500/25 text-primary-400 shadow-lg shadow-primary-500/10">
                    <UIcon name="i-lucide-image" class="size-8 animate-pulse" />
                    <span class="absolute -top-1 -right-1 flex size-3">
                      <span class="absolute inline-flex size-full animate-ping rounded-full bg-primary-400 opacity-75" />
                      <span class="relative inline-flex size-3 rounded-full bg-primary-500" />
                    </span>
                  </div>

                  <div class="w-full space-y-2 text-center">
                    <div class="h-3.5 w-44 mx-auto rounded-md bg-white/[0.08] animate-pulse" />
                    <div class="h-2.5 w-28 mx-auto rounded-md bg-white/[0.05] animate-pulse" />
                  </div>

                  <div class="flex items-center gap-2 rounded-full border border-white/[0.06] bg-[#182032] px-3.5 py-1.5 text-[11px] text-zinc-400 font-mono">
                    <UIcon name="i-lucide-loader-2" class="size-3.5 animate-spin text-primary-400" />
                    <span>Streaming image from cloud...</span>
                  </div>
                </div>
              </div>

              <!-- Error State -->
              <div
                v-if="mediaError"
                class="flex flex-col items-center gap-3 px-8 text-center"
              >
                <div class="flex size-14 items-center justify-center rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-400">
                  <UIcon name="i-lucide-image-off" class="size-7" />
                </div>
                <div>
                  <p class="text-xs font-semibold text-zinc-200">Unable to render image</p>
                  <p class="text-[11px] text-zinc-400 mt-1 max-w-xs leading-relaxed">The image could not be loaded from the cloud provider stream.</p>
                </div>
                <button
                  type="button"
                  class="mt-1 px-3.5 py-1.5 text-xs font-semibold text-zinc-300 bg-[#1c2231] hover:bg-[#262d3e] border border-white/10 rounded-xl transition-colors cursor-pointer"
                  @click="resetPreviewState()"
                >
                  Retry Loading
                </button>
              </div>

              <!-- Loaded Image with Pan/Zoom & Double Click Toggle -->
              <div
                v-show="!isMediaLoading && !mediaError"
                class="size-full overflow-auto flex items-center justify-center p-4 select-none"
              >
                <img
                  :ref="onImageMounted"
                  :src="inlineUrl"
                  :alt="file.name"
                  class="max-h-full max-w-full object-contain rounded-xl shadow-2xl transition-transform duration-200 cursor-zoom-in"
                  :style="{ transform: `scale(${zoomLevel / 100})` }"
                  @load="onMediaLoad"
                  @error="onMediaError"
                  @dblclick="toggleZoom"
                >
              </div>
            </div>
          </div>

          <!-- 2. VIDEO PLAYER -->
          <div v-else-if="fileType === 'video'" class="flex flex-1 flex-col gap-3 min-h-0 h-full">
            <div class="relative flex-1 min-h-0 h-full rounded-2xl border border-white/[0.08] bg-black overflow-hidden flex items-center justify-center">
              <!-- Video Shimmer Loading Overlay -->
              <div
                v-if="isMediaLoading"
                class="absolute inset-0 z-10 flex flex-col items-center justify-center p-6 bg-[#080b11]"
              >
                <div class="absolute inset-0 overflow-hidden pointer-events-none">
                  <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.08] to-transparent animate-shimmer" />
                </div>
                <div class="relative flex flex-col items-center gap-4 max-w-sm w-full p-8 rounded-3xl border border-white/[0.07] bg-[#121724]/90 shadow-2xl backdrop-blur-md">
                  <div class="relative flex size-16 items-center justify-center rounded-2xl bg-primary-500/15 border border-primary-500/25 text-primary-400 shadow-lg shadow-primary-500/10">
                    <UIcon name="i-lucide-video" class="size-8 animate-pulse" />
                    <span class="absolute -top-1 -right-1 flex size-3">
                      <span class="absolute inline-flex size-full animate-ping rounded-full bg-primary-400 opacity-75" />
                      <span class="relative inline-flex size-3 rounded-full bg-primary-500" />
                    </span>
                  </div>
                  <div class="w-full space-y-2 text-center">
                    <div class="h-3.5 w-44 mx-auto rounded-md bg-white/[0.08] animate-pulse" />
                    <div class="h-2.5 w-28 mx-auto rounded-md bg-white/[0.05] animate-pulse" />
                  </div>
                  <div class="flex items-center gap-2 rounded-full border border-white/[0.06] bg-[#182032] px-3.5 py-1.5 text-[11px] text-zinc-400 font-mono">
                    <UIcon name="i-lucide-loader-2" class="size-3.5 animate-spin text-primary-400" />
                    <span>Buffering video stream...</span>
                  </div>
                </div>
              </div>

              <video
                controls
                class="max-h-full max-w-full object-contain"
                :src="inlineUrl"
                @loadeddata="onMediaLoad"
                @error="onMediaError"
              >
                Your browser does not support HTML5 video playback.
              </video>
            </div>

            <div class="flex shrink-0 items-center justify-between rounded-2xl border border-white/[0.07] bg-[#151a27] px-4 py-2.5 text-xs">
              <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-primary-500/10 text-primary-400 border border-primary-500/20 text-[11px] font-semibold">
                <span class="size-2 rounded-full bg-primary-400 animate-pulse" />
                Streaming directly from {{ file.account_label || 'provider' }}
              </span>
              <span class="text-zinc-400 font-mono text-[11px]">{{ formatBytes(file.size_bytes) }}</span>
            </div>
          </div>

          <!-- 3. AUDIO PLAYER -->
          <div v-else-if="fileType === 'audio'" class="flex-1 min-h-0 h-full flex flex-col items-center justify-center p-8">
            <div class="w-full max-w-md p-8 rounded-3xl bg-[#151a27] border border-white/[0.08] shadow-2xl space-y-6 text-center">
              <div class="flex size-20 items-center justify-center rounded-3xl bg-primary-500/15 text-primary-400 border border-primary-500/25 mx-auto shadow-lg shadow-primary-500/10">
                <UIcon name="i-lucide-music" class="size-10" />
              </div>

              <div>
                <h3 class="font-bold text-sm text-white truncate">{{ file.name }}</h3>
                <p class="text-xs text-zinc-400 mt-1 font-mono">{{ formatBytes(file.size_bytes) }} • {{ file.mime || 'audio' }}</p>
              </div>

              <div class="relative">
                <div
                  v-if="isMediaLoading"
                  class="h-14 w-full rounded-2xl bg-[#1a2133] border border-white/10 relative overflow-hidden flex items-center justify-center gap-2"
                >
                  <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.1] to-transparent animate-shimmer" />
                  <UIcon name="i-lucide-loader-2" class="size-4 animate-spin text-primary-400" />
                  <span class="text-xs text-zinc-400 font-mono">Loading audio stream...</span>
                </div>
                <audio
                  v-show="!isMediaLoading"
                  controls
                  class="w-full"
                  :src="inlineUrl"
                  @loadeddata="onMediaLoad"
                  @error="onMediaError"
                >
                  Your browser does not support HTML5 audio playback.
                </audio>
              </div>
            </div>
          </div>

          <!-- 4. PDF DOCUMENT VIEWER -->
          <div v-else-if="fileType === 'pdf'" class="flex min-h-0 flex-1 flex-col gap-3 h-full">
            <div class="flex shrink-0 items-center justify-between gap-3 rounded-2xl border border-white/[0.07] bg-[#151a27] px-4 py-2.5 text-xs">
              <span class="truncate font-mono text-[11px] text-zinc-400">Rendered by built-in PDF engine</span>
              <div class="flex items-center gap-3">
                <span class="shrink-0 font-mono text-[11px] text-zinc-400">{{ formatBytes(file.size_bytes) }}</span>
                <a
                  :href="inlineUrl"
                  target="_blank"
                  title="Open PDF in new tab"
                  class="flex size-7 items-center justify-center rounded-lg text-zinc-400 hover:bg-white/[0.08] hover:text-white transition-colors"
                >
                  <UIcon name="i-lucide-external-link" class="size-3.5" />
                </a>
              </div>
            </div>

            <div class="relative min-h-0 flex-1 h-full overflow-hidden rounded-2xl border border-white/[0.08] bg-[#080b11]">
              <div
                v-if="isMediaLoading"
                class="absolute inset-0 z-10 flex flex-col items-center justify-center p-6 bg-[#080b11]"
              >
                <div class="absolute inset-0 overflow-hidden pointer-events-none">
                  <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.08] to-transparent animate-shimmer" />
                </div>
                <div class="relative flex flex-col items-center gap-4 max-w-sm w-full p-8 rounded-3xl border border-white/[0.07] bg-[#121724]/90 shadow-2xl backdrop-blur-md">
                  <div class="relative flex size-16 items-center justify-center rounded-2xl bg-rose-500/15 border border-rose-500/25 text-rose-400 shadow-lg shadow-rose-500/10">
                    <UIcon name="i-lucide-file-text" class="size-8 animate-pulse" />
                  </div>
                  <div class="w-full space-y-2 text-center">
                    <div class="h-3.5 w-44 mx-auto rounded-md bg-white/[0.08] animate-pulse" />
                    <div class="h-2.5 w-28 mx-auto rounded-md bg-white/[0.05] animate-pulse" />
                  </div>
                  <div class="flex items-center gap-2 rounded-full border border-white/[0.06] bg-[#182032] px-3.5 py-1.5 text-[11px] text-zinc-400 font-mono">
                    <UIcon name="i-lucide-loader-2" class="size-3.5 animate-spin text-rose-400" />
                    <span>Loading PDF document...</span>
                  </div>
                </div>
              </div>
              <iframe
                :src="inlineUrl"
                :title="file.name"
                class="size-full bg-white"
                @load="onMediaLoad"
              />
            </div>
          </div>

          <!-- 5. TEXT & CONFIG FILES -->
          <div v-else-if="fileType === 'text'" class="flex min-h-0 flex-1 flex-col gap-3 h-full">
            <div class="flex shrink-0 items-center justify-between gap-3 rounded-2xl border border-white/[0.07] bg-[#151a27] px-4 py-2 text-xs">
              <div class="flex items-center gap-2 min-w-0">
                <UIcon name="i-lucide-file-code-2" class="size-4 text-primary-400 shrink-0" />
                <span class="truncate font-mono text-[11px] text-primary-400 font-semibold">{{ file.name }}</span>
                <span class="text-zinc-600">•</span>
                <span class="font-mono text-[11px] text-zinc-400">{{ formatBytes(file.size_bytes) }}</span>
              </div>
              <div class="flex items-center gap-2">
                <button
                  v-if="textContent"
                  type="button"
                  class="inline-flex shrink-0 cursor-pointer items-center gap-1.5 rounded-xl bg-[#212736] px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-[#262d3e] hover:text-white"
                  @click="copySnippet"
                >
                  <UIcon :name="isCopied ? 'i-lucide-check' : 'i-lucide-copy'" class="size-3.5" :class="isCopied ? 'text-primary-400' : ''" />
                  <span>{{ isCopied ? 'Copied' : 'Copy Text' }}</span>
                </button>
                <a
                  :href="inlineUrl"
                  target="_blank"
                  title="Open raw file in new tab"
                  class="flex size-7 items-center justify-center rounded-xl text-zinc-400 hover:bg-white/[0.08] hover:text-white transition-colors"
                >
                  <UIcon name="i-lucide-external-link" class="size-3.5" />
                </a>
              </div>
            </div>

            <!-- File teks raksasa -->
            <div
              v-if="isTextTooLarge"
              class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-white/[0.08] bg-[#080b11] p-8 text-center"
            >
              <UIcon name="i-lucide-file-text" class="size-10 text-zinc-500" />
              <p class="text-xs text-zinc-400 max-w-sm leading-relaxed">
                File is too large to preview directly ({{ formatBytes(file.size_bytes) }}).
                Please download it to inspect the content.
              </p>
            </div>

            <!-- Shimmering Code Skeleton -->
            <div
              v-else-if="isLoadingText"
              class="min-h-0 flex-1 overflow-hidden rounded-2xl border border-white/[0.08] bg-[#080b11] p-5 relative"
            >
              <div class="absolute inset-0 overflow-hidden pointer-events-none">
                <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.07] to-transparent animate-shimmer" />
              </div>
              <div class="space-y-3 font-mono">
                <div
                  v-for="(w, i) in SKELETON_LINES"
                  :key="i"
                  class="flex items-center gap-4"
                >
                  <div class="h-3 w-7 rounded bg-white/[0.04] text-right text-[10px] text-zinc-600 font-mono select-none">{{ i + 1 }}</div>
                  <div class="h-3.5 rounded-md bg-white/[0.07]" :style="{ width: w }" />
                </div>
              </div>
            </div>

            <div
              v-else-if="textError"
              class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-red-500/20 bg-[#080b11] p-8 text-center"
            >
              <UIcon name="i-lucide-circle-alert" class="size-8 text-red-400" />
              <p class="text-xs text-red-400">{{ textError }}</p>
              <button
                type="button"
                class="px-3.5 py-1.5 text-xs font-semibold text-zinc-300 bg-[#1c2231] hover:bg-[#262d3e] border border-white/10 rounded-xl transition-colors cursor-pointer"
                @click="loadTextPreview"
              >
                Retry
              </button>
            </div>

            <div v-else class="flex-1 overflow-auto rounded-2xl border border-white/[0.08] bg-[#080b11] font-mono text-xs p-4 shadow-inner">
              <table class="w-full border-collapse">
                <tbody>
                  <tr v-for="(line, idx) in codeLines" :key="idx" class="hover:bg-white/[0.03] transition-colors leading-6">
                    <td class="w-12 select-none text-right pr-4 text-zinc-600 text-[11px] align-top font-mono">{{ idx + 1 }}</td>
                    <td class="text-zinc-200 whitespace-pre font-mono selection:bg-primary-500/30">{{ line }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 6. UNSUPPORTED FORMATS -->
          <div v-else class="flex-1 min-h-0 h-full flex flex-col items-center justify-center p-8 space-y-5 text-center">
            <div class="flex size-24 items-center justify-center rounded-3xl bg-[#1b2130] border border-white/[0.08] text-primary-400 shadow-xl shadow-primary-500/5">
              <UIcon :name="getFileIcon(file.mime, file.name).icon" class="size-12" />
            </div>
            <div>
              <h3 class="font-bold text-base text-white">{{ file.name }}</h3>
              <p class="text-xs text-zinc-400 max-w-sm mt-1.5 leading-relaxed">
                This format cannot be rendered directly in the browser. Download it to open with an application on your system.
              </p>
            </div>
            <button
              type="button"
              class="px-5 py-2.5 rounded-xl text-xs font-bold bg-primary-600 hover:bg-primary-500 text-white flex items-center gap-2 cursor-pointer shadow-lg shadow-primary-600/20 transition-all hover:scale-[1.02]"
              @click="emit('download', file)"
            >
              <UIcon name="i-lucide-download" class="size-4" />
              <span>Download ({{ formatBytes(file.size_bytes) }})</span>
            </button>
          </div>
        </div>

        <!-- TAB 2: METADATA & TELEMETRY -->
        <div v-else class="flex-1 overflow-y-auto p-6 space-y-5 text-xs">
          <!-- Summary Banner -->
          <div class="flex items-center gap-4 p-4 rounded-2xl bg-[#141925] border border-white/[0.08]">
            <div class="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-[#1c2231] border border-white/[0.08] shadow-inner">
              <UIcon
                :name="getFileIcon(file.mime, file.name).icon"
                class="size-7"
                :class="getFileIcon(file.mime, file.name).color"
              />
            </div>
            <div class="min-w-0 flex-1">
              <h3 class="font-bold text-sm text-white truncate">{{ file.name }}</h3>
              <p class="font-mono text-xs text-zinc-400 mt-0.5 truncate">{{ file.virtual_path }}</p>
              <div class="mt-2 flex flex-wrap items-center gap-2">
                <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-[#1c2231] border border-white/[0.06] text-[10px] font-semibold text-zinc-300">
                  <UIcon :name="getProviderMeta(file.provider).icon" class="size-3" />
                  {{ file.account_label }}
                </span>
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-primary-500/10 text-primary-400 border border-primary-500/20 text-[10px] font-medium font-mono">
                  {{ file.is_chunked ? 'Model B (Chunked)' : 'Model A (Whole-File)' }}
                </span>
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-zinc-800 text-zinc-400 text-[10px] font-mono">
                  {{ formatBytes(file.size_bytes) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Metadata Grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Card 1: Virtual Filesystem Details -->
            <div class="p-4 rounded-2xl bg-[#141925] border border-white/[0.07] space-y-3">
              <h4 class="font-bold text-zinc-200 flex items-center gap-2 text-xs">
                <UIcon name="i-lucide-folder-tree" class="size-4 text-primary-400" />
                Virtual Filesystem (VFS)
              </h4>
              <div class="space-y-2.5 text-xs">
                <div class="flex items-center justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">Virtual Path</span>
                  <div class="flex items-center gap-1.5 max-w-[65%]">
                    <span class="font-mono text-primary-400 font-semibold truncate text-[11px]" :title="file.virtual_path">
                      {{ file.virtual_path }}
                    </span>
                    <button
                      type="button"
                      title="Copy Virtual Path"
                      class="shrink-0 p-1 rounded-md text-zinc-400 hover:text-white hover:bg-white/[0.08] transition-colors cursor-pointer"
                      @click="copyMetadata(file.virtual_path, 'vpath')"
                    >
                      <UIcon :name="copiedField === 'vpath' ? 'i-lucide-check' : 'i-lucide-copy'" class="size-3" :class="copiedField === 'vpath' ? 'text-primary-400' : ''" />
                    </button>
                  </div>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">File ID</span>
                  <div class="flex items-center gap-1.5 max-w-[65%]">
                    <span class="font-mono text-zinc-300 truncate text-[11px]">{{ file.id }}</span>
                    <button
                      type="button"
                      title="Copy File ID"
                      class="shrink-0 p-1 rounded-md text-zinc-400 hover:text-white hover:bg-white/[0.08] transition-colors cursor-pointer"
                      @click="copyMetadata(file.id, 'id')"
                    >
                      <UIcon :name="copiedField === 'id' ? 'i-lucide-check' : 'i-lucide-copy'" class="size-3" :class="copiedField === 'id' ? 'text-primary-400' : ''" />
                    </button>
                  </div>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">File Size</span>
                  <span class="font-mono text-zinc-200 font-bold text-[11px]">
                    {{ formatBytes(file.size_bytes) }} <span class="text-zinc-500 font-normal">({{ file.size_bytes.toLocaleString() }} B)</span>
                  </span>
                </div>
                <div class="flex items-center justify-between py-1">
                  <span class="text-zinc-400">MIME Type</span>
                  <span class="font-mono text-zinc-300 text-[11px]">{{ file.mime || 'application/octet-stream' }}</span>
                </div>
              </div>
            </div>

            <!-- Card 2: Cloud Storage Placement -->
            <div class="p-4 rounded-2xl bg-[#141925] border border-white/[0.07] space-y-3">
              <h4 class="font-bold text-zinc-200 flex items-center gap-2 text-xs">
                <UIcon name="i-lucide-cloud" class="size-4 text-primary-400" />
                Physical Cloud Storage
              </h4>
              <div class="space-y-2.5 text-xs">
                <div class="flex items-center justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">Cloud Provider</span>
                  <span class="font-semibold text-zinc-200 flex items-center gap-1.5">
                    <UIcon :name="getProviderMeta(file.provider).icon" class="size-3.5" />
                    {{ file.account_label }}
                  </span>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">Storage Architecture</span>
                  <span class="font-mono text-primary-400 font-medium text-[11px]">
                    {{ file.is_chunked ? 'Model B (Chunked Blocks)' : 'Model A (Whole-File)' }}
                  </span>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">Daemon Remote</span>
                  <span class="font-mono text-zinc-300 text-[11px]">acc_{{ file.account_id.replace('acc-', '') }}:</span>
                </div>
                <div class="flex items-center justify-between py-1">
                  <span class="text-zinc-400">Last Modified</span>
                  <span class="text-zinc-300 text-[11px]">{{ formatDate(file.modified_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Modal Footer Actions -->
    <template #footer>
      <div v-if="file" class="flex w-full flex-wrap items-center justify-between gap-3">
        <!-- Delete Button -->
        <button
          type="button"
          class="inline-flex cursor-pointer items-center gap-1.5 rounded-xl px-3.5 py-2 text-xs font-semibold text-rose-400 transition-colors hover:bg-rose-500/10 hover:text-rose-300"
          @click="emit('delete', file); emit('update:open', false)"
        >
          <UIcon name="i-lucide-trash-2" class="size-4" />
          <span>Delete</span>
        </button>

        <!-- Right Buttons: Rename, Move, Migrate, Download -->
        <div class="flex flex-wrap items-center gap-2">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-[#1c2231] hover:bg-[#262d3e] text-zinc-300 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
            @click="emit('rename', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-edit-3" class="size-4 text-sky-400" />
            <span>Rename</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-[#1c2231] hover:bg-[#262d3e] text-zinc-300 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
            @click="emit('moveToFolder', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-folder-input" class="size-4 text-amber-400" />
            <span>Move Folder</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-[#1c2231] hover:bg-[#262d3e] text-zinc-300 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
            @click="emit('migrateProvider', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-arrow-right-left" class="size-4 text-primary-400" />
            <span>Migrate Cloud</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-bold bg-primary-600 hover:bg-primary-500 text-white transition-all cursor-pointer shadow-xs shadow-primary-600/20"
            @click="emit('download', file)"
          >
            <UIcon name="i-lucide-download" class="size-4" />
            <span>Download</span>
          </button>
        </div>
      </div>
    </template>
  </UModal>
</template>
