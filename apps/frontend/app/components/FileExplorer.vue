<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { FileEntry, FolderEntry } from '~/types'

const filesStore = useFilesStore()
const accountsStore = useAccountsStore()
const { formatBytes, formatDate, getProviderMeta, getFileIcon } = useFormatters()
const toast = useToast()

// Modal & Selection States
const selectedFile = ref<FileEntry | null>(null)
const previewFile = ref<FileEntry | null>(null)
const isPreviewModalOpen = ref(false)

// Cloud Migration Modal (Physical Remote Transfer)
const isMoveModalOpen = ref(false)
const targetMoveAccount = ref<string>('')

// VFS Virtual Organization Modals
const newFolderName = ref('')
const newFolderError = ref('')

// Dibagi lewat store: navbar halaman punya tombolnya sendiri.
const isNewFolderModalOpen = computed({
  get: () => filesStore.isNewFolderModalOpen,
  set: (v: boolean) => { filesStore.isNewFolderModalOpen = v }
})

const isRenameFolderModalOpen = ref(false)
const folderToRename = ref<FolderEntry | null>(null)
const renamedFolderName = ref('')

const isDeleteFolderModalOpen = ref(false)
const folderToDelete = ref<FolderEntry | null>(null)

const isMoveToFolderModalOpen = ref(false)
const fileToMoveVfs = ref<FileEntry | null>(null)
const selectedTargetVfsFolderId = ref<string | null>(null)

const isRenameFileModalOpen = ref(false)
const fileToRename = ref<FileEntry | null>(null)
const renamedFileName = ref('')

const isDeleteFileModalOpen = ref(false)
const fileToDelete = ref<FileEntry | null>(null)

// VFS: pindahkan folder ke induk lain
const isMoveFolderModalOpen = ref(false)
const folderToMove = ref<FolderEntry | null>(null)
const targetParentFolderId = ref<string | null>(null)

// Aksi massal atas file terpilih
const isBulkDeleteOpen = ref(false)
const isBulkMoveOpen = ref(false)
const isBulkMigrateOpen = ref(false)
const bulkTargetFolderId = ref<string | null>(null)
const bulkTargetAccountId = ref<string>('')

// Status bersama untuk operasi yang menunggu backend.
const isBusy = ref(false)
const actionError = ref('')

// Jangkar shift-click: file terakhir yang dipilih lewat klik biasa.
const lastClickedId = ref<string | null>(null)

// Seret-lepas di dalam explorer: pindahkan file/folder dengan menjatuhkannya
// ke folder lain, atau ke breadcrumb untuk naik ke induk.
const dnd = useExplorerDnd()

const categories = [
  { id: 'all', label: 'All Items', icon: 'i-lucide-layers' },
  { id: 'docs', label: 'Documents', icon: 'i-lucide-file-text' },
  { id: 'media', label: 'Photos & Videos', icon: 'i-lucide-film' },
  { id: 'archives', label: 'Archives & Zip', icon: 'i-lucide-file-archive' },
  { id: 'code', label: 'Code & DB', icon: 'i-lucide-file-code' }
]

function getFileMenuItems(file: FileEntry): DropdownMenuItem[][] {
  return [
    [
      {
        label: 'Preview & Details',
        icon: 'i-lucide-eye',
        onSelect: () => openPreview(file)
      },
      {
        label: 'Download File',
        icon: 'i-lucide-download',
        onSelect: () => handleDownload(file)
      },
      {
        label: 'Rename File',
        icon: 'i-lucide-edit-3',
        onSelect: () => openRenameFile(file)
      }
    ],
    [
      {
        label: 'Move to Virtual Folder...',
        icon: 'i-lucide-folder-input',
        onSelect: () => openMoveToFolder(file)
      },
      {
        label: 'Migrate Cloud Provider...',
        icon: 'i-lucide-arrow-right-left',
        onSelect: () => openMoveModal(file)
      },
      {
        label: 'Copy Virtual Path',
        icon: 'i-lucide-copy',
        onSelect: () => navigator.clipboard?.writeText(file.virtual_path)
      }
    ],
    [
      {
        label: 'Delete from Cloud',
        icon: 'i-lucide-trash-2',
        color: 'error',
        onSelect: () => openDeleteFile(file)
      }
    ]
  ]
}

function getFolderMenuItems(folder: FolderEntry): DropdownMenuItem[][] {
  return [
    [
      {
        label: 'Open Folder',
        icon: 'i-lucide-folder-open',
        onSelect: () => { void filesStore.navigateToFolder(folder.id) }
      },
      {
        label: 'Rename Folder',
        icon: 'i-lucide-edit-2',
        onSelect: () => openRenameFolder(folder)
      },
      {
        label: 'Move Folder...',
        icon: 'i-lucide-folder-input',
        onSelect: () => openMoveFolder(folder)
      },
      {
        label: 'Copy Folder Path',
        icon: 'i-lucide-copy',
        onSelect: () => navigator.clipboard?.writeText(folder.path)
      }
    ],
    [
      {
        label: 'Delete Folder',
        icon: 'i-lucide-trash-2',
        color: 'error',
        onSelect: () => openDeleteFolder(folder)
      }
    ]
  ]
}

