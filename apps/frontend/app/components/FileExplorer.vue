<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { FileEntry, FolderEntry } from '~/types'

const filesStore = useFilesStore()
const accountsStore = useAccountsStore()
const { formatBytes, formatDate, getProviderMeta, getFileIcon } = useFormatters()

// Modal & Selection States
const selectedFile = ref<FileEntry | null>(null)
const previewFile = ref<FileEntry | null>(null)
const isPreviewModalOpen = ref(false)

// Cloud Migration Modal (Physical Remote Transfer)
const isMoveModalOpen = ref(false)
const targetMoveAccount = ref<string>('')

// VFS Virtual Organization Modals
const isNewFolderModalOpen = ref(false)
const newFolderName = ref('')
const newFolderError = ref('')

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
        onSelect: () => navigator.clipboard?.writeText(file.virtual_path || (file.path + '/' + file.name))
      }
    ],
    [
      {
        label: 'Delete from Cloud',
        icon: 'i-lucide-trash-2',
        color: 'error',
        onSelect: () => filesStore.deleteFile(file.id)
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
        onSelect: () => filesStore.navigateToFolder(folder.id)
      },
      {
        label: 'Rename Folder',
        icon: 'i-lucide-edit-2',
        onSelect: () => openRenameFolder(folder)
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

function handleDownload(file: FileEntry) {
  const blob = new Blob([`Poly Cloud Stream-through payload for ${file.name}`], { type: file.mime || 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = file.name
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
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

function confirmMove() {
  if (!selectedFile.value || !targetMoveAccount.value) return
  const dest = accountsStore.accounts.find(a => a.id === targetMoveAccount.value)
  if (dest) {
    filesStore.moveFile(selectedFile.value.id, {
      id: dest.id,
      label: dest.label,
      provider: dest.provider
    })
  }
  isMoveModalOpen.value = false
}

const moveOptions = computed(() => {
  if (!selectedFile.value) return []
  return accountsStore.accounts
    .filter(a => a.id !== selectedFile.value?.account_id)
    .map(a => ({
      value: a.id,
      label: `${a.label} (${formatBytes(a.free_bytes)} free)`
    }))
})

// VFS: Create New Virtual Folder
function openNewFolder() {
  newFolderName.value = ''
  newFolderError.value = ''
  isNewFolderModalOpen.value = true
}

function confirmCreateFolder() {
  if (!newFolderName.value.trim()) {
    newFolderError.value = 'Please enter a folder name.'
    return
  }
  try {
    filesStore.createFolder(newFolderName.value)
    isNewFolderModalOpen.value = false
  } catch (err: any) {
    newFolderError.value = err.message || 'Error creating folder'
  }
}

// VFS: Rename Folder
function openRenameFolder(folder: FolderEntry) {
  folderToRename.value = folder
  renamedFolderName.value = folder.name
  isRenameFolderModalOpen.value = true
}

function confirmRenameFolder() {
  if (!folderToRename.value || !renamedFolderName.value.trim()) return
  filesStore.renameFolder(folderToRename.value.id, renamedFolderName.value)
  isRenameFolderModalOpen.value = false
}

// VFS: Delete Folder
function openDeleteFolder(folder: FolderEntry) {
  folderToDelete.value = folder
  isDeleteFolderModalOpen.value = true
}

function confirmDeleteFolder() {
  if (!folderToDelete.value) return
  filesStore.deleteFolder(folderToDelete.value.id, true)
  isDeleteFolderModalOpen.value = false
}

// VFS: Move File to Virtual Folder
function openMoveToFolder(file: FileEntry) {
  fileToMoveVfs.value = file
  selectedTargetVfsFolderId.value = file.folder_id || null
  isMoveToFolderModalOpen.value = true
}

function confirmMoveToFolder() {
  if (!fileToMoveVfs.value) return
  filesStore.moveFileToFolder(fileToMoveVfs.value.id, selectedTargetVfsFolderId.value)
  isMoveToFolderModalOpen.value = false
}

// VFS: Rename File
function openRenameFile(file: FileEntry) {
  fileToRename.value = file
  renamedFileName.value = file.name
  isRenameFileModalOpen.value = true
}

function confirmRenameFile() {
  if (!fileToRename.value || !renamedFileName.value.trim()) return
  filesStore.renameFile(fileToRename.value.id, renamedFileName.value)
  isRenameFileModalOpen.value = false
}
</script>

<template>
  <div class="space-y-4">
    <!-- VFS Breadcrumbs & Location Bar -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 rounded-2xl bg-[#0c0c0e] border border-white/[0.08] shadow-xs">
      <!-- Breadcrumb Path Trail -->
      <div class="flex items-center gap-1.5 overflow-x-auto py-1 text-xs">
        <!-- Up one level button (if not in root) -->
        <button
          v-if="filesStore.currentFolderId !== null"
          type="button"
          title="Go up one folder"
          class="flex size-7 items-center justify-center rounded-lg bg-[#15151a] hover:bg-[#1a1a20] text-zinc-400 hover:text-white border border-white/[0.08] transition-colors cursor-pointer shrink-0 mr-1"
          @click="filesStore.navigateUp()"
        >
          <UIcon name="i-lucide-corner-left-up" class="size-3.5" />
        </button>

        <div class="flex items-center gap-1 font-mono">
          <template v-for="(crumb, idx) in filesStore.breadcrumbs" :key="crumb.path">
            <button
              type="button"
              class="inline-flex items-center gap-1 px-2 py-1 rounded-lg transition-colors cursor-pointer"
              :class="[
                idx === filesStore.breadcrumbs.length - 1
                  ? 'bg-emerald-500/15 text-emerald-300 font-semibold border border-emerald-500/25'
                  : 'text-zinc-400 hover:text-white hover:bg-white/[0.05]'
              ]"
              @click="filesStore.navigateToFolder(crumb.id)"
            >
              <UIcon
                :name="idx === 0 ? 'i-lucide-hard-drive' : 'i-lucide-folder'"
                class="size-3.5"
                :class="idx === filesStore.breadcrumbs.length - 1 ? 'text-emerald-400' : 'text-zinc-500'"
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
          {{ filesStore.currentFolders.length }} folders, {{ filesStore.filteredFiles.length }} files
        </span>

        <button
          type="button"
          class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-[#16161b] hover:bg-[#1c1c22] text-zinc-200 hover:text-white border border-white/[0.08] transition-all cursor-pointer shadow-xs"
          @click="openNewFolder"
        >
          <UIcon name="i-lucide-folder-plus" class="size-3.5 text-emerald-400" />
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
          class="min-w-44 rounded-xl bg-[#121215] border border-white/[0.08] text-zinc-200 shadow-xs"
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
          class="w-36 rounded-xl bg-[#121215] border border-white/[0.08] text-zinc-200 shadow-xs"
          size="md"
        />

        <!-- View Mode Toggle -->
        <div class="flex items-center rounded-xl border border-white/[0.08] bg-[#121215] p-1 shadow-xs">
          <UButton
            icon="i-lucide-list"
            size="xs"
            square
            class="rounded-lg"
            :color="filesStore.viewMode === 'table' ? 'emerald' : 'neutral'"
            :variant="filesStore.viewMode === 'table' ? 'solid' : 'ghost'"
            @click="filesStore.viewMode = 'table'"
          />
          <UButton
            icon="i-lucide-layout-grid"
            size="xs"
            square
            class="rounded-lg"
            :color="filesStore.viewMode === 'grid' ? 'emerald' : 'neutral'"
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
            ? 'bg-emerald-600 text-white shadow-xs font-semibold'
            : 'bg-[#131316] hover:bg-[#18181d] text-zinc-400 hover:text-zinc-200 border border-white/[0.06]'
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

    <!-- Files View: Table Mode -->
    <div
      v-if="filesStore.viewMode === 'table'"
      class="rounded-3xl border border-white/[0.08] bg-[#0c0c0e] overflow-hidden shadow-sm"
    >
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse text-xs">
          <thead>
            <tr class="border-b border-white/[0.06] bg-[#111114] text-zinc-400 font-semibold uppercase tracking-wider text-[10px]">
              <th class="py-3.5 px-5">Name & Virtual Path</th>
              <th class="py-3.5 px-4">Storage Provider</th>
              <th class="py-3.5 px-4">Size</th>
              <th class="py-3.5 px-4">Last Modified</th>
              <th class="py-3.5 px-5 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/[0.04]">
            <!-- Virtual Folders Rows (Displayed first) -->
            <tr
              v-for="folder in filesStore.currentFolders"
              :key="folder.id"
              class="hover:bg-[#141418] transition-colors group cursor-pointer"
              @click="filesStore.navigateToFolder(folder.id)"
            >
              <!-- Folder Name & Icon -->
              <td class="py-3 px-5">
                <div class="flex items-center gap-3.5 min-w-[260px]">
                  <div class="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20 group-hover:scale-105 transition-transform shadow-xs">
                    <UIcon name="i-lucide-folder" class="size-5 fill-amber-500/20" />
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-semibold text-zinc-200 truncate group-hover:text-emerald-400 transition-colors">
                        {{ folder.name }}
                      </span>
                      <UBadge
                        label="Virtual Folder"
                        color="amber"
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
                  {{ folder.item_count || 0 }} items
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
                      class="rounded-lg hover:text-emerald-400"
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
              class="hover:bg-[#141418] transition-colors group cursor-pointer"
              @click="openPreview(file)"
            >
              <!-- Name & Icon -->
              <td class="py-3 px-5">
                <div class="flex items-center gap-3.5 min-w-[260px]">
                  <div class="p-2 rounded-xl bg-[#15151a] shrink-0 border border-white/[0.06] group-hover:scale-105 transition-transform shadow-xs">
                    <UIcon
                      :name="getFileIcon(file.mime, file.name).icon"
                      class="size-5"
                      :class="getFileIcon(file.mime, file.name).color"
                    />
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-semibold text-zinc-200 truncate max-w-xs md:max-w-md group-hover:text-emerald-400 transition-colors">
                        {{ file.name }}
                      </span>
                      <UBadge
                        v-if="file.is_chunked"
                        label="Chunked"
                        color="emerald"
                        variant="subtle"
                        size="xs"
                        class="text-[9px] px-1.5 py-0 rounded-md font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                      />
                    </div>
                    <span class="text-[11px] text-zinc-500 font-mono truncate block mt-0.5">
                      {{ file.virtual_path || (file.path + '/' + file.name) }}
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
                      class="rounded-lg hover:text-emerald-400"
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

            <!-- Empty State -->
            <tr v-if="filesStore.currentFolders.length === 0 && filesStore.filteredFiles.length === 0">
              <td colspan="5" class="py-16 text-center text-zinc-400">
                <div class="flex flex-col items-center justify-center gap-3">
                  <div class="p-4 rounded-3xl bg-[#121215] border border-white/[0.08]">
                    <UIcon name="i-lucide-folder-open" class="size-10 text-zinc-500" />
                  </div>
                  <div>
                    <h4 class="font-bold text-sm text-zinc-200">This virtual folder is empty</h4>
                    <p class="text-xs text-zinc-400 mt-0.5">Upload a file or create a subfolder here to get started</p>
                  </div>
                  <div class="flex items-center gap-2 mt-2">
                    <button
                      type="button"
                      class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white transition-all cursor-pointer shadow-xs"
                      @click="filesStore.isUploadModalOpen = true"
                    >
                      Upload File
                    </button>
                    <button
                      type="button"
                      class="px-3 py-1.5 rounded-xl text-xs font-semibold bg-[#16161b] hover:bg-[#1c1c22] text-zinc-300 border border-white/[0.08] transition-all cursor-pointer shadow-xs"
                      @click="openNewFolder"
                    >
                      New Folder
                    </button>
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
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
            class="flex items-center justify-between p-3.5 rounded-2xl border border-white/[0.08] bg-[#111114] hover:border-amber-500/40 hover:bg-[#15151a] transition-all group cursor-pointer shadow-xs"
            @click="filesStore.navigateToFolder(folder.id)"
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
                  {{ folder.item_count || 0 }} items
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
          <UIcon name="i-lucide-files" class="size-3.5 text-emerald-400" />
          Files ({{ filesStore.filteredFiles.length }})
        </h3>

        <div
          v-if="filesStore.filteredFiles.length > 0"
          class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
        >
          <div
            v-for="file in filesStore.filteredFiles"
            :key="file.id"
            class="flex flex-col justify-between p-4 rounded-3xl border border-white/[0.08] bg-[#111114] hover:border-emerald-500/30 hover:bg-[#141418] transition-all group cursor-pointer shadow-xs"
            @click="openPreview(file)"
          >
            <div class="space-y-3">
              <div class="flex items-start justify-between">
                <div class="p-2.5 rounded-2xl bg-[#16161b] border border-white/[0.06] group-hover:scale-105 transition-transform shadow-xs">
                  <UIcon
                    :name="getFileIcon(file.mime, file.name).icon"
                    class="size-6"
                    :class="getFileIcon(file.mime, file.name).color"
                  />
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
                <h4 class="font-semibold text-xs text-zinc-200 line-clamp-2 group-hover:text-emerald-400 transition-colors" :title="file.name">
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
                color="emerald"
                variant="subtle"
                size="xs"
                class="text-[9px] px-1 py-0 rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
              />
            </div>
          </div>
        </div>

        <div v-else-if="filesStore.currentFolders.length === 0" class="py-16 text-center text-zinc-400">
          <div class="flex flex-col items-center justify-center gap-3">
            <div class="p-4 rounded-3xl bg-[#121215] border border-white/[0.08]">
              <UIcon name="i-lucide-folder-open" class="size-10 text-zinc-500" />
            </div>
            <div>
              <h4 class="font-bold text-sm text-zinc-200">This virtual folder is empty</h4>
              <p class="text-xs text-zinc-400 mt-0.5">Upload a file or create a subfolder here to get started</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Preview Modal -->
    <UModal
      v-model:open="isPreviewModalOpen"
      :ui="{
        content: 'sm:max-w-xl bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div v-if="previewFile" class="px-6 pt-6 pb-2 flex items-start justify-between gap-4">
          <div class="flex items-center gap-3">
            <div class="p-3 rounded-2xl bg-[#15151a] border border-white/[0.06]">
              <UIcon
                :name="getFileIcon(previewFile.mime, previewFile.name).icon"
                class="size-7"
                :class="getFileIcon(previewFile.mime, previewFile.name).color"
              />
            </div>
            <div class="min-w-0">
              <h3 class="font-bold text-sm text-white truncate">{{ previewFile.name }}</h3>
              <p class="text-xs text-zinc-400 font-mono mt-0.5">{{ previewFile.virtual_path || previewFile.path }}</p>
            </div>
          </div>

          <button
            type="button"
            class="flex size-8 items-center justify-center rounded-xl bg-zinc-800/60 hover:bg-zinc-700 text-zinc-400 hover:text-white border border-white/[0.06] transition-colors cursor-pointer"
            @click="isPreviewModalOpen = false"
          >
            <UIcon name="i-lucide-x" class="size-4" />
          </button>
        </div>
      </template>

      <template #body>
        <div v-if="previewFile" class="space-y-4">
          <!-- Metadata Key-Value List -->
          <div class="p-4 rounded-2xl border border-white/[0.07] bg-[#121215] space-y-2.5 text-xs">
            <div class="flex justify-between py-1 border-b border-white/[0.06]">
              <span class="text-zinc-400">File Size</span>
              <span class="font-mono font-bold text-zinc-200">{{ formatBytes(previewFile.size_bytes) }} ({{ previewFile.size_bytes.toLocaleString() }} bytes)</span>
            </div>
            <div class="flex justify-between py-1 border-b border-white/[0.06]">
              <span class="text-zinc-400">Virtual Filesystem Path</span>
              <span class="font-mono font-semibold text-emerald-400">{{ previewFile.virtual_path || previewFile.path }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-white/[0.06]">
              <span class="text-zinc-400">Content Type (MIME)</span>
              <span class="font-mono font-semibold text-zinc-200">{{ previewFile.mime || 'application/octet-stream' }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-white/[0.06]">
              <span class="text-zinc-400">Origin Cloud Provider</span>
              <span class="font-semibold text-zinc-200 flex items-center gap-1.5">
                <UIcon :name="getProviderMeta(previewFile.provider).icon" class="size-3.5" />
                {{ previewFile.account_label }}
              </span>
            </div>
            <div class="flex justify-between py-1 border-b border-white/[0.06]">
              <span class="text-zinc-400">Last Synchronized</span>
              <span class="text-zinc-200">{{ formatDate(previewFile.modified_at) }}</span>
            </div>
            <div class="flex justify-between py-1">
              <span class="text-zinc-400">Storage Architecture Model</span>
              <span class="font-mono text-zinc-200">
                {{ previewFile.is_chunked ? 'Model B (Distributed Chunked Store)' : 'Model A (Whole-File Pass Through)' }}
              </span>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div v-if="previewFile" class="flex items-center justify-between w-full">
          <UButton
            label="Delete File"
            icon="i-lucide-trash-2"
            color="error"
            variant="ghost"
            class="rounded-xl text-xs"
            @click="filesStore.deleteFile(previewFile.id); isPreviewModalOpen = false"
          />

          <div class="flex items-center gap-2">
            <UButton
              label="Move Folder"
              icon="i-lucide-folder-input"
              color="neutral"
              variant="outline"
              class="rounded-xl text-xs"
              @click="openMoveToFolder(previewFile); isPreviewModalOpen = false"
            />
            <UButton
              label="Download File"
              icon="i-lucide-download"
              color="emerald"
              variant="solid"
              class="rounded-xl font-bold text-xs px-4 bg-emerald-600 hover:bg-emerald-500 text-white"
              @click="handleDownload(previewFile)"
            />
          </div>
        </div>
      </template>
    </UModal>

    <!-- Physical Migration Modal (Between Cloud Providers) -->
    <UModal
      v-model:open="isMoveModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
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
          <div class="p-3.5 rounded-2xl bg-[#121215] border border-white/[0.07] text-xs space-y-1">
            <span class="text-zinc-400 block">File:</span>
            <span class="font-bold text-white block truncate">{{ selectedFile.name }} ({{ formatBytes(selectedFile.size_bytes) }})</span>
            <span class="text-zinc-400 block mt-1">Current provider: <strong class="text-emerald-400">{{ selectedFile.account_label }}</strong></span>
          </div>

          <div>
            <label class="block text-xs font-bold text-zinc-300 mb-1.5">
              Select Destination Cloud Provider
            </label>
            <USelect
              v-model="targetMoveAccount"
              :items="moveOptions"
              class="w-full rounded-xl bg-[#16161a] border border-white/[0.08] text-white"
              icon="i-lucide-cloud-upload"
              size="md"
            />
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isMoveModalOpen = false" />
          <UButton label="Migrate Now" icon="i-lucide-arrow-right-left" color="emerald" class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-500 text-white" @click="confirmMove" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 1: Create New Folder -->
    <UModal
      v-model:open="isNewFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-folder-plus" class="size-5 text-emerald-400" />
            Create Virtual Folder
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">
            Creating under: <code class="text-emerald-400 font-mono">{{ filesStore.currentPath }}</code>
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
            <span class="text-emerald-400 font-semibold">VFS Architecture Note:</span> Folders exist purely in database metadata. No empty folder overhead in your cloud accounts.
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isNewFolderModalOpen = false" />
          <UButton label="Create Folder" color="emerald" class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-500 text-white" @click="confirmCreateFolder" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 2: Rename Folder -->
    <UModal
      v-model:open="isRenameFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-edit-2" class="size-5 text-emerald-400" />
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
          <UButton label="Save Changes" color="emerald" class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-500 text-white" @click="confirmRenameFolder" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 3: Delete Folder Confirmation -->
    <UModal
      v-model:open="isDeleteFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
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
          <div class="p-3.5 rounded-2xl bg-[#151214] border border-rose-500/20 text-xs space-y-1.5">
            <div class="flex items-center gap-2 text-rose-300 font-semibold">
              <UIcon name="i-lucide-alert-triangle" class="size-4 shrink-0" />
              <span>Recursive Deletion Warning</span>
            </div>
            <p class="text-zinc-400 text-[11px] leading-relaxed">
              Deleting <strong class="text-white font-mono">{{ folderToDelete.path }}</strong> will delete this virtual folder, its subfolders, and all indexed files contained within it.
            </p>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isDeleteFolderModalOpen = false" />
          <UButton label="Delete Everything" color="error" class="rounded-xl font-bold" @click="confirmDeleteFolder" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 4: Move File to Virtual Folder -->
    <UModal
      v-model:open="isMoveToFolderModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-folder-input" class="size-5 text-emerald-400" />
            Move File to Virtual Folder
          </h3>
          <p class="text-xs text-zinc-400 mt-0.5">Instant database update without moving physical provider blocks.</p>
        </div>
      </template>

      <template #body>
        <div v-if="fileToMoveVfs" class="space-y-4">
          <div class="p-3.5 rounded-2xl bg-[#121215] border border-white/[0.07] text-xs space-y-1">
            <span class="text-zinc-400 block">File:</span>
            <span class="font-bold text-white block truncate">{{ fileToMoveVfs.name }}</span>
            <span class="text-zinc-400 block mt-1">Current Virtual Path: <strong class="text-emerald-400 font-mono">{{ fileToMoveVfs.virtual_path || fileToMoveVfs.path }}</strong></span>
          </div>

          <div>
            <label class="block text-xs font-bold text-zinc-300 mb-1.5">
              Select Destination Virtual Folder
            </label>
            <div class="max-h-60 overflow-y-auto space-y-1 rounded-2xl border border-white/[0.08] bg-[#121215] p-1.5">
              <button
                v-for="item in filesStore.allFoldersHierarchical"
                :key="item.path"
                type="button"
                class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs transition-colors cursor-pointer text-left"
                :class="[
                  selectedTargetVfsFolderId === item.id
                    ? 'bg-emerald-500/15 text-emerald-300 font-semibold border border-emerald-500/25'
                    : 'text-zinc-300 hover:bg-white/[0.05]'
                ]"
                :style="{ paddingLeft: `${item.depth * 14 + 12}px` }"
                @click="selectedTargetVfsFolderId = item.id"
              >
                <span class="flex items-center gap-2 truncate">
                  <UIcon :name="item.id === null ? 'i-lucide-hard-drive' : 'i-lucide-folder'" class="size-3.5 text-zinc-400" />
                  <span class="truncate">{{ item.name }}</span>
                </span>
                <UIcon v-if="selectedTargetVfsFolderId === item.id" name="i-lucide-check" class="size-3.5 text-emerald-400 shrink-0" />
              </button>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isMoveToFolderModalOpen = false" />
          <UButton label="Move to Folder" color="emerald" class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-500 text-white" @click="confirmMoveToFolder" />
        </div>
      </template>
    </UModal>

    <!-- VFS Modal 5: Rename File -->
    <UModal
      v-model:open="isRenameFileModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon name="i-lucide-edit-3" class="size-5 text-emerald-400" />
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
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isRenameFileModalOpen = false" />
          <UButton label="Save Name" color="emerald" class="rounded-xl font-bold bg-emerald-600 hover:bg-emerald-500 text-white" @click="confirmRenameFile" />
        </div>
      </template>
    </UModal>
  </div>
</template>
