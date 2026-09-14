<script setup lang="ts">
import type { FileEntry } from '~/types'
import JsonViewer from './preview/JsonViewer.vue'
import HtmlViewer from './preview/HtmlViewer.vue'
import MarkdownViewer from './preview/MarkdownViewer.vue'
import CodeViewer from './preview/CodeViewer.vue'

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
const copiedField = ref<string | null>(null)

const isMediaLoading = ref(true)
const mediaError = ref(false)
const imageDimensions = ref('')

const SKELETON_LINES = ['82%', '64%', '91%', '47%', '73%', '88%', '35%', '69%', '79%', '52%']

function onMediaLoad(event?: Event) {
  isMediaLoading.value = false
  mediaError.value = false

  const el = event?.target as HTMLImageElement | undefined
  if (el?.naturalWidth) {
    imageDimensions.value = `${el.naturalWidth} × ${el.naturalHeight}`
  }
}

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

// Ekstensi berkas kode & teks yang didukung
const CODE_EXTENSIONS = [
  'txt', 'log', 'csv', 'tsv', 'yaml', 'yml', 'toml',
  'ini', 'env', 'conf', 'sql', 'sh', 'bash', 'zsh', 'go', 'py', 'rb', 'java',
  'c', 'h', 'cpp', 'rs', 'php', 'ts', 'tsx', 'js', 'jsx', 'vue', 'css', 'scss',
  'xml', 'svg', 'gitignore', 'dockerfile', 'proto', 'graphql', 'diff', 'patch'
]

const fileType = computed(() => {
  if (!props.file) return 'unknown'
  const name = props.file.name.toLowerCase()
  const mime = (props.file.mime || '').toLowerCase()
  const ext = name.includes('.') ? name.split('.').pop()! : ''

  // 1. PDF
  if (mime.includes('pdf') || ext === 'pdf') return 'pdf'

  // 2. JSON
  if (mime.includes('json') || ['json', 'jsonld', 'geojson', 'map'].includes(ext)) return 'json'

  // 3. HTML
  if (mime.includes('html') || ['html', 'htm'].includes(ext)) return 'html'

  // 4. Markdown
  if (mime.includes('markdown') || ['md', 'markdown', 'mdown'].includes(ext)) return 'markdown'

  // 5. Images
  if ((mime.startsWith('image/') && !mime.includes('svg')) || ['png', 'jpg', 'jpeg', 'webp', 'gif', 'bmp', 'avif', 'ico'].includes(ext)) return 'image'

  // 6. Videos
  if (mime.startsWith('video/') || ['mp4', 'webm', 'mov', 'mkv', 'avi'].includes(ext)) return 'video'

  // 7. Audio
  if (mime.startsWith('audio/') || ['mp3', 'wav', 'ogg', 'm4a', 'flac', 'aac'].includes(ext)) return 'audio'

  // 8. General Code & Config
  if (mime.startsWith('text/') || mime.includes('yaml') || mime.includes('xml') || mime.includes('sql') || mime.includes('javascript') || CODE_EXTENSIONS.includes(ext)) {
    return 'code'
  }

  // 9. Unsupported / Office / Archives
  return 'other'
})

const isTextType = computed(() => ['json', 'html', 'markdown', 'code'].includes(fileType.value))

const filesStore = useFilesStore()

const inlineUrl = computed(() =>
  props.file ? `${filesStore.downloadUrl(props.file.id)}?inline=1` : '')

const textContent = ref('')
const isLoadingText = ref(false)
const textError = ref('')

const TEXT_PREVIEW_LIMIT = 1024 * 1024 // 1 MB
const isTextTooLarge = computed(() => (props.file?.size_bytes ?? 0) > TEXT_PREVIEW_LIMIT)

async function loadTextPreview() {
  if (!props.file || !isTextType.value || isTextTooLarge.value) return
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
  if (props.open && isTextType.value) void loadTextPreview()
}, { immediate: true })