// Unduhan diserahkan ke browser: backend menstream langsung dari provider
// (stream-through), jadi file besar tak perlu lewat memori JS.
function handleDownload(file: FileEntry) {
  const a = document.createElement('a')
  a.href = filesStore.downloadUrl(file.id)
  a.download = file.name
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function openPreview(file: FileEntry) {
  previewFile.value = file
  isPreviewModalOpen.value = true
}

// Physical Migration Between Cloud Providers (POST /files/{id}/move)
function openMoveModal(file: FileEntry) {
  selectedFile.value = file
  const otherAccs = accountsStore.accounts.filter(a => a.id !== file.account_id)
  targetMoveAccount.value = otherAccs[0]?.id || ''
  isMoveModalOpen.value = true
}

async function confirmMove() {
  if (!selectedFile.value || !targetMoveAccount.value) return
  isBusy.value = true
  actionError.value = ''
  try {
    // Transfer data nyata antar provider — beda dari pindah folder virtual.
    await filesStore.moveFile(selectedFile.value.id, targetMoveAccount.value)
    isMoveModalOpen.value = false
    toast.add({ title: 'File migrated to another account', color: 'success' })
  } catch (err) {
    actionError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

const moveOptions = computed(() => {
  if (!selectedFile.value) return []
  return accountsStore.activeAccounts
    .filter(a => a.id !== selectedFile.value?.account_id)
    .map(a => ({
      value: a.id,
      label: `${a.label} (${formatBytes(a.free_bytes)} free)`
    }))
})

// VFS: buat folder virtual baru
function openNewFolder() {
  newFolderName.value = ''
  newFolderError.value = ''
  isNewFolderModalOpen.value = true
}

async function confirmCreateFolder() {
  if (!newFolderName.value.trim()) {
    newFolderError.value = 'Enter a folder name.'
    return
  }
  isBusy.value = true
  newFolderError.value = ''
  try {
    await filesStore.createFolder(newFolderName.value)
    isNewFolderModalOpen.value = false
  } catch (err) {
    newFolderError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

// VFS: rename folder — path seluruh turunan ikut dihitung ulang di backend
function openRenameFolder(folder: FolderEntry) {
  folderToRename.value = folder
  renamedFolderName.value = folder.name
  actionError.value = ''
  isRenameFolderModalOpen.value = true
}

async function confirmRenameFolder() {
  if (!folderToRename.value || !renamedFolderName.value.trim()) return
  isBusy.value = true
  actionError.value = ''
  try {
    await filesStore.renameFolder(folderToRename.value.id, renamedFolderName.value)
    isRenameFolderModalOpen.value = false
  } catch (err) {
    actionError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

// VFS: hapus folder
function openDeleteFolder(folder: FolderEntry) {
  folderToDelete.value = folder
  actionError.value = ''
  isDeleteFolderModalOpen.value = true
}

async function confirmDeleteFolder(recursive: boolean) {
  if (!folderToDelete.value) return
  isBusy.value = true
  actionError.value = ''
  try {
    await filesStore.deleteFolder(folderToDelete.value.id, recursive)
    isDeleteFolderModalOpen.value = false
    toast.add({ title: 'Folder deleted', color: 'success' })
  } catch (err) {
    actionError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

// VFS: pindahkan file ke folder lain (DB murni, tak menyentuh provider)
function openMoveToFolder(file: FileEntry) {
  fileToMoveVfs.value = file
  selectedTargetVfsFolderId.value = file.folder_id || null
  actionError.value = ''
  isMoveToFolderModalOpen.value = true
}

async function confirmMoveToFolder() {
  if (!fileToMoveVfs.value) return
  isBusy.value = true
  actionError.value = ''
  try {
    await filesStore.moveFileToFolder(fileToMoveVfs.value.id, selectedTargetVfsFolderId.value)
    isMoveToFolderModalOpen.value = false
  } catch (err) {
    actionError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

// VFS: rename file
function openRenameFile(file: FileEntry) {
  fileToRename.value = file
  renamedFileName.value = file.name
  actionError.value = ''
  isRenameFileModalOpen.value = true
}

async function confirmRenameFile() {
  if (!fileToRename.value || !renamedFileName.value.trim()) return
  isBusy.value = true
  actionError.value = ''
  try {
    await filesStore.renameFile(fileToRename.value.id, renamedFileName.value)
    isRenameFileModalOpen.value = false
  } catch (err) {
    actionError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

// Hapus file: objek fisik di provider ikut terhapus
function openDeleteFile(file: FileEntry) {
  fileToDelete.value = file
  actionError.value = ''
  isDeleteFileModalOpen.value = true
}

async function confirmDeleteFile() {
  if (!fileToDelete.value) return
  isBusy.value = true
  actionError.value = ''
  try {
    await filesStore.deleteFile(fileToDelete.value.id)
    isDeleteFileModalOpen.value = false
    toast.add({ title: 'File deleted', color: 'success' })
  } catch (err) {
    actionError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

// VFS: pindahkan folder ke induk lain
function openMoveFolder(folder: FolderEntry) {
  folderToMove.value = folder
  targetParentFolderId.value = folder.parent_id
  actionError.value = ''
  isMoveFolderModalOpen.value = true
}

/**
 * Tujuan yang sah bagi folder yang sedang dipindah: bukan dirinya sendiri, bukan
 * turunannya (itu akan memutus pohon), dan bukan induknya yang sekarang.
 */
const moveFolderOptions = computed(() => {
  const moving = folderToMove.value
  if (!moving) return []
  const prefix = moving.path.endsWith('/') ? moving.path : `${moving.path}/`
  return filesStore.allFoldersHierarchical.filter(item =>
    item.id !== moving.id
    && item.id !== moving.parent_id
    && !item.path.startsWith(prefix))
})

async function confirmMoveFolder() {
  if (!folderToMove.value) return
  isBusy.value = true
  actionError.value = ''
  try {
    await filesStore.moveFolder(folderToMove.value.id, targetParentFolderId.value)
    isMoveFolderModalOpen.value = false
    toast.add({ title: 'Folder moved', color: 'success' })
  } catch (err) {
    actionError.value = friendlyMessage(err)
  } finally {
    isBusy.value = false
  }
}

// ---- Pilihan & aksi massal ----

/**
 * Klik pada baris file: shift memperluas pilihan dari jangkar terakhir, meta/ctrl
 * menambah satu per satu, klik biasa membuka pratinjau seperti sebelumnya.
 */
function handleRowClick(file: FileEntry, event: MouseEvent) {
  if (event.shiftKey && lastClickedId.value) {
    filesStore.selectRange(lastClickedId.value, file.id)
    return
  }
  if (event.metaKey || event.ctrlKey) {
    filesStore.toggleSelection(file.id)
    lastClickedId.value = file.id
    return
  }
  openPreview(file)
}

function handleCheckboxClick(file: FileEntry, event: MouseEvent) {
  if (event.shiftKey && lastClickedId.value) {
    filesStore.selectRange(lastClickedId.value, file.id)
  } else {
    filesStore.toggleSelection(file.id)
  }
  lastClickedId.value = file.id
}

/** Laporkan hasil satu aksi massal: sebagian gagal tetap harus terlihat. */
function reportBulk(
  res: { ok: number; failed: { id: string; error: unknown }[] },
  verb: string
) {
  if (res.failed.length === 0) {
    toast.add({ title: `${res.ok} file${res.ok === 1 ? '' : 's'} ${verb}`, color: 'success' })
    return
  }
  toast.add({
    title: `${res.failed.length} of ${res.ok + res.failed.length} could not be ${verb}`,
    description: friendlyMessage(res.failed[0]!.error),
    color: res.ok > 0 ? 'warning' : 'error'
  })
}

function openBulkMove() {
  bulkTargetFolderId.value = filesStore.currentFolderId
  isBulkMoveOpen.value = true
}

function openBulkMigrate() {
  bulkTargetAccountId.value = accountsStore.activeAccounts[0]?.id || ''
  isBulkMigrateOpen.value = true
}

async function confirmBulkDelete() {
  isBusy.value = true
  try {
    reportBulk(await filesStore.deleteSelected(), 'deleted')
    isBulkDeleteOpen.value = false
  } finally {
    isBusy.value = false
  }
}

async function confirmBulkMove() {
  isBusy.value = true
  try {
    reportBulk(await filesStore.moveSelectedToFolder(bulkTargetFolderId.value), 'moved')
    isBulkMoveOpen.value = false
  } finally {
    isBusy.value = false
  }
}

async function confirmBulkMigrate() {
  if (!bulkTargetAccountId.value) return
  isBusy.value = true
  try {
    reportBulk(await filesStore.migrateSelected(bulkTargetAccountId.value), 'migrated')
    isBulkMigrateOpen.value = false
  } finally {
    isBusy.value = false
  }
}

function downloadSelected() {
  for (const file of filesStore.selectedFiles) handleDownload(file)
  filesStore.clearSelection()
}

const bulkMigrateOptions = computed(() =>
  accountsStore.activeAccounts.map(a => ({
    value: a.id,
    label: `${a.label} (${formatBytes(a.free_bytes)} free)`
  })))

// ---- Urutan & halaman ----

/** Klik header kolom: kolom sama membalik arah, kolom lain mulai menaik. */
function sortByColumn(column: 'name' | 'size' | 'modified') {
  if (filesStore.sortBy === column) {
    filesStore.sortDirection = filesStore.sortDirection === 'asc' ? 'desc' : 'asc'
  } else {
    filesStore.sortBy = column
    filesStore.sortDirection = 'asc'
  }
}

function sortIcon(column: 'name' | 'size' | 'modified') {
  if (filesStore.sortBy !== column) return 'i-lucide-chevrons-up-down'
  return filesStore.sortDirection === 'asc' ? 'i-lucide-arrow-up' : 'i-lucide-arrow-down'
}

// Penyaring & urutan dikerjakan backend, jadi tiap perubahan memuat ulang dari
// halaman pertama — bukan menyaring ulang daftar yang kebetulan sedang tampil.
watch(
  () => [filesStore.selectedAccountId, filesStore.selectedCategory,
         filesStore.sortBy, filesStore.sortDirection],
  () => { void filesStore.applyFilters() }
)

async function retryLoad() {
  await filesStore.loadAll().catch(() => null)
}

/** Daftar kosong = tak ada folder maupun file yang bisa dirender. */
const isListEmpty = computed(() =>
  filesStore.currentFolders.length === 0 && filesStore.filteredFiles.length === 0)

/** Kosong karena penyaring, bukan karena folder memang kosong — beda pesannya. */
const hasActiveFilter = computed(() =>
  filesStore.searchQuery.trim() !== ''
  || filesStore.selectedCategory !== 'all'
  || filesStore.selectedAccountId !== 'all')

function clearFilters() {
  filesStore.searchQuery = ''
  filesStore.selectedCategory = 'all'
  filesStore.selectedAccountId = 'all'
  void filesStore.applyFilters()
}

// Pencarian dijalankan backend atas index DB; debounce menahan permintaan
// sampai user berhenti mengetik.
let searchTimer: ReturnType<typeof setTimeout> | undefined
watch(() => filesStore.searchQuery, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    // Kata kunci baru berarti kumpulan hasil baru — mulai dari halaman pertama.
    void filesStore.applyFilters()
  }, 300)
})

onBeforeUnmount(() => clearTimeout(searchTimer))
</script>

<template>
  <div class="space-y-4">
    <!-- VFS Breadcrumbs & Location Bar -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-2xl bg-[#0d111a] border border-white/[0.08] shadow-xs">
      <!-- Breadcrumb Path Trail -->
      <div class="flex items-center gap-1.5 overflow-x-auto py-1 text-xs">
        <!-- Up one level button (if not in root) -->
        <button
          v-if="filesStore.currentFolderId !== null"
          type="button"
          title="Go up one folder"
          class="flex size-7 items-center justify-center rounded-lg bg-[#1b2130] hover:bg-[#1a1a20] text-zinc-400 hover:text-white border border-white/[0.08] transition-colors cursor-pointer shrink-0 mr-1"
          @click="filesStore.navigateUp()"
        >
          <UIcon name="i-lucide-corner-left-up" class="size-3.5" />
        </button>

        <div class="flex items-center gap-1 font-mono">
          <template v-for="(crumb, idx) in filesStore.breadcrumbs" :key="crumb.path">
            <!-- Breadcrumb juga menerima jatuhan: satu-satunya cara memindahkan
                 item ke folder INDUK tanpa menutup folder yang sedang dibuka. -->
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 rounded-lg px-2 py-1 transition-colors"
              :class="[
                dnd.isDropTarget(crumb.id)
                  ? 'bg-primary-500/25 text-primary-200 font-semibold ring-1 ring-primary-400/60'
                  : idx === filesStore.breadcrumbs.length - 1
                    ? 'bg-primary-500/15 text-primary-300 font-semibold border border-primary-500/25'
                    : 'text-zinc-400 hover:text-white hover:bg-white/[0.05]'
              ]"
              @click="filesStore.navigateToFolder(crumb.id)"
              @dragover="dnd.onDragOverFolder($event, crumb.id)"
              @dragleave="dnd.onDragLeaveFolder(crumb.id)"
              @drop="dnd.dropInto($event, crumb.id)"
            >
              <UIcon
                :name="idx === 0 ? 'i-lucide-hard-drive' : 'i-lucide-folder'"
                class="size-3.5"
                :class="idx === filesStore.breadcrumbs.length - 1 ? 'text-primary-400' : 'text-zinc-500'"
              />
              <span>{{ crumb.name }}</span>
            </button>

            <span v-if="idx < filesStore.breadcrumbs.length - 1" class="text-zinc-600 select-none">/</span>
          </template>
        </div>
      </div>

      <!-- Quick Actions in Bar: New Folder & Summary Pill -->
      <div class="flex items-center gap-2 shrink-0">
        <span class="text-[11px] text-zinc-400 font-mono hidden md:inline">
          {{ filesStore.currentFolders.length }} folders, {{ filesStore.totalFiles }} files
        </span>

        <button
          type="button"
          class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-[#1c2231] hover:bg-[#262d3e] text-zinc-200 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
          @click="openNewFolder"
        >
          <UIcon name="i-lucide-folder-plus" class="size-3.5 text-primary-400" />
          <span>New Folder</span>
        </button>
      </div>
    </div>

    <!-- Toolbar: Search, Category Filters, Provider Filter, View Toggle -->
    <div class="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
      <!-- Search Bar -->
      <div class="relative flex-1 max-w-md">
        <UInput
          v-model="filesStore.searchQuery"
          icon="i-lucide-search"
          placeholder="Search items, paths, or providers..."
          class="w-full rounded-xl"
          size="md"
        />
        <button
          v-if="filesStore.searchQuery"
          type="button"
          class="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-white"
          @click="filesStore.searchQuery = ''"
        >
          <UIcon name="i-lucide-x" class="size-3.5" />
        </button>
      </div>

      <!-- Filters & Actions -->
      <div class="flex flex-wrap items-center gap-2">
        <!-- Account Provider Filter -->
        <USelect
          v-model="filesStore.selectedAccountId"
          :items="[
            { value: 'all', label: 'All Cloud Accounts' },
            ...accountsStore.accounts.map(a => ({ value: a.id, label: `${a.label}` }))
          ]"
          icon="i-lucide-cloud"
          class="min-w-44 rounded-xl bg-[#151a27] border border-white/[0.08] text-zinc-200 shadow-xs"
          size="md"
        />

        <!-- Sort By -->
        <USelect
          v-model="filesStore.sortBy"
          :items="[
            { value: 'name', label: 'Name' },
            { value: 'modified', label: 'Date Modified' },
            { value: 'size', label: 'File Size' }
          ]"
          icon="i-lucide-arrow-up-down"
          class="w-36 rounded-xl bg-[#151a27] border border-white/[0.08] text-zinc-200 shadow-xs"
          size="md"
        />

        <!-- View Mode Toggle -->
        <div class="flex items-center rounded-xl border border-white/[0.08] bg-[#151a27] p-1 shadow-xs">
          <UButton
            icon="i-lucide-list"
            size="xs"
            square
            class="rounded-lg"
            :color="filesStore.viewMode === 'table' ? 'primary' : 'neutral'"
            :variant="filesStore.viewMode === 'table' ? 'solid' : 'ghost'"
            @click="filesStore.viewMode = 'table'"
          />
          <UButton
            icon="i-lucide-layout-grid"
            size="xs"
            square
            class="rounded-lg"
            :color="filesStore.viewMode === 'grid' ? 'primary' : 'neutral'"
            :variant="filesStore.viewMode === 'grid' ? 'solid' : 'ghost'"
            @click="filesStore.viewMode = 'grid'"
          />
        </div>
      </div>
    </div>

    <!-- Category Filter Tabs -->
    <div class="flex items-center gap-2 overflow-x-auto pb-1">
      <button
        v-for="cat in categories"
        :key="cat.id"
        type="button"
        class="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-medium whitespace-nowrap transition-all duration-150 cursor-pointer shadow-xs"
        :class="[
          filesStore.selectedCategory === cat.id
            ? 'bg-primary-600 text-white shadow-xs font-semibold'
            : 'bg-[#171c2a] hover:bg-[#212736] text-zinc-400 hover:text-zinc-200 border border-white/[0.06]'
        ]"
        @click="filesStore.selectedCategory = cat.id"
      >
        <UIcon :name="cat.icon" class="size-3.5" />
        <span>{{ cat.label }}</span>
      </button>

      <div class="ml-auto hidden sm:flex items-center gap-2 text-xs text-zinc-500 font-mono">
        <span v-if="filesStore.searchQuery">Search results across all virtual folders</span>
        <span v-else>Active path: <strong class="text-zinc-300 font-medium">{{ filesStore.currentPath }}</strong></span>
      </div>
    </div>

    <!-- Bulk action bar: muncul hanya saat ada yang dipilih -->
    <div
      v-if="filesStore.selectedFileIds.length > 0"
      class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-2xl border border-primary-500/30 bg-primary-500/[0.06] shadow-xs"
    >
      <div class="flex items-center gap-3 text-xs">
        <span class="inline-flex items-center gap-2 px-2.5 py-1 rounded-xl bg-primary-500/15 text-primary-300 border border-primary-500/25 font-semibold">
          <UIcon name="i-lucide-check-square" class="size-3.5" />
          {{ filesStore.selectedFileIds.length }} selected
        </span>
        <span class="text-zinc-400 font-mono">{{ formatBytes(filesStore.selectedBytes) }}</span>
        <button
          type="button"
          class="text-zinc-400 hover:text-white underline underline-offset-4 cursor-pointer"
          @click="filesStore.clearSelection()"
        >
          Clear
        </button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <UButton
          label="Download"
          icon="i-lucide-download"
          color="neutral"
          variant="subtle"
          size="xs"
          class="rounded-xl font-semibold"
          @click="downloadSelected"
        />
        <UButton
          label="Move to Folder"
          icon="i-lucide-folder-input"
          color="neutral"
          variant="subtle"
          size="xs"
          class="rounded-xl font-semibold"
          @click="openBulkMove"
        />
        <UButton
          label="Migrate"
          icon="i-lucide-arrow-right-left"
          color="neutral"
          variant="subtle"
          size="xs"
          class="rounded-xl font-semibold"
          :disabled="accountsStore.activeAccounts.length === 0"
          @click="openBulkMigrate"
        />
        <UButton
          label="Delete"
          icon="i-lucide-trash-2"
          color="error"
          variant="subtle"
          size="xs"
          class="rounded-xl font-semibold"
          @click="isBulkDeleteOpen = true"
        />
      </div>
    </div>

    <!-- Files View: Table Mode -->
    <div
      v-if="filesStore.viewMode === 'table'"
      class="rounded-3xl border border-white/[0.08] bg-[#0d111a] overflow-hidden shadow-sm"
    >
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse text-xs">
          <thead>
            <tr class="border-b border-white/[0.06] bg-[#141925] text-zinc-400 font-semibold uppercase tracking-wider text-[10px]">
              <th class="py-3.5 pl-5 pr-1 w-9">
                <UCheckbox
                  :model-value="filesStore.allVisibleSelected"
                  :indeterminate="filesStore.selectedFileIds.length > 0 && !filesStore.allVisibleSelected"
                  aria-label="Select all files on this page"
                  :disabled="filesStore.filteredFiles.length === 0"
                  @update:model-value="filesStore.toggleSelectAll()"
                />
              </th>
              <th class="py-3.5 px-4">
                <button type="button" class="inline-flex items-center gap-1.5 uppercase tracking-wider hover:text-white transition-colors cursor-pointer" @click="sortByColumn('name')">
                  Name &amp; Virtual Path
                  <UIcon :name="sortIcon('name')" class="size-3" :class="filesStore.sortBy === 'name' ? 'text-primary-400' : 'text-zinc-600'" />
                </button>
              </th>
              <th class="py-3.5 px-4">Storage Provider</th>
              <th class="py-3.5 px-4">
                <button type="button" class="inline-flex items-center gap-1.5 uppercase tracking-wider hover:text-white transition-colors cursor-pointer" @click="sortByColumn('size')">
                  Size
                  <UIcon :name="sortIcon('size')" class="size-3" :class="filesStore.sortBy === 'size' ? 'text-primary-400' : 'text-zinc-600'" />
                </button>
              </th>
              <th class="py-3.5 px-4">
                <button type="button" class="inline-flex items-center gap-1.5 uppercase tracking-wider hover:text-white transition-colors cursor-pointer" @click="sortByColumn('modified')">
                  Last Modified
                  <UIcon :name="sortIcon('modified')" class="size-3" :class="filesStore.sortBy === 'modified' ? 'text-primary-400' : 'text-zinc-600'" />
                </button>
              </th>
              <th class="py-3.5 px-5 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/[0.04]">
            <!-- Virtual Folders Rows (Displayed first) -->
            <tr
              v-for="folder in filesStore.currentFolders"
              :key="folder.id"
              draggable="true"
              class="group cursor-pointer transition-colors"
              :class="dnd.isDropTarget(folder.id)
                ? 'bg-primary-500/15 ring-1 ring-inset ring-primary-400/50'
                : 'hover:bg-[#1a1f2d]'"
              @click="filesStore.navigateToFolder(folder.id)"
              @dragstart="dnd.startFolderDrag($event, folder)"
              @dragend="dnd.endDrag()"
              @dragover="dnd.onDragOverFolder($event, folder.id)"
              @dragleave="dnd.onDragLeaveFolder(folder.id)"
              @drop="dnd.dropInto($event, folder.id)"
            >
              <!-- Folder tak ikut pilihan massal: aksinya beda dari file. -->
              <td class="py-3 pl-5 pr-1" />

              <!-- Folder Name & Icon -->
              <td class="py-3 px-4">
                <div class="flex items-center gap-3.5 min-w-[260px]">
                  <div class="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20 group-hover:scale-105 transition-transform shadow-xs">
                    <UIcon name="i-lucide-folder" class="size-5 fill-amber-500/20" />
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-semibold text-zinc-200 truncate group-hover:text-primary-400 transition-colors">
                        {{ folder.name }}
                      </span>
                      <UBadge
                        label="Virtual Folder"
                        color="warning"
                        variant="subtle"
                        size="xs"
                        class="text-[9px] px-1.5 py-0 rounded-md font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20"
                      />
                    </div>
                    <span class="text-[11px] text-zinc-500 font-mono truncate block mt-0.5">{{ folder.path }}</span>
                  </div>
                </div>
              </td>

              <!-- Organization Info -->
              <td class="py-3 px-4 whitespace-nowrap text-zinc-400 font-mono text-[11px]">
                <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-zinc-800/80 border border-white/[0.06] text-zinc-300">
                  <UIcon name="i-lucide-layers" class="size-3 text-zinc-400" />
                  {{ folder.subfolder_count }} subfolder
                </span>
              </td>

              <!-- Size -->
              <td class="py-3 px-4 whitespace-nowrap font-mono text-zinc-500">
                —
              </td>

              <!-- Date Created -->
              <td class="py-3 px-4 whitespace-nowrap text-zinc-400">
                {{ formatDate(folder.created_at) }}
              </td>

              <!-- Actions -->
              <td class="py-3 px-5 text-right whitespace-nowrap" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <UTooltip text="Open Folder">
                    <UButton
                      icon="i-lucide-folder-open"
                      color="neutral"
                      variant="ghost"
                      size="xs"
                      square
                      class="rounded-lg hover:text-primary-400"
                      @click="filesStore.navigateToFolder(folder.id)"
                    />
                  </UTooltip>

                  <UDropdownMenu :items="getFolderMenuItems(folder)" :content="{ align: 'end' }">
                    <UButton
                      icon="i-lucide-more-horizontal"
                      color="neutral"
                      variant="ghost"
                      size="xs"
                      square
                      class="rounded-lg"
                    />
                  </UDropdownMenu>
                </div>
              </td>
            </tr>

            <!-- Files Rows -->
            <tr
              v-for="file in filesStore.filteredFiles"
              :key="file.id"
              draggable="true"
              class="group cursor-pointer transition-colors"
              :class="filesStore.isSelected(file.id) ? 'bg-primary-500/[0.07]' : 'hover:bg-[#1a1f2d]'"
              @click="handleRowClick(file, $event)"
              @dragstart="dnd.startFileDrag($event, file)"
              @dragend="dnd.endDrag()"
            >
              <td class="py-3 pl-5 pr-1" @click.stop>
                <UCheckbox
                  :model-value="filesStore.isSelected(file.id)"
                  :aria-label="`Select ${file.name}`"
                  @click="handleCheckboxClick(file, $event)"
                />
              </td>

              <!-- Name & Icon -->
              <td class="py-3 px-4">
                <div class="flex items-center gap-3.5 min-w-[260px]">
                  <div class="p-2 rounded-xl bg-[#1b2130] shrink-0 border border-white/[0.06] group-hover:scale-105 transition-transform shadow-xs">
                    <UIcon
                      :name="getFileIcon(file.mime, file.name).icon"
                      class="size-5"
                      :class="getFileIcon(file.mime, file.name).color"
                    />
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-semibold text-zinc-200 truncate max-w-xs md:max-w-md group-hover:text-primary-400 transition-colors">
                        {{ file.name }}
                      </span>
                      <UBadge
                        v-if="file.is_chunked"
                        label="Chunked"
                        color="primary"
                        variant="subtle"
                        size="xs"
                        class="text-[9px] px-1.5 py-0 rounded-md font-medium bg-primary-500/10 text-primary-400 border border-primary-500/20"
                      />
                    </div>
                    <span class="text-[11px] text-zinc-500 font-mono truncate block mt-0.5">
                      {{ file.virtual_path }}
                    </span>
                  </div>
                </div>
              </td>

              <!-- Origin Cloud Badge -->
              <td class="py-3 px-4 whitespace-nowrap" @click.stop>
                <div
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-xl border text-[11px] font-medium"
                  :class="getProviderMeta(file.provider).badgeColor"
                >
                  <UIcon :name="getProviderMeta(file.provider).icon" class="size-3.5" />
                  <span>{{ file.account_label }}</span>
                </div>
              </td>

              <!-- Size -->
              <td class="py-3 px-4 whitespace-nowrap font-mono text-zinc-200 font-medium">
                {{ formatBytes(file.size_bytes) }}
              </td>

              <!-- Date -->
              <td class="py-3 px-4 whitespace-nowrap text-zinc-400">
                {{ formatDate(file.modified_at) }}
              </td>

              <!-- Actions -->
              <td class="py-3 px-5 text-right whitespace-nowrap" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <UTooltip text="Download File">
                    <UButton
                      icon="i-lucide-download"
                      color="neutral"
                      variant="ghost"
                      size="xs"
                      square
                      class="rounded-lg hover:text-primary-400"
                      @click="handleDownload(file)"
                    />
                  </UTooltip>

                  <UDropdownMenu :items="getFileMenuItems(file)" :content="{ align: 'end' }">
                    <UButton
                      icon="i-lucide-more-horizontal"
                      color="neutral"
                      variant="ghost"
                      size="xs"
                      square
                      class="rounded-lg"
                    />
                  </UDropdownMenu>
                </div>
              </td>
            </tr>

            <!-- Memuat / gagal / kosong: tiga keadaan berbeda, bukan satu. -->
            <tr v-if="isListEmpty">
              <td colspan="6" class="p-0">
                <StateNotice
                  v-if="filesStore.isLoading || filesStore.isSearching"
                  variant="loading"
                  title="Loading files..."
                />
                <StateNotice
                  v-else-if="filesStore.loadError"
                  variant="error"
                  title="Could not load your files"
                  :description="filesStore.loadError"
                  @retry="retryLoad"
                />
                <StateNotice
                  v-else-if="hasActiveFilter"
                  title="No files match these filters"
                  description="Try a different category, account, or search term."
                  icon="i-lucide-search-x"
                >
                  <template #actions>
                    <UButton
                      label="Clear filters"
                      icon="i-lucide-filter-x"
                      color="neutral"
                      variant="subtle"
                      size="xs"
                      class="rounded-xl font-semibold"
                      @click="clearFilters"
                    />
                  </template>
                </StateNotice>
                <StateNotice
                  v-else
                  title="This virtual folder is empty"
                  description="Upload a file or create a subfolder here to get started"
                >
                  <template #actions>
                    <UButton
                      label="Upload File"
                      icon="i-lucide-upload-cloud"
                      color="primary"
                      size="xs"
                      class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white"
                      @click="filesStore.isUploadModalOpen = true"
                    />
                    <UButton
                      label="New Folder"
                      icon="i-lucide-folder-plus"
                      color="neutral"
                      variant="subtle"
                      size="xs"
                      class="rounded-xl font-semibold"
                      @click="openNewFolder"
                    />
                  </template>
                </StateNotice>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <ExplorerPager />
    </div>

    <!-- Files View: Grid Mode -->
    <div v-else class="space-y-6">
      <!-- Grid Virtual Folders Section -->
      <div v-if="filesStore.currentFolders.length > 0" class="space-y-2.5">
        <h3 class="text-xs font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-1.5">
          <UIcon name="i-lucide-folder" class="size-3.5 text-amber-400" />
          Virtual Folders ({{ filesStore.currentFolders.length }})
        </h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3.5">
          <div
            v-for="folder in filesStore.currentFolders"
            :key="folder.id"
            draggable="true"
            class="group flex cursor-pointer items-center justify-between rounded-2xl border p-3.5 shadow-xs transition-all"
            :class="dnd.isDropTarget(folder.id)
              ? 'border-primary-400/60 bg-primary-500/15 ring-1 ring-primary-400/40'
              : 'border-white/[0.08] bg-[#141925] hover:border-amber-500/40 hover:bg-[#1b2130]'"
            @click="filesStore.navigateToFolder(folder.id)"
            @dragstart="dnd.startFolderDrag($event, folder)"
            @dragend="dnd.endDrag()"
            @dragover="dnd.onDragOverFolder($event, folder.id)"
            @dragleave="dnd.onDragLeaveFolder(folder.id)"
            @drop="dnd.dropInto($event, folder.id)"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div class="p-2.5 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20 group-hover:scale-105 transition-transform shadow-xs">
                <UIcon name="i-lucide-folder" class="size-5 fill-amber-500/20" />
              </div>
              <div class="min-w-0">
                <h4 class="font-semibold text-xs text-zinc-200 truncate group-hover:text-amber-300 transition-colors">
                  {{ folder.name }}
                </h4>
                <span class="text-[11px] font-mono text-zinc-500 block">
                  {{ folder.subfolder_count }} subfolder
                </span>
              </div>
            </div>

            <div @click.stop>
              <UDropdownMenu :items="getFolderMenuItems(folder)" :content="{ align: 'end' }">
                <UButton
                  icon="i-lucide-more-vertical"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  square
                  class="rounded-lg text-zinc-400 hover:text-white"
                />
              </UDropdownMenu>
            </div>
          </div>
        </div>
      </div>

      <!-- Grid Files Section -->
      <div class="space-y-2.5">
        <h3 v-if="filesStore.currentFolders.length > 0" class="text-xs font-bold text-zinc-400 uppercase tracking-wider flex items-center gap-1.5">
          <UIcon name="i-lucide-files" class="size-3.5 text-primary-400" />
          Files ({{ filesStore.totalFiles }})
        </h3>

        <div
          v-if="filesStore.filteredFiles.length > 0"
          class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
        >
          <div
            v-for="file in filesStore.filteredFiles"
            :key="file.id"
            draggable="true"
            class="group flex cursor-pointer flex-col justify-between rounded-3xl border p-4 shadow-xs transition-all"
            :class="filesStore.isSelected(file.id)
              ? 'border-primary-500/50 bg-primary-500/[0.07]'
              : 'border-white/[0.08] bg-[#141925] hover:border-primary-500/30 hover:bg-[#1a1f2d]'"
            @click="handleRowClick(file, $event)"
            @dragstart="dnd.startFileDrag($event, file)"
            @dragend="dnd.endDrag()"
          >
            <div class="space-y-3">
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-2">
                  <span @click.stop>
                    <UCheckbox
                      :model-value="filesStore.isSelected(file.id)"
                      :aria-label="`Select ${file.name}`"
                      class="transition-opacity"
                      :class="filesStore.isSelected(file.id) ? '' : 'opacity-0 group-hover:opacity-100'"
                      @click="handleCheckboxClick(file, $event)"
                    />
                  </span>
                  <div class="p-2.5 rounded-2xl bg-[#1c2231] border border-white/[0.06] group-hover:scale-105 transition-transform shadow-xs">
                    <UIcon
                      :name="getFileIcon(file.mime, file.name).icon"
                      class="size-6"
                      :class="getFileIcon(file.mime, file.name).color"
                    />
                  </div>
                </div>

                <div @click.stop>
                  <UDropdownMenu :items="getFileMenuItems(file)" :content="{ align: 'end' }">
                    <UButton
                      icon="i-lucide-more-vertical"
                      color="neutral"
                      variant="ghost"
                      size="xs"
                      square
                      class="rounded-lg text-zinc-400 hover:text-white"
                    />
                  </UDropdownMenu>
                </div>
              </div>

              <div>
                <h4 class="font-semibold text-xs text-zinc-200 line-clamp-2 group-hover:text-primary-400 transition-colors" :title="file.name">
                  {{ file.name }}
                </h4>
                <div class="flex items-center justify-between text-[11px] font-mono text-zinc-500 mt-1.5">
                  <span>{{ formatBytes(file.size_bytes) }}</span>
                  <span>{{ formatDate(file.modified_at).split(',')[0] }}</span>
                </div>
              </div>
            </div>

            <div class="mt-4 pt-3 border-t border-white/[0.06] flex items-center justify-between gap-2">
              <div class="flex items-center gap-1.5 text-[10px] font-medium truncate" :class="getProviderMeta(file.provider).badgeColor">
                <UIcon :name="getProviderMeta(file.provider).icon" class="size-3 shrink-0" />
                <span class="truncate">{{ file.account_label }}</span>
              </div>

              <UBadge
                v-if="file.is_chunked"
                label="Chunked"
                color="primary"
                variant="subtle"
                size="xs"
                class="text-[9px] px-1 py-0 rounded-md bg-primary-500/10 text-primary-400 border border-primary-500/20"
              />
            </div>
          </div>
        </div>

        <template v-else-if="filesStore.currentFolders.length === 0">
          <StateNotice
            v-if="filesStore.isLoading || filesStore.isSearching"
            variant="loading"
            title="Loading files..."
          />
          <StateNotice
            v-else-if="filesStore.loadError"
            variant="error"
            title="Could not load your files"
            :description="filesStore.loadError"
            @retry="retryLoad"
          />
          <StateNotice
            v-else-if="hasActiveFilter"
            title="No files match these filters"
            description="Try a different category, account, or search term."
            icon="i-lucide-search-x"
          >
            <template #actions>
              <UButton
                label="Clear filters"
                icon="i-lucide-filter-x"
                color="neutral"
                variant="subtle"
                size="xs"
                class="rounded-xl font-semibold"
                @click="clearFilters"
              />
            </template>
          </StateNotice>
          <StateNotice
            v-else
            title="This virtual folder is empty"
            description="Upload a file or create a subfolder here to get started"
          >
            <template #actions>
              <UButton
                label="Upload File"
                icon="i-lucide-upload-cloud"
                color="primary"
                size="xs"
                class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white"
                @click="filesStore.isUploadModalOpen = true"
              />
              <UButton
                label="New Folder"
                icon="i-lucide-folder-plus"
                color="neutral"
                variant="subtle"
                size="xs"
                class="rounded-xl font-semibold"
                @click="openNewFolder"
              />
            </template>
          </StateNotice>
        </template>
      </div>

      <div v-if="filesStore.totalFiles > 0" class="rounded-2xl border border-white/[0.08] bg-[#0d111a] overflow-hidden">
        <ExplorerPager />
      </div>
    </div>

    <!-- Interactive Rich File Preview Modal (Images, Videos, Audio, PDF, Code, Spreadsheets, Archives) -->
    <FilePreviewModal
      v-model:open="isPreviewModalOpen"
      :file="previewFile"
      @download="handleDownload"
      @move-to-folder="openMoveToFolder"
      @migrate-provider="openMoveModal"
      @delete="openDeleteFile"
      @rename="openRenameFile"
    />

    <!-- Physical Migration Modal (Between Cloud Providers) -->
    <UModal
      v-model:open="isMoveModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white">Migrate File Across Cloud Providers</h3>
          <p class="text-xs text-zinc-400 mt-0.5">Stream-through physical transfer without local disk buffering.</p>
        </div>
      </template>

      <template #body>
        <div v-if="selectedFile" class="space-y-4">
          <div class="p-3.5 rounded-2xl bg-[#151a27] border border-white/[0.07] text-xs space-y-1">
            <span class="text-zinc-400 block">File:</span>
            <span class="font-bold text-white block truncate">{{ selectedFile.name }} ({{ formatBytes(selectedFile.size_bytes) }})</span>
            <span class="text-zinc-400 block mt-1">Current provider: <strong class="text-primary-400">{{ selectedFile.account_label }}</strong></span>
          </div>

          <div>
            <label class="block text-xs font-bold text-zinc-300 mb-1.5">
              Select Destination Cloud Provider
            </label>
            <USelect
              v-model="targetMoveAccount"
              :items="moveOptions"
              class="w-full rounded-xl bg-[#1c2231] border border-white/[0.08] text-white"
              icon="i-lucide-cloud-upload"
              size="md"
            />
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isMoveModalOpen = false" />
          <UButton label="Migrate Now" icon="i-lucide-arrow-right-left" color="primary" class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white" @click="confirmMove" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 1: Create New Folder -->
    <UModal
      v-model:open="isNewFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-folder-plus" class="size-5 text-primary-400" />
            Create Virtual Folder
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">
            Creating under: <code class="text-primary-400 font-mono">{{ filesStore.currentPath }}</code>
          </p>
        </div>
      </template>

      <template #body>
        <div class="space-y-3">
          <div>
            <label class="block text-xs font-semibold text-zinc-300 mb-1.5">Folder Name</label>
            <UInput
              v-model="newFolderName"
              placeholder="e.g. Q4_Reports, Invoices, Marketing"
              icon="i-lucide-folder"
              class="w-full rounded-xl"
              size="md"
              autofocus
              @keyup.enter="confirmCreateFolder"
            />
            <p v-if="newFolderError" class="text-xs text-rose-400 mt-1 font-medium">{{ newFolderError }}</p>
          </div>

          <div class="p-3 rounded-xl bg-zinc-900/80 border border-white/[0.06] text-[11px] text-zinc-400">
            <span class="text-primary-400 font-semibold">VFS Architecture Note:</span> Folders exist purely in database metadata. No empty folder overhead in your cloud accounts.
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isNewFolderModalOpen = false" />
          <UButton :loading="isBusy" label="Create Folder" color="primary" class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white" @click="confirmCreateFolder" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 2: Rename Folder -->
    <UModal
      v-model:open="isRenameFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-edit-2" class="size-5 text-primary-400" />
            Rename Virtual Folder
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Renaming cascades instantly to all child virtual paths.</p>
        </div>
      </template>

      <template #body>
        <div v-if="folderToRename" class="space-y-3">
          <div>
            <label class="block text-xs font-semibold text-zinc-300 mb-1.5">New Name</label>
            <UInput
              v-model="renamedFolderName"
              icon="i-lucide-folder"
              class="w-full rounded-xl"
              size="md"
              autofocus
              @keyup.enter="confirmRenameFolder"
            />
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isRenameFolderModalOpen = false" />
          <UButton label="Save Changes" color="primary" class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white" :loading="isBusy" @click="confirmRenameFolder" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 3: Delete Folder Confirmation -->
    <UModal
      v-model:open="isDeleteFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2 text-rose-400">
            <UIcon name="i-lucide-trash-2" class="size-5" />
            Delete Virtual Folder
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Are you sure you want to remove this folder?</p>
        </div>
      </template>

      <template #body>
        <div v-if="folderToDelete" class="space-y-3">
          <div class="p-3.5 rounded-2xl bg-[#1d1620] border border-rose-500/20 text-xs space-y-1.5">
            <div class="flex items-center gap-2 text-rose-300 font-semibold">
              <UIcon name="i-lucide-alert-triangle" class="size-4 shrink-0" />
              <span>Permanent deletion</span>
            </div>
            <p class="text-zinc-400 text-[11px] leading-relaxed">
              <strong class="text-white font-mono">{{ folderToDelete.path }}</strong> —
              a recursive delete also removes subfolders <em>and the physical files at the
              provider</em>, not just the index records.
            </p>
          </div>

          <p v-if="actionError" class="text-[11px] text-red-400 leading-snug">{{ actionError }}</p>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isBusy" @click="isDeleteFolderModalOpen = false" />
          <UButton
            label="Delete if empty"
            color="neutral"
            variant="soft"
            class="rounded-xl font-semibold"
            :loading="isBusy"
            @click="confirmDeleteFolder(false)"
          />
          <UButton
            label="Delete with contents"
            color="error"
            class="rounded-xl font-bold"
            :loading="isBusy"
            @click="confirmDeleteFolder(true)"
          />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 4: Move File to Virtual Folder -->
    <UModal
      v-model:open="isMoveToFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-folder-input" class="size-5 text-primary-400" />
            Move File to Virtual Folder
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Instant database update without moving physical provider blocks.</p>
        </div>
      </template>

      <template #body>
        <div v-if="fileToMoveVfs" class="space-y-4">
          <div class="p-3.5 rounded-2xl bg-[#151a27] border border-white/[0.07] text-xs space-y-1">
            <span class="text-zinc-400 block">File:</span>
            <span class="font-bold text-white block truncate">{{ fileToMoveVfs.name }}</span>
            <span class="text-zinc-400 block mt-1">Current Virtual Path: <strong class="text-primary-400 font-mono">{{ fileToMoveVfs.virtual_path }}</strong></span>
          </div>

          <div>
            <label class="block text-xs font-bold text-zinc-300 mb-1.5">
              Select Destination Virtual Folder
            </label>
            <div class="max-h-60 overflow-y-auto space-y-1 rounded-2xl border border-white/[0.08] bg-[#151a27] p-1.5">
              <button
                v-for="item in filesStore.allFoldersHierarchical"
                :key="item.path"
                type="button"
                class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs transition-colors cursor-pointer text-left"
                :class="[
                  selectedTargetVfsFolderId === item.id
                    ? 'bg-primary-500/15 text-primary-300 font-semibold border border-primary-500/25'
                    : 'text-zinc-300 hover:bg-white/[0.05]'
                ]"
                :style="{ paddingLeft: `${item.depth * 14 + 12}px` }"
                @click="selectedTargetVfsFolderId = item.id"
              >
                <span class="flex items-center gap-2 truncate">
                  <UIcon :name="item.id === null ? 'i-lucide-hard-drive' : 'i-lucide-folder'" class="size-3.5 text-zinc-400" />
                  <span class="truncate">{{ item.name }}</span>
                </span>
                <UIcon v-if="selectedTargetVfsFolderId === item.id" name="i-lucide-check" class="size-3.5 text-primary-400 shrink-0" />
              </button>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isMoveToFolderModalOpen = false" />
          <UButton label="Move to Folder" color="primary" class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white" :loading="isBusy" @click="confirmMoveToFolder" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 5: Rename File -->
    <UModal
      v-model:open="isRenameFileModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-edit-3" class="size-5 text-primary-400" />
            Rename File
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Updates file name and virtual path in the database index.</p>
        </div>
      </template>

      <template #body>
        <div v-if="fileToRename" class="space-y-3">
          <div>
            <label class="block text-xs font-semibold text-zinc-300 mb-1.5">File Name</label>
            <UInput
              v-model="renamedFileName"
              icon="i-lucide-file"
              class="w-full rounded-xl"
              size="md"
              autofocus
              @keyup.enter="confirmRenameFile"
            />
          </div>
          <p v-if="actionError" class="text-[11px] text-red-400 leading-snug">{{ actionError }}</p>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isBusy" @click="isRenameFileModalOpen = false" />
          <UButton label="Save Name" color="primary" class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white" :loading="isBusy" @click="confirmRenameFile" />
        </div>
      </template>
    </UModal>

    <!-- Konfirmasi hapus file: objek fisik di provider ikut terhapus -->
    <UModal
      v-model:open="isDeleteFileModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold flex items-center gap-2 text-rose-400">
            <UIcon name="i-lucide-trash-2" class="size-5" />
            Delete File
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">The file is removed from the provider, not just the index.</p>
        </div>
      </template>

      <template #body>
        <div v-if="fileToDelete" class="space-y-3">
          <div class="p-3.5 rounded-2xl bg-[#1d1620] border border-rose-500/20 text-xs space-y-1">
            <span class="font-bold text-white block truncate">{{ fileToDelete.name }}</span>
            <span class="text-zinc-400 block">
              Stored on <strong class="text-zinc-300">{{ fileToDelete.account_label }}</strong> —
              {{ formatBytes(fileToDelete.size_bytes) }}
            </span>
          </div>
          <p v-if="actionError" class="text-[11px] text-red-400 leading-snug">{{ actionError }}</p>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isBusy" @click="isDeleteFileModalOpen = false" />
          <UButton label="Delete" color="error" class="rounded-xl font-bold" :loading="isBusy" @click="confirmDeleteFile" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 6: Move Folder ke induk lain -->
    <UModal
      v-model:open="isMoveFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-folder-input" class="size-5 text-primary-400" />
            Move Virtual Folder
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Paths of every item inside are recalculated.</p>
        </div>
      </template>

      <template #body>
        <div v-if="folderToMove" class="space-y-4">
          <div class="p-3.5 rounded-2xl bg-[#151a27] border border-white/[0.07] text-xs space-y-1">
            <span class="text-zinc-400 block">Folder:</span>
            <span class="font-bold text-white block truncate">{{ folderToMove.name }}</span>
            <span class="text-zinc-400 block mt-1">
              Current path: <strong class="text-primary-400 font-mono">{{ folderToMove.path }}</strong>
            </span>
          </div>

          <div>
            <label class="block text-xs font-bold text-zinc-300 mb-1.5">Select New Parent Folder</label>
            <div class="max-h-60 overflow-y-auto space-y-1 rounded-2xl border border-white/[0.08] bg-[#151a27] p-1.5">
              <button
                v-for="item in moveFolderOptions"
                :key="item.path"
                type="button"
                class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs transition-colors cursor-pointer text-left"
                :class="[
                  targetParentFolderId === item.id
                    ? 'bg-primary-500/15 text-primary-300 font-semibold border border-primary-500/25'
                    : 'text-zinc-300 hover:bg-white/[0.05]'
                ]"
                :style="{ paddingLeft: `${item.depth * 14 + 12}px` }"
                @click="targetParentFolderId = item.id"
              >
                <span class="flex items-center gap-2 truncate">
                  <UIcon :name="item.id === null ? 'i-lucide-hard-drive' : 'i-lucide-folder'" class="size-3.5 text-zinc-400" />
                  <span class="truncate">{{ item.name }}</span>
                </span>
                <UIcon v-if="targetParentFolderId === item.id" name="i-lucide-check" class="size-3.5 text-primary-400 shrink-0" />
              </button>
            </div>
            <p v-if="moveFolderOptions.length === 0" class="text-[11px] text-zinc-500 mt-2">
              No other folder can hold this one.
            </p>
          </div>

          <p v-if="actionError" class="text-[11px] text-red-400 leading-snug">{{ actionError }}</p>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isBusy" @click="isMoveFolderModalOpen = false" />
          <UButton
            label="Move Folder"
            color="primary"
            class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white"
            :loading="isBusy"
            :disabled="moveFolderOptions.length === 0"
            @click="confirmMoveFolder"
          />
        </div>
      </template>
    </UModal>

    <!-- Aksi massal 1: pindahkan file terpilih ke satu folder virtual -->
    <UModal
      v-model:open="isBulkMoveOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-folder-input" class="size-5 text-primary-400" />
            Move {{ filesStore.selectedFileIds.length }} Files
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Database-only change; nothing moves at the provider.</p>
        </div>
      </template>

      <template #body>
        <div>
          <label class="block text-xs font-bold text-zinc-300 mb-1.5">Select Destination Virtual Folder</label>
            <div class="max-h-60 overflow-y-auto space-y-1 rounded-2xl border border-white/[0.08] bg-[#151a27] p-1.5">
              <button
                v-for="item in filesStore.allFoldersHierarchical"
                :key="item.path"
                type="button"
                class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs transition-colors cursor-pointer text-left"
                :class="[
                  bulkTargetFolderId === item.id
                    ? 'bg-primary-500/15 text-primary-300 font-semibold border border-primary-500/25'
                    : 'text-zinc-300 hover:bg-white/[0.05]'
                ]"
                :style="{ paddingLeft: `${item.depth * 14 + 12}px` }"
                @click="bulkTargetFolderId = item.id"
              >
                <span class="flex items-center gap-2 truncate">
                  <UIcon :name="item.id === null ? 'i-lucide-hard-drive' : 'i-lucide-folder'" class="size-3.5 text-zinc-400" />
                  <span class="truncate">{{ item.name }}</span>
                </span>
                <UIcon v-if="bulkTargetFolderId === item.id" name="i-lucide-check" class="size-3.5 text-primary-400 shrink-0" />
              </button>
            </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isBusy" @click="isBulkMoveOpen = false" />
          <UButton label="Move Files" color="primary" class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white" :loading="isBusy" @click="confirmBulkMove" />
        </div>
      </template>
    </UModal>

    <!-- Aksi massal 2: migrasikan file terpilih ke akun lain -->
    <UModal
      v-model:open="isBulkMigrateOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-arrow-right-left" class="size-5 text-primary-400" />
            Migrate {{ filesStore.selectedFileIds.length }} Files
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">
            Real transfer of {{ formatBytes(filesStore.selectedBytes) }}, one file at a time.
          </p>
        </div>
      </template>

      <template #body>
        <div class="space-y-3">
          <label class="block text-xs font-bold text-zinc-300">Select Destination Cloud Provider</label>
          <USelect
            v-model="bulkTargetAccountId"
            :items="bulkMigrateOptions"
            class="w-full rounded-xl bg-[#1c2231] border border-white/[0.08] text-white"
            icon="i-lucide-cloud-upload"
            size="md"
          />
          <p class="text-[11px] text-zinc-500 leading-relaxed">
            Files already on the destination account are migrated too; the router does not skip them.
          </p>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isBusy" @click="isBulkMigrateOpen = false" />
          <UButton
            label="Migrate Now"
            color="primary"
            class="rounded-xl font-bold bg-primary-600 hover:bg-primary-500 text-white"
            :loading="isBusy"
            :disabled="!bulkTargetAccountId"
            @click="confirmBulkMigrate"
          />
        </div>
      </template>
    </UModal>

    <!-- Aksi massal 3: hapus file terpilih (objek di provider ikut terhapus) -->
    <UModal
      v-model:open="isBulkDeleteOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold flex items-center gap-2 text-rose-400">
            <UIcon name="i-lucide-trash-2" class="size-5" />
            Delete {{ filesStore.selectedFileIds.length }} Files
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Files are removed from the providers, not just the index.</p>
        </div>
      </template>

      <template #body>
        <div class="p-3.5 rounded-2xl bg-[#1d1620] border border-rose-500/20 text-xs space-y-1.5">
          <div class="flex items-center gap-2 text-rose-300 font-semibold">
            <UIcon name="i-lucide-alert-triangle" class="size-4 shrink-0" />
            <span>Permanent deletion</span>
          </div>
          <p class="text-zinc-400 text-[11px] leading-relaxed">
            {{ filesStore.selectedFileIds.length }} files totalling
            <strong class="text-white">{{ formatBytes(filesStore.selectedBytes) }}</strong>
            will be deleted at their providers. This cannot be undone.
          </p>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isBusy" @click="isBulkDeleteOpen = false" />
          <UButton label="Delete Files" color="error" class="rounded-xl font-bold" :loading="isBusy" @click="confirmBulkDelete" />
        </div>
      </template>
    </UModal>
  </div>
</template>
