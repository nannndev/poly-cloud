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

watch(() => props.file, () => {
  activeTab.value = 'preview'
  zoomLevel.value = 100
})

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

watch(() => props.file?.id, () => {
  textContent.value = ''
  textError.value = ''
  if (props.open) void loadTextPreview()
}, { immediate: true })

watch(() => props.open, (open) => {
  if (open) void loadTextPreview()
})

const codeLines = computed(() => textContent.value.split('\n'))

function copySnippet() {
  navigator.clipboard?.writeText(textContent.value)
  isCopied.value = true
  setTimeout(() => {
    isCopied.value = false
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
</script>

<template>
  <UModal
    :open="open"
    :ui="{
      content: 'sm:max-w-4xl w-full bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200 flex flex-col max-h-[92vh]',
      body: 'p-0 flex-1 overflow-hidden flex flex-col min-h-0',
      footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06] shrink-0'
    }"
    @update:open="emit('update:open', $event)"
  >
    <!-- Modal Header -->
    <template #header>
      <div class="px-6 py-4 bg-[#111114] flex items-center justify-between gap-4 border-b border-white/[0.06] shrink-0">
        <div class="flex items-center gap-3.5 min-w-0">
          <div class="p-2.5 rounded-2xl bg-[#16161b] border border-white/[0.06] shrink-0 shadow-xs">
            <UIcon
              v-if="file"
              :name="getFileIcon(file.mime, file.name).icon"
              class="size-6"
              :class="getFileIcon(file.mime, file.name).color"
            />
          </div>
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <h2 class="text-sm font-bold text-white truncate max-w-md" :title="file?.name">
                {{ file?.name }}
              </h2>
              <UBadge
                v-if="file?.is_chunked"
                label="Chunked Store"
                color="primary"
                variant="subtle"
                size="xs"
                class="rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[9px] px-1.5 py-0"
              />
            </div>
            <div class="flex items-center gap-2 text-[11px] text-zinc-400 font-mono mt-0.5">
              <span>{{ file?.virtual_path }}</span>
              <span>•</span>
              <span class="text-emerald-400">{{ file ? formatBytes(file.size_bytes) : '' }}</span>
            </div>
          </div>
        </div>

        <!-- Right Header: Tabs & Close -->
        <div class="flex items-center gap-3 shrink-0">
          <!-- View Tab Switcher -->
          <div class="flex items-center rounded-xl border border-white/[0.08] bg-[#16161b] p-1 shadow-xs">
            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold transition-colors cursor-pointer"
              :class="[
                activeTab === 'preview'
                  ? 'bg-emerald-600 text-white shadow-xs'
                  : 'text-zinc-400 hover:text-white'
              ]"
              @click="activeTab = 'preview'"
            >
              <UIcon name="i-lucide-eye" class="size-3.5" />
              <span>Preview</span>
            </button>

            <button
              type="button"
              class="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-semibold transition-colors cursor-pointer"
              :class="[
                activeTab === 'metadata'
                  ? 'bg-emerald-600 text-white shadow-xs'
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
            @click="emit('update:open', false)"
          >
            <UIcon name="i-lucide-x" class="size-4" />
          </button>
        </div>
      </div>
    </template>

    <!-- Modal Body -->
    <template #body>
      <div v-if="file" class="flex-1 overflow-y-auto flex flex-col min-h-0 bg-[#0c0c0e]">
        <!-- TAB 1: INTERACTIVE CONTENT PREVIEW -->
        <div v-if="activeTab === 'preview'" class="flex-1 flex flex-col min-h-0 p-5">
          <!-- 1. IMAGE VIEWER -->
          <div v-if="fileType === 'image'" class="flex-1 flex flex-col items-center justify-center space-y-4">
            <div class="flex items-center justify-end w-full max-w-lg px-4 py-2 rounded-2xl bg-[#121215] border border-white/[0.07] text-xs">
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  title="Zoom out"
                  class="size-7 flex items-center justify-center rounded-lg hover:bg-white/[0.08] text-zinc-400 hover:text-white cursor-pointer"
                  @click="zoomOut"
                >
                  <UIcon name="i-lucide-zoom-out" class="size-3.5" />
                </button>
                <span class="text-[11px] font-mono text-emerald-400 px-2 min-w-12 text-center">{{ zoomLevel }}%</span>
                <button
                  type="button"
                  title="Zoom in"
                  class="size-7 flex items-center justify-center rounded-lg hover:bg-white/[0.08] text-zinc-400 hover:text-white cursor-pointer"
                  @click="zoomIn"
                >
                  <UIcon name="i-lucide-zoom-in" class="size-3.5" />
                </button>
                <button
                  type="button"
                  class="px-2 py-1 rounded-lg hover:bg-white/[0.08] text-zinc-400 hover:text-white text-[10px] font-mono ml-1 cursor-pointer"
                  @click="resetZoom"
                >
                  Reset
                </button>
              </div>
            </div>

            <div class="relative w-full max-h-[55vh] flex items-center justify-center overflow-hidden rounded-2xl border border-white/[0.08] bg-[#09090b] p-4">
              <img
                :src="inlineUrl"
                :alt="file.name"
                class="max-h-[50vh] max-w-full object-contain rounded-xl transition-transform duration-200 shadow-2xl"
                :style="{ transform: `scale(${zoomLevel / 100})` }"
              >
            </div>
          </div>

          <!-- 2. VIDEO PLAYER -->
          <div v-else-if="fileType === 'video'" class="flex-1 flex flex-col items-center justify-center space-y-4">
            <div class="w-full max-w-2xl rounded-2xl overflow-hidden border border-white/[0.08] bg-black shadow-2xl">
              <video controls class="w-full max-h-[55vh] object-contain" :src="inlineUrl">
                Your browser does not support HTML5 video playback.
              </video>
            </div>

            <div class="flex items-center justify-between w-full max-w-2xl px-4 py-2.5 rounded-2xl bg-[#121215] border border-white/[0.07] text-xs">
              <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[10px] font-semibold">
                <span class="size-1.5 rounded-full bg-emerald-400 animate-pulse" />
                Streaming directly from provider
              </span>
              <span class="text-zinc-400 font-mono text-[11px]">{{ formatBytes(file.size_bytes) }}</span>
            </div>
          </div>

          <!-- 3. AUDIO PLAYER -->
          <div v-else-if="fileType === 'audio'" class="flex-1 flex flex-col items-center justify-center p-8 space-y-6">
            <div class="w-full max-w-md p-6 rounded-3xl bg-[#121215] border border-white/[0.08] shadow-2xl space-y-5 text-center">
              <div class="flex size-20 items-center justify-center rounded-3xl bg-emerald-500/15 text-emerald-400 border border-emerald-500/25 mx-auto shadow-lg shadow-emerald-500/10">
                <UIcon name="i-lucide-music" class="size-10" />
              </div>

              <div>
                <h3 class="font-bold text-sm text-white truncate">{{ file.name }}</h3>
                <p class="text-xs text-zinc-400 mt-1 font-mono">{{ formatBytes(file.size_bytes) }}</p>
              </div>

              <audio controls class="w-full" :src="inlineUrl">
                Your browser does not support HTML5 audio playback.
              </audio>
            </div>
          </div>

          <!-- 4. PDF DOCUMENT VIEWER -->
          <div v-else-if="fileType === 'pdf'" class="flex-1 flex flex-col space-y-3">
            <div class="flex items-center justify-between p-2.5 rounded-2xl bg-[#121215] border border-white/[0.07] text-xs">
              <span class="text-zinc-400 font-mono text-[11px]">Rendered by your browser's built-in PDF viewer</span>
              <span class="text-zinc-400 font-mono text-[11px]">{{ formatBytes(file.size_bytes) }}</span>
            </div>

            <iframe
              :src="inlineUrl"
              :title="file.name"
              class="flex-1 min-h-[55vh] w-full rounded-2xl border border-white/[0.08] bg-white"
            />
          </div>

          <!-- 5. TEXT & CONFIG FILES -->
          <div v-else-if="fileType === 'text'" class="flex-1 flex flex-col space-y-2.5">
            <div class="flex items-center justify-between px-4 py-2 rounded-2xl bg-[#121215] border border-white/[0.07] text-xs">
              <span class="font-mono text-emerald-400 text-[11px]">{{ file.name }}</span>
              <button
                v-if="textContent"
                type="button"
                class="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg bg-[#18181d] hover:bg-[#202028] text-zinc-300 hover:text-white transition-colors cursor-pointer text-xs font-medium"
                @click="copySnippet"
              >
                <UIcon :name="isCopied ? 'i-lucide-check' : 'i-lucide-copy'" class="size-3.5" :class="isCopied ? 'text-emerald-400' : ''" />
                <span>{{ isCopied ? 'Copied' : 'Copy' }}</span>
              </button>
            </div>

            <!-- File teks raksasa tak dimuat ke browser; unduh saja. -->
            <div
              v-if="isTextTooLarge"
              class="flex-1 flex flex-col items-center justify-center gap-3 rounded-2xl border border-white/[0.08] bg-[#09090b] p-8 text-center"
            >
              <UIcon name="i-lucide-file-text" class="size-10 text-zinc-500" />
              <p class="text-xs text-zinc-400 max-w-sm">
                Too large to preview ({{ formatBytes(file.size_bytes) }}).
                Download it to open the file.
              </p>
            </div>

            <div
              v-else-if="isLoadingText"
              class="flex-1 flex items-center justify-center rounded-2xl border border-white/[0.08] bg-[#09090b]"
            >
              <UIcon name="i-lucide-loader-2" class="size-6 text-emerald-400 animate-spin" />
            </div>

            <div
              v-else-if="textError"
              class="flex-1 flex flex-col items-center justify-center gap-2 rounded-2xl border border-red-500/20 bg-[#09090b] p-8 text-center"
            >
              <UIcon name="i-lucide-circle-alert" class="size-8 text-red-400" />
              <p class="text-xs text-red-400">{{ textError }}</p>
            </div>

            <div v-else class="flex-1 overflow-auto rounded-2xl border border-white/[0.08] bg-[#09090b] font-mono text-xs p-4 shadow-inner">
              <table class="w-full border-collapse">
                <tbody>
                  <tr v-for="(line, idx) in codeLines" :key="idx" class="hover:bg-white/[0.03] transition-colors leading-6">
                    <td class="w-10 select-none text-right pr-4 text-zinc-600 text-[11px] align-top">{{ idx + 1 }}</td>
                    <td class="text-zinc-300 whitespace-pre font-mono">{{ line }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Format yang tak bisa dirender browser (dokumen, arsip, spreadsheet) -->
          <div v-else class="flex-1 flex flex-col items-center justify-center p-8 space-y-4 text-center">
            <div class="p-5 rounded-3xl bg-[#15151a] border border-white/[0.08] text-emerald-400">
              <UIcon :name="getFileIcon(file.mime, file.name).icon" class="size-12" />
            </div>
            <div>
              <h3 class="font-bold text-base text-white">{{ file.name }}</h3>
              <p class="text-xs text-zinc-400 max-w-sm mt-1">
                This format cannot be rendered in the browser. Download it to open the file
                with an app on your device.
              </p>
            </div>
            <button
              type="button"
              class="px-4 py-2 rounded-xl text-xs font-bold bg-emerald-600 hover:bg-emerald-500 text-white flex items-center gap-2 cursor-pointer shadow-xs"
              @click="emit('download', file)"
            >
              <UIcon name="i-lucide-download" class="size-4" />
              <span>Download ({{ formatBytes(file.size_bytes) }})</span>
            </button>
          </div>
        </div>

        <!-- TAB 2: METADATA & TELEMETRY -->
        <div v-else class="flex-1 overflow-y-auto p-6 space-y-4 text-xs">
          <!-- Metadata Grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Card 1: Virtual Filesystem Details -->
            <div class="p-4 rounded-2xl bg-[#121215] border border-white/[0.07] space-y-3">
              <h3 class="font-bold text-zinc-200 flex items-center gap-2 text-xs">
                <UIcon name="i-lucide-folder-tree" class="size-4 text-emerald-400" />
                Virtual Filesystem (VFS)
              </h3>
              <div class="space-y-2 text-xs">
                <div class="flex justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">Virtual Path</span>
                  <span class="font-mono text-emerald-400 font-semibold">{{ file.virtual_path }}</span>
                </div>
                <div class="flex justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">File ID</span>
                  <span class="font-mono text-zinc-300">{{ file.id }}</span>
                </div>
                <div class="flex justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">File Size</span>
                  <span class="font-mono text-zinc-200 font-bold">{{ formatBytes(file.size_bytes) }} ({{ file.size_bytes.toLocaleString() }} B)</span>
                </div>
                <div class="flex justify-between py-1">
                  <span class="text-zinc-400">MIME Type</span>
                  <span class="font-mono text-zinc-300">{{ file.mime || 'application/octet-stream' }}</span>
                </div>
              </div>
            </div>

            <!-- Card 2: Cloud Storage Placement -->
            <div class="p-4 rounded-2xl bg-[#121215] border border-white/[0.07] space-y-3">
              <h3 class="font-bold text-zinc-200 flex items-center gap-2 text-xs">
                <UIcon name="i-lucide-cloud" class="size-4 text-emerald-400" />
                Physical Remote Storage
              </h3>
              <div class="space-y-2 text-xs">
                <div class="flex justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">Cloud Provider</span>
                  <span class="font-semibold text-zinc-200 flex items-center gap-1.5">
                    <UIcon :name="getProviderMeta(file.provider).icon" class="size-3.5" />
                    {{ file.account_label }}
                  </span>
                </div>
                <div class="flex justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">Storage Architecture</span>
                  <span class="font-mono text-emerald-400">
                    {{ file.is_chunked ? 'Model B (Chunked Blocks)' : 'Model A (Whole-File)' }}
                  </span>
                </div>
                <div class="flex justify-between py-1 border-b border-white/[0.06]">
                  <span class="text-zinc-400">rclone Daemon Remote</span>
                  <span class="font-mono text-zinc-300">acc_{{ file.account_id.replace('acc-', '') }}:</span>
                </div>
                <div class="flex justify-between py-1">
                  <span class="text-zinc-400">Last Modified</span>
                  <span class="text-zinc-300">{{ formatDate(file.modified_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Modal Footer Actions -->
    <template #footer>
      <div v-if="file" class="flex items-center justify-between w-full">
        <!-- Delete Button -->
        <button
          type="button"
          class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 transition-colors cursor-pointer"
          @click="emit('delete', file); emit('update:open', false)"
        >
          <UIcon name="i-lucide-trash-2" class="size-4" />
          <span>Delete</span>
        </button>

        <!-- Right Buttons: Move, Migrate, Download -->
        <div class="flex items-center gap-2">
          <!-- Rename: emit ini sudah dideklarasi dan didengarkan explorer, tapi
               sebelumnya tak pernah punya pemicu. -->
          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-[#16161b] hover:bg-[#1c1c22] text-zinc-300 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
            @click="emit('rename', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-edit-3" class="size-4 text-sky-400" />
            <span>Rename</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-[#16161b] hover:bg-[#1c1c22] text-zinc-300 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
            @click="emit('moveToFolder', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-folder-input" class="size-4 text-amber-400" />
            <span>Move Folder</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-[#16161b] hover:bg-[#1c1c22] text-zinc-300 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
            @click="emit('migrateProvider', file); emit('update:open', false)"
          >
            <UIcon name="i-lucide-arrow-right-left" class="size-4 text-emerald-400" />
            <span>Migrate Cloud</span>
          </button>

          <button
            type="button"
            class="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-bold bg-emerald-600 hover:bg-emerald-500 text-white transition-all cursor-pointer shadow-xs"
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