watch(() => props.open, (open) => {
  if (open) {
    resetPreviewState()
    if (isTextType.value) void loadTextPreview()
  }
})

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
      content: 'sm:max-w-4xl w-full h-[88vh] max-h-[840px] min-h-[580px] bg-white dark:bg-[#0d111a] border border-slate-200 dark:border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-slate-800 dark:text-zinc-200 flex flex-col',
      header: 'p-0 block shrink-0',
      body: 'p-0 flex-1 overflow-hidden flex flex-col min-h-0',
      footer: 'px-6 py-4 bg-slate-50 dark:bg-[#0b0e14] border-t border-slate-200/80 dark:border-white/[0.06] shrink-0'
    }"
    @update:open="emit('update:open', $event)"
  >
    <!-- Modal Header -->
    <template #header>
      <div class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200/80 dark:border-white/[0.06] bg-slate-50/90 dark:bg-[#141925] px-6 py-4">
        <div class="flex min-w-0 flex-1 items-center gap-3.5">
          <div class="flex size-11 items-center justify-center rounded-2xl bg-white dark:bg-[#1c2231] border border-slate-200 dark:border-white/[0.08] shrink-0 shadow-xs">
            <UIcon
              v-if="file"
              :name="getFileIcon(file.mime, file.name).icon"
              class="size-6"
              :class="getFileIcon(file.mime, file.name).color"
            />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <h2 class="truncate text-sm font-bold text-slate-900 dark:text-white tracking-tight" :title="file?.name">
                {{ file?.name }}
              </h2>
              <UBadge
                v-if="file?.is_chunked"
                label="Chunked Store"
                color="primary"
                variant="subtle"
                size="xs"
                class="rounded-md bg-primary-500/10 text-primary-600 dark:text-primary-400 border border-primary-500/20 text-[9px] px-1.5 py-0 shrink-0"
              />
            </div>
            <div class="mt-0.5 flex items-center gap-2 font-mono text-[11px] text-slate-500 dark:text-zinc-400">
              <span class="truncate max-w-[280px] sm:max-w-sm">{{ file?.virtual_path }}</span>
              <span class="shrink-0 text-slate-400 dark:text-zinc-600">•</span>
              <span class="shrink-0 font-medium text-primary-600 dark:text-primary-400">{{ file ? formatBytes(file.size_bytes) : '' }}</span>
            </div>
          </div>
        </div>

        <!-- Right Header: Tabs & Close -->
        <div class="flex items-center gap-3 shrink-0">
          <!-- View Tab Switcher -->
          <div class="flex items-center rounded-xl border border-slate-300/80 dark:border-white/[0.08] bg-slate-100 dark:bg-[#1c2231] p-1 shadow-2xs">
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer"
              :class="[
                activeTab === 'preview'
                  ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                  : 'text-slate-600 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white'
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
                  ? 'bg-white dark:bg-primary-600 text-slate-900 dark:text-white shadow-xs'
                  : 'text-slate-600 dark:text-zinc-400 hover:text-slate-900 dark:hover:text-white'
              ]"
              @click="activeTab = 'metadata'"
            >
              <UIcon name="i-lucide-info" class="size-3.5" />
              <span>Details</span>
            </button>
          </div>

          <button
            type="button"
            class="flex size-8 items-center justify-center rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-500 hover:text-slate-900 dark:bg-zinc-800/60 dark:hover:bg-zinc-700 dark:text-zinc-400 dark:hover:text-white border border-slate-200 dark:border-white/[0.06] transition-colors cursor-pointer"
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
      <div v-if="file" class="flex-1 overflow-hidden flex flex-col min-h-0 bg-slate-50 dark:bg-[#0d111a]">
        <!-- TAB 1: INTERACTIVE CONTENT PREVIEW -->
        <div v-if="activeTab === 'preview'" class="flex-1 min-h-0 h-full flex flex-col p-4 sm:p-5">
          <!-- 1. IMAGE VIEWER -->
          <div v-if="fileType === 'image'" class="flex flex-1 flex-col gap-3 min-h-0 h-full">
            <!-- Toolbar -->
            <div class="flex shrink-0 items-center justify-between gap-3 rounded-2xl border border-slate-200/80 dark:border-white/[0.07] bg-white dark:bg-[#151a27] px-4 py-2 text-xs">
              <div class="flex items-center gap-2 min-w-0">
                <span class="inline-flex items-center gap-1.5 rounded-lg bg-primary-500/10 px-2 py-0.5 font-mono text-[11px] font-medium text-primary-600 dark:text-primary-400 border border-primary-500/20">
                  {{ file.mime || 'image' }}
                </span>
                <span v-if="imageDimensions" class="truncate font-mono text-[11px] text-slate-600 dark:text-zinc-400">
                  {{ imageDimensions }}
                </span>
                <span v-else-if="isMediaLoading" class="font-mono text-[11px] text-slate-400 dark:text-zinc-500 animate-pulse">
                  Detecting dimensions...
                </span>
              </div>

              <div class="flex shrink-0 items-center gap-1">
                <button
                  type="button"
                  title="Zoom out"
                  class="flex size-7 cursor-pointer items-center justify-center rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-100 dark:text-zinc-400 dark:hover:bg-white/[0.08] dark:hover:text-white transition-colors disabled:cursor-not-allowed disabled:opacity-40"
                  :disabled="zoomLevel <= 50"
                  @click="zoomOut"
                >
                  <UIcon name="i-lucide-zoom-out" class="size-3.5" />
                </button>
                <span class="min-w-12 px-1 text-center font-mono text-[11px] text-primary-600 dark:text-primary-400 font-semibold select-none">{{ zoomLevel }}%</span>
                <button
                  type="button"
                  title="Zoom in"
                  class="flex size-7 cursor-pointer items-center justify-center rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-100 dark:text-zinc-400 dark:hover:bg-white/[0.08] dark:hover:text-white transition-colors disabled:cursor-not-allowed disabled:opacity-40"
                  :disabled="zoomLevel >= 200"
                  @click="zoomIn"
                >
                  <UIcon name="i-lucide-zoom-in" class="size-3.5" />
                </button>
                <button
                  type="button"
                  class="ml-1 cursor-pointer rounded-lg px-2 py-1 font-mono text-[10px] text-slate-500 hover:text-slate-900 hover:bg-slate-100 dark:text-zinc-400 dark:hover:bg-white/[0.08] dark:hover:text-white transition-colors disabled:opacity-40"
                  :disabled="zoomLevel === 100"
                  @click="resetZoom"
                >
                  Reset
                </button>
                <div class="w-px h-3.5 bg-slate-200 dark:bg-white/10 mx-1" />
                <a
                  :href="inlineUrl"
                  target="_blank"
                  title="Open original image in new tab"
                  class="flex size-7 cursor-pointer items-center justify-center rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-100 dark:text-zinc-400 dark:hover:bg-white/[0.08] dark:hover:text-white transition-colors"
                >
                  <UIcon name="i-lucide-external-link" class="size-3.5" />
                </a>
              </div>
            </div>

            <!-- Image Stage Canvas -->
            <div class="relative flex-1 min-h-0 h-full rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-slate-100 dark:bg-[#080b11] overflow-hidden flex items-center justify-center">
              <div
                v-if="isMediaLoading"
                class="absolute inset-0 z-10 flex flex-col items-center justify-center p-6 bg-slate-100/90 dark:bg-[#080b11]"
              >
                <div class="relative flex flex-col items-center gap-3 p-6 rounded-2xl bg-white dark:bg-[#121724] border border-slate-200 dark:border-white/10 shadow-lg">
                  <UIcon name="i-lucide-loader-2" class="size-6 animate-spin text-primary-500" />
                  <span class="text-xs text-slate-600 dark:text-zinc-400 font-mono">Loading image stream...</span>
                </div>
              </div>

              <!-- Error State -->
              <div v-if="mediaError" class="flex flex-col items-center gap-3 px-8 text-center">
                <UIcon name="i-lucide-image-off" class="size-8 text-rose-500" />
                <p class="text-xs font-semibold text-slate-800 dark:text-zinc-200">Unable to render image</p>
                <button
                  type="button"
                  class="px-3.5 py-1.5 text-xs font-semibold text-slate-700 bg-white dark:bg-[#1c2231] hover:bg-slate-100 dark:hover:bg-[#262d3e] border border-slate-300 dark:border-white/10 rounded-xl transition-colors cursor-pointer"
                  @click="resetPreviewState()"
                >
                  Retry
                </button>
              </div>

              <div
                v-show="!isMediaLoading && !mediaError"
                class="size-full overflow-auto flex items-center justify-center p-4 select-none"
              >
                <img
                  :ref="onImageMounted"
                  :src="inlineUrl"
                  :alt="file.name"
                  class="max-h-full max-w-full object-contain rounded-xl shadow-md transition-transform duration-200 cursor-zoom-in"
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
            <div class="relative flex-1 min-h-0 h-full rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-black overflow-hidden flex items-center justify-center">
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
            <div class="flex shrink-0 items-center justify-between rounded-2xl border border-slate-200/80 dark:border-white/[0.07] bg-white dark:bg-[#151a27] px-4 py-2.5 text-xs">
              <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-primary-500/10 text-primary-600 dark:text-primary-400 border border-primary-500/20 text-[11px] font-semibold">
                <span class="size-2 rounded-full bg-primary-500 animate-pulse" />
                Streaming directly from {{ file.account_label || 'provider' }}
              </span>
              <span class="text-slate-500 dark:text-zinc-400 font-mono text-[11px]">{{ formatBytes(file.size_bytes) }}</span>
            </div>
          </div>

          <!-- 3. AUDIO PLAYER -->
          <div v-else-if="fileType === 'audio'" class="flex-1 min-h-0 h-full flex flex-col items-center justify-center p-8">
            <div class="w-full max-w-md p-8 rounded-3xl bg-white dark:bg-[#151a27] border border-slate-200 dark:border-white/[0.08] shadow-xl space-y-6 text-center">
              <div class="flex size-20 items-center justify-center rounded-3xl bg-primary-500/15 text-primary-600 dark:text-primary-400 border border-primary-500/25 mx-auto shadow-md">
                <UIcon name="i-lucide-music" class="size-10" />
              </div>
              <div>
                <h3 class="font-bold text-sm text-slate-900 dark:text-white truncate">{{ file.name }}</h3>
                <p class="text-xs text-slate-500 dark:text-zinc-400 mt-1 font-mono">{{ formatBytes(file.size_bytes) }} • {{ file.mime || 'audio' }}</p>
              </div>
              <audio controls class="w-full" :src="inlineUrl" @loadeddata="onMediaLoad" @error="onMediaError">
                Your browser does not support HTML5 audio playback.
              </audio>
            </div>
          </div>

          <!-- 4. PDF DOCUMENT VIEWER -->
          <div v-else-if="fileType === 'pdf'" class="flex min-h-0 flex-1 flex-col gap-3 h-full">
            <div class="flex shrink-0 items-center justify-between gap-3 rounded-2xl border border-slate-200/80 dark:border-white/[0.07] bg-white dark:bg-[#151a27] px-4 py-2.5 text-xs">
              <div class="flex items-center gap-2 min-w-0">
                <UIcon name="i-lucide-file-text" class="size-4 text-rose-500 shrink-0" />
                <span class="truncate font-mono text-[11px] text-slate-800 dark:text-zinc-200 font-semibold">{{ file.name }}</span>
                <span class="text-slate-400 dark:text-zinc-600 hidden sm:inline">•</span>
                <span class="font-mono text-[11px] text-slate-500 dark:text-zinc-400 hidden sm:inline">{{ formatBytes(file.size_bytes) }}</span>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <a
                  :href="inlineUrl"
                  target="_blank"
                  title="Open PDF in new tab"
                  class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-slate-100 hover:bg-slate-200 text-slate-700 hover:text-slate-900 dark:bg-[#1c2231] dark:hover:bg-[#262d3e] dark:text-zinc-200 dark:hover:text-white border border-slate-200 dark:border-white/10 transition-colors"
                >
                  <UIcon name="i-lucide-external-link" class="size-3.5" />
                  <span>Open in Tab</span>
                </a>
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-primary-600 hover:bg-primary-500 text-white transition-colors shadow-xs"
                  @click="emit('download', file)"
                >
                  <UIcon name="i-lucide-download" class="size-3.5" />
                  <span>Download</span>
                </button>
              </div>
            </div>

            <div class="relative min-h-0 flex-1 h-full overflow-hidden rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-slate-100 dark:bg-[#080b11]">
              <iframe
                :src="inlineUrl"
                :title="file.name"
                class="size-full bg-white rounded-xl"
              />
            </div>
          </div>

          <!-- 5. JSON VIEWER -->
          <div v-else-if="fileType === 'json'" class="flex min-h-0 flex-1 flex-col h-full">
            <div v-if="isLoadingText" class="flex-1 min-h-0 overflow-hidden rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-white dark:bg-[#080b11] p-5">
              <div class="space-y-3 font-mono">
                <div v-for="(w, i) in SKELETON_LINES" :key="i" class="flex items-center gap-4">
                  <div class="h-3 w-7 rounded bg-slate-200 dark:bg-white/[0.04] text-right text-[10px] text-slate-400 dark:text-zinc-600 font-mono">{{ i + 1 }}</div>
                  <div class="h-3.5 rounded-md bg-slate-200 dark:bg-white/[0.07]" :style="{ width: w }" />
                </div>
              </div>
            </div>
            <div v-else-if="textError" class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-rose-500/20 bg-white dark:bg-[#080b11] p-8 text-center">
              <UIcon name="i-lucide-circle-alert" class="size-8 text-rose-500" />
              <p class="text-xs text-rose-500 font-mono">{{ textError }}</p>
              <button type="button" class="px-3.5 py-1.5 text-xs font-semibold rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-[#1c2231] dark:hover:bg-[#262d3e]" @click="loadTextPreview">
                Retry
              </button>
            </div>
            <JsonViewer
              v-else
              :raw-json="textContent"
              :filename="file.name"
              :file-size="file.size_bytes"
            />
          </div>

          <!-- 6. HTML VIEWER -->
          <div v-else-if="fileType === 'html'" class="flex min-h-0 flex-1 flex-col h-full">
            <div v-if="isLoadingText" class="flex-1 min-h-0 overflow-hidden rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-white dark:bg-[#080b11] p-5">
              <div class="space-y-3 font-mono">
                <div v-for="(w, i) in SKELETON_LINES" :key="i" class="flex items-center gap-4">
                  <div class="h-3 w-7 rounded bg-slate-200 dark:bg-white/[0.04] text-right text-[10px] text-slate-400 dark:text-zinc-600 font-mono">{{ i + 1 }}</div>
                  <div class="h-3.5 rounded-md bg-slate-200 dark:bg-white/[0.07]" :style="{ width: w }" />
                </div>
              </div>
            </div>
            <div v-else-if="textError" class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-rose-500/20 bg-white dark:bg-[#080b11] p-8 text-center">
              <UIcon name="i-lucide-circle-alert" class="size-8 text-rose-500" />
              <p class="text-xs text-rose-500 font-mono">{{ textError }}</p>
              <button type="button" class="px-3.5 py-1.5 text-xs font-semibold rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-[#1c2231] dark:hover:bg-[#262d3e]" @click="loadTextPreview">
                Retry
              </button>
            </div>
            <HtmlViewer
              v-else
              :html-content="textContent"
              :filename="file.name"
              :file-size="file.size_bytes"
            />
          </div>

          <!-- 7. MARKDOWN VIEWER -->
          <div v-else-if="fileType === 'markdown'" class="flex min-h-0 flex-1 flex-col h-full">
            <div v-if="isLoadingText" class="flex-1 min-h-0 overflow-hidden rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-white dark:bg-[#080b11] p-5">
              <div class="space-y-3 font-mono">
                <div v-for="(w, i) in SKELETON_LINES" :key="i" class="flex items-center gap-4">
                  <div class="h-3 w-7 rounded bg-slate-200 dark:bg-white/[0.04] text-right text-[10px] text-slate-400 dark:text-zinc-600 font-mono">{{ i + 1 }}</div>
                  <div class="h-3.5 rounded-md bg-slate-200 dark:bg-white/[0.07]" :style="{ width: w }" />
                </div>
              </div>
            </div>
            <div v-else-if="textError" class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-rose-500/20 bg-white dark:bg-[#080b11] p-8 text-center">
              <UIcon name="i-lucide-circle-alert" class="size-8 text-rose-500" />
              <p class="text-xs text-rose-500 font-mono">{{ textError }}</p>
              <button type="button" class="px-3.5 py-1.5 text-xs font-semibold rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-[#1c2231] dark:hover:bg-[#262d3e]" @click="loadTextPreview">
                Retry
              </button>
            </div>
            <MarkdownViewer
              v-else
              :markdown="textContent"
              :filename="file.name"
              :file-size="file.size_bytes"
            />
          </div>

          <!-- 8. UNIVERSAL CODE & TEXT VIEWER -->
          <div v-else-if="fileType === 'code'" class="flex min-h-0 flex-1 flex-col h-full">
            <div v-if="isTextTooLarge" class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-white dark:bg-[#080b11] p-8 text-center">
              <UIcon name="i-lucide-file-text" class="size-10 text-slate-400 dark:text-zinc-500" />
              <p class="text-xs text-slate-600 dark:text-zinc-400 max-w-sm">
                File is too large to preview directly ({{ formatBytes(file.size_bytes) }}). Please download to inspect.
              </p>
            </div>
            <div v-else-if="isLoadingText" class="flex-1 min-h-0 overflow-hidden rounded-2xl border border-slate-200 dark:border-white/[0.08] bg-white dark:bg-[#080b11] p-5">
              <div class="space-y-3 font-mono">
                <div v-for="(w, i) in SKELETON_LINES" :key="i" class="flex items-center gap-4">
                  <div class="h-3 w-7 rounded bg-slate-200 dark:bg-white/[0.04] text-right text-[10px] text-slate-400 dark:text-zinc-600 font-mono">{{ i + 1 }}</div>
                  <div class="h-3.5 rounded-md bg-slate-200 dark:bg-white/[0.07]" :style="{ width: w }" />
                </div>
              </div>
            </div>
            <div v-else-if="textError" class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-rose-500/20 bg-white dark:bg-[#080b11] p-8 text-center">
              <UIcon name="i-lucide-circle-alert" class="size-8 text-rose-500" />
              <p class="text-xs text-rose-500 font-mono">{{ textError }}</p>
              <button type="button" class="px-3.5 py-1.5 text-xs font-semibold rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-[#1c2231] dark:hover:bg-[#262d3e]" @click="loadTextPreview">
                Retry
              </button>
            </div>
            <CodeViewer
              v-else
              :code="textContent"
              :language="file.name.includes('.') ? file.name.split('.').pop()! : 'plaintext'"
              :filename="file.name"
              :file-size="file.size_bytes"
            />
          </div>

          <!-- 9. UNSUPPORTED FORMATS -->
          <div v-else class="flex-1 min-h-0 h-full flex flex-col items-center justify-center p-8 space-y-5 text-center">
            <div class="flex size-24 items-center justify-center rounded-3xl bg-slate-100 dark:bg-[#1b2130] border border-slate-200 dark:border-white/[0.08] text-primary-600 dark:text-primary-400 shadow-lg">
              <UIcon :name="getFileIcon(file.mime, file.name).icon" class="size-12" />
            </div>
            <div>
              <h3 class="font-bold text-base text-slate-900 dark:text-white">{{ file.name }}</h3>
              <p class="text-xs text-slate-500 dark:text-zinc-400 max-w-sm mt-1.5 leading-relaxed">
                This format cannot be rendered directly in the browser. Download it to open with an application on your device.
              </p>
            </div>
            <button
              type="button"
              class="px-5 py-2.5 rounded-xl text-xs font-bold bg-primary-600 hover:bg-primary-500 text-white flex items-center gap-2 cursor-pointer shadow-md transition-all hover:scale-[1.02]"
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
          <div class="flex items-center gap-4 p-4 rounded-2xl bg-white dark:bg-[#141925] border border-slate-200 dark:border-white/[0.08] shadow-xs">
            <div class="flex size-14 shrink-0 items-center justify-center rounded-2xl bg-slate-50 dark:bg-[#1c2231] border border-slate-200 dark:border-white/[0.08] shadow-inner">
              <UIcon
                :name="getFileIcon(file.mime, file.name).icon"
                class="size-7"
                :class="getFileIcon(file.mime, file.name).color"
              />
            </div>
            <div class="min-w-0 flex-1">
              <h3 class="font-bold text-sm text-slate-900 dark:text-white truncate">{{ file.name }}</h3>
              <p class="font-mono text-xs text-slate-500 dark:text-zinc-400 mt-0.5 truncate">{{ file.virtual_path }}</p>
              <div class="mt-2 flex flex-wrap items-center gap-2">
                <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-slate-100 dark:bg-[#1c2231] border border-slate-200 dark:border-white/[0.06] text-[10px] font-semibold text-slate-700 dark:text-zinc-300">
                  <UIcon :name="getProviderMeta(file.provider).icon" class="size-3" />
                  {{ file.account_label }}
                </span>
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-primary-500/10 text-primary-600 dark:text-primary-400 border border-primary-500/20 text-[10px] font-medium font-mono">
                  {{ file.is_chunked ? 'Model B (Chunked)' : 'Model A (Whole-File)' }}
                </span>
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-slate-200 dark:bg-zinc-800 text-slate-700 dark:text-zinc-400 text-[10px] font-mono">
                  {{ formatBytes(file.size_bytes) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Metadata Grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Card 1: Virtual Filesystem Details -->
            <div class="p-4 rounded-2xl bg-white dark:bg-[#141925] border border-slate-200 dark:border-white/[0.07] space-y-3 shadow-xs">
              <h4 class="font-bold text-slate-900 dark:text-zinc-200 flex items-center gap-2 text-xs">
                <UIcon name="i-lucide-folder-tree" class="size-4 text-primary-500 dark:text-primary-400" />
                Virtual Filesystem (VFS)
              </h4>
              <div class="space-y-2.5 text-xs">
                <div class="flex items-center justify-between py-1 border-b border-slate-100 dark:border-white/[0.06]">
                  <span class="text-slate-500 dark:text-zinc-400">Virtual Path</span>
                  <div class="flex items-center gap-1.5 max-w-[65%]">
                    <span class="font-mono text-primary-600 dark:text-primary-400 font-semibold truncate text-[11px]" :title="file.virtual_path">
                      {{ file.virtual_path }}
                    </span>
                    <button
                      type="button"
                      title="Copy Virtual Path"
                      class="shrink-0 p-1 rounded-md text-slate-400 hover:text-slate-700 hover:bg-slate-100 dark:text-zinc-400 dark:hover:text-white dark:hover:bg-white/[0.08] transition-colors cursor-pointer"
                      @click="copyMetadata(file.virtual_path, 'vpath')"
                    >
                      <UIcon :name="copiedField === 'vpath' ? 'i-lucide-check' : 'i-lucide-copy'" class="size-3" :class="copiedField === 'vpath' ? 'text-primary-500' : ''" />
                    </button>
                  </div>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-slate-100 dark:border-white/[0.06]">
                  <span class="text-slate-500 dark:text-zinc-400">File ID</span>
                  <div class="flex items-center gap-1.5 max-w-[65%]">
                    <span class="font-mono text-slate-700 dark:text-zinc-300 truncate text-[11px]">{{ file.id }}</span>
                    <button
                      type="button"
                      title="Copy File ID"
                      class="shrink-0 p-1 rounded-md text-slate-400 hover:text-slate-700 hover:bg-slate-100 dark:text-zinc-400 dark:hover:text-white dark:hover:bg-white/[0.08] transition-colors cursor-pointer"
                      @click="copyMetadata(file.id, 'id')"
                    >
                      <UIcon :name="copiedField === 'id' ? 'i-lucide-check' : 'i-lucide-copy'" class="size-3" :class="copiedField === 'id' ? 'text-primary-500' : ''" />
                    </button>
                  </div>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-slate-100 dark:border-white/[0.06]">
                  <span class="text-slate-500 dark:text-zinc-400">File Size</span>
                  <span class="font-mono text-slate-800 dark:text-zinc-200 font-bold text-[11px]">
                    {{ formatBytes(file.size_bytes) }} <span class="text-slate-400 dark:text-zinc-500 font-normal">({{ file.size_bytes.toLocaleString() }} B)</span>
                  </span>
                </div>
                <div class="flex items-center justify-between py-1">
                  <span class="text-slate-500 dark:text-zinc-400">MIME Type</span>
                  <span class="font-mono text-slate-700 dark:text-zinc-300 text-[11px]">{{ file.mime || 'application/octet-stream' }}</span>
                </div>
              </div>
            </div>

            <!-- Card 2: Cloud Storage Placement -->
            <div class="p-4 rounded-2xl bg-white dark:bg-[#141925] border border-slate-200 dark:border-white/[0.07] space-y-3 shadow-xs">
              <h4 class="font-bold text-slate-900 dark:text-zinc-200 flex items-center gap-2 text-xs">
                <UIcon name="i-lucide-cloud" class="size-4 text-primary-500 dark:text-primary-400" />
                Physical Cloud Storage
              </h4>
              <div class="space-y-2.5 text-xs">
                <div class="flex items-center justify-between py-1 border-b border-slate-100 dark:border-white/[0.06]">
                  <span class="text-slate-500 dark:text-zinc-400">Cloud Provider</span>
                  <span class="font-semibold text-slate-800 dark:text-zinc-200 flex items-center gap-1.5">
                    <UIcon :name="getProviderMeta(file.provider).icon" class="size-3.5" />
                    {{ file.account_label }}
                  </span>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-slate-100 dark:border-white/[0.06]">
                  <span class="text-slate-500 dark:text-zinc-400">Storage Architecture</span>
                  <span class="font-mono text-primary-600 dark:text-primary-400 font-medium text-[11px]">
                    {{ file.is_chunked ? 'Model B (Chunked Blocks)' : 'Model A (Whole-File)' }}
                  </span>
                </div>
                <div class="flex items-center justify-between py-1 border-b border-slate-100 dark:border-white/[0.06]">
                  <span class="text-slate-500 dark:text-zinc-400">Daemon Remote</span>
                  <span class="font-mono text-slate-700 dark:text-zinc-300 text-[11px]">acc_{{ file.account_id.replace('acc-', '') }}:</span>
                </div>
                <div class="flex items-center justify-between py-1">
                  <span class="text-slate-500 dark:text-zinc-400">Last Modified</span>
                  <span class="text-slate-700 dark:text-zinc-300 text-[11px]">{{ formatDate(file.modified_at) }}</span>
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
          class="inline-flex cursor-pointer items-center gap-1.5 rounded-xl px-3.5 py-2 text-xs font-semibold text-rose-600 dark:text-rose-400 transition-colors hover:bg-rose-50 dark:hover:bg-rose-500/10 hover:text-rose-700 dark:hover:text-rose-300"
          @click="emit('delete', file); emit('update:open', false)"
        >
          <UIcon name="i-lucide-trash-2" class="size-4" />
          <span>Delete</span>
        </button>

        <!-- Right Buttons: Rename, Move, Migrate, Download -->
        <div class="flex flex-wrap items-center gap-2">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-white hover:bg-slate-100 text-slate-700 hover:text-slate-900 border border-slate-200 dark:bg-[#1c2231] dark:hover:bg-[#262d3e] dark:text-zinc-300 dark:hover:text-white dark:border-white/[0.08] transition-all cursor-pointer shadow-2xs"
            @click="emit('rename', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-edit-3" class="size-4 text-sky-500 dark:text-sky-400" />
            <span>Rename</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-white hover:bg-slate-100 text-slate-700 hover:text-slate-900 border border-slate-200 dark:bg-[#1c2231] dark:hover:bg-[#262d3e] dark:text-zinc-300 dark:hover:text-white dark:border-white/[0.08] transition-all cursor-pointer shadow-2xs"
            @click="emit('moveToFolder', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-folder-input" class="size-4 text-amber-500 dark:text-amber-400" />
            <span>Move Folder</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-white hover:bg-slate-100 text-slate-700 hover:text-slate-900 border border-slate-200 dark:bg-[#1c2231] dark:hover:bg-[#262d3e] dark:text-zinc-300 dark:hover:text-white dark:border-white/[0.08] transition-all cursor-pointer shadow-2xs"
            @click="emit('migrateProvider', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-arrow-right-left" class="size-4 text-primary-600 dark:text-primary-400" />
            <span>Migrate Cloud</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-bold bg-primary-600 hover:bg-primary-500 text-white transition-all cursor-pointer shadow-xs"
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
