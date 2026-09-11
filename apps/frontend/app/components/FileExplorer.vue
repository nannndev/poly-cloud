<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'
import type { FileEntry } from '~/types'

const filesStore = useFilesStore()
const accountsStore = useAccountsStore()
const { formatBytes, formatDate, getProviderMeta, getFileIcon } = useFormatters()

const selectedFile = ref<FileEntry | null>(null)
const previewFile = ref<FileEntry | null>(null)
const isPreviewModalOpen = ref(false)
const isMoveModalOpen = ref(false)
const targetMoveAccount = ref<string>('')

const categories = [
  { id: 'all', label: 'All Files', icon: 'i-lucide-layers' },
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
        label: 'Download File (Direct Stream)',
        icon: 'i-lucide-download',
        onSelect: () => handleDownload(file)
      },
      {
        label: 'Move to Another Provider',
        icon: 'i-lucide-arrow-right-left',
        onSelect: () => openMoveModal(file)
      },
      {
        label: 'Copy Virtual Path',
        icon: 'i-lucide-copy',
        onSelect: () => navigator.clipboard?.writeText(file.path + '/' + file.name)
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
</script>

<template>
  <div class="space-y-4">
    <!-- Toolbar: Search, Category Filters, Provider Filter, View Toggle -->
    <div class="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
      <!-- Search Bar -->
      <div class="relative flex-1 max-w-md">
        <UInput
          v-model="filesStore.searchQuery"
          icon="i-lucide-search"
          placeholder="Search files, paths, or providers..."
          class="w-full rounded-xl"
          size="md"
        />
        <button
          v-if="filesStore.searchQuery"
          type="button"
          class="absolute right-3 top-1/2 -translate-y-1/2 text-muted hover:text-highlighted"
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
            { value: 'modified', label: 'Date Modified' },
            { value: 'name', label: 'File Name' },
            { value: 'size', label: 'File Size' }
          ]"
          icon="i-lucide-arrow-up-down"
          class="w-40 rounded-xl bg-[#121215] border border-white/[0.08] text-zinc-200 shadow-xs"
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
        <span>Showing <strong class="text-zinc-300 font-medium">{{ filesStore.filteredFiles.length }}</strong> items</span>
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
              <th class="py-3.5 px-5">File Name & Path</th>
              <th class="py-3.5 px-4">Origin Provider</th>
              <th class="py-3.5 px-4">Size</th>
              <th class="py-3.5 px-4">Last Modified</th>
              <th class="py-3.5 px-5 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-white/[0.04]">
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
                    <span class="text-[11px] text-zinc-500 font-mono truncate block mt-0.5">{{ file.path }}</span>
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
              <td class="py-3 px-4 whitespace-nowrap font-mono text-highlighted font-medium">
                {{ formatBytes(file.size_bytes) }}
              </td>

              <!-- Date -->
              <td class="py-3 px-4 whitespace-nowrap text-muted">
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
                      class="rounded-lg hover:text-sky-500"
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
            <tr v-if="filesStore.filteredFiles.length === 0">
              <td colspan="5" class="py-16 text-center text-muted">
                <div class="flex flex-col items-center justify-center gap-3">
                  <div class="p-4 rounded-3xl bg-elevated/80 border border-default/70">
                    <UIcon name="i-lucide-folder-search" class="size-10 text-muted/60" />
                  </div>
                  <div>
                    <h4 class="font-bold text-sm text-highlighted">No matching files found</h4>
                    <p class="text-xs text-muted mt-0.5">Try adjusting your search query or clearing provider filters</p>
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Files View: Grid Mode -->
    <div
      v-else
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

        <div class="mt-4 pt-3 border-t border-default/50 flex items-center justify-between gap-2">
          <div class="flex items-center gap-1.5 text-[10px] font-medium truncate" :class="getProviderMeta(file.provider).badgeColor">
            <UIcon :name="getProviderMeta(file.provider).icon" class="size-3 shrink-0" />
            <span class="truncate">{{ file.account_label }}</span>
          </div>

          <UBadge
            v-if="file.is_chunked"
            label="Chunked"
            color="sky"
            variant="subtle"
            size="xs"
            class="text-[9px] px-1 py-0 rounded-md"
          />
        </div>
      </div>
    </div>

    <!-- Quick Preview Modal -->
    <UModal
      v-model:open="isPreviewModalOpen"
      :ui="{
        content: 'sm:max-w-xl bg-card/95 backdrop-blur-2xl border border-default/80 rounded-3xl shadow-2xl p-0 overflow-hidden',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-elevated/40 border-t border-default/60'
      }"
    >
      <template #header>
        <div v-if="previewFile" class="px-6 pt-6 pb-2 flex items-start justify-between gap-4">
          <div class="flex items-center gap-3">
            <div class="p-3 rounded-2xl bg-elevated/80 border border-default/60">
              <UIcon
                :name="getFileIcon(previewFile.mime, previewFile.name).icon"
                class="size-7"
                :class="getFileIcon(previewFile.mime, previewFile.name).color"
              />
            </div>
            <div class="min-w-0">
              <h3 class="font-bold text-sm text-highlighted truncate">{{ previewFile.name }}</h3>
              <p class="text-xs text-muted font-mono mt-0.5">{{ previewFile.path }}</p>
            </div>
          </div>

          <UButton
            icon="i-lucide-x"
            color="neutral"
            variant="ghost"
            size="sm"
            square
            class="rounded-xl"
            @click="isPreviewModalOpen = false"
          />
        </div>
      </template>

      <template #body>
        <div v-if="previewFile" class="space-y-4">
          <!-- Metadata Key-Value List -->
          <div class="p-4 rounded-2xl border border-default/60 bg-elevated/30 space-y-2.5 text-xs">
            <div class="flex justify-between py-1 border-b border-default/40">
              <span class="text-muted">File Size</span>
              <span class="font-mono font-bold text-highlighted">{{ formatBytes(previewFile.size_bytes) }} ({{ previewFile.size_bytes.toLocaleString() }} bytes)</span>
            </div>
            <div class="flex justify-between py-1 border-b border-default/40">
              <span class="text-muted">Content Type (MIME)</span>
              <span class="font-mono font-semibold text-highlighted">{{ previewFile.mime || 'application/octet-stream' }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-default/40">
              <span class="text-muted">Origin Cloud Provider</span>
              <span class="font-semibold text-sky-500 flex items-center gap-1">
                <UIcon :name="getProviderMeta(previewFile.provider).icon" class="size-3.5" />
                {{ previewFile.account_label }}
              </span>
            </div>
            <div class="flex justify-between py-1 border-b border-default/40">
              <span class="text-muted">Last Synchronized</span>
              <span class="text-highlighted">{{ formatDate(previewFile.modified_at) }}</span>
            </div>
            <div class="flex justify-between py-1">
              <span class="text-muted">Storage Architecture Model</span>
              <span class="font-mono text-highlighted">
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
              label="Move Provider"
              icon="i-lucide-arrow-right-left"
              color="neutral"
              variant="outline"
              class="rounded-xl text-xs"
              @click="openMoveModal(previewFile); isPreviewModalOpen = false"
            />
            <UButton
              label="Download File"
              icon="i-lucide-download"
              color="sky"
              variant="solid"
              class="rounded-xl font-bold text-xs px-4"
              @click="handleDownload(previewFile)"
            />
          </div>
        </div>
      </template>
    </UModal>

    <!-- Move File Modal -->
    <UModal
      v-model:open="isMoveModalOpen"
      :ui="{
        content: 'sm:max-w-md bg-card/95 backdrop-blur-2xl border border-default/80 rounded-3xl shadow-2xl p-0 overflow-hidden',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-elevated/40 border-t border-default/60'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold text-highlighted">Move File Between Cloud Providers</h3>
          <p class="text-xs text-muted mt-0.5">Stream-through transfer directly without local disk writes.</p>
        </div>
      </template>

      <template #body>
        <div v-if="selectedFile" class="space-y-4">
          <div class="p-3.5 rounded-2xl bg-elevated/40 border border-default/60 text-xs space-y-1">
            <span class="text-muted block">File:</span>
            <span class="font-bold text-highlighted block truncate">{{ selectedFile.name }} ({{ formatBytes(selectedFile.size_bytes) }})</span>
            <span class="text-muted block mt-1">Current provider: <strong class="text-sky-500">{{ selectedFile.account_label }}</strong></span>
          </div>

          <div>
            <label class="block text-xs font-bold text-highlighted mb-1.5">
              Select Destination Cloud Provider
            </label>
            <USelect
              v-model="targetMoveAccount"
              :items="moveOptions"
              class="w-full rounded-xl"
              icon="i-lucide-cloud-upload"
              size="md"
            />
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" @click="isMoveModalOpen = false" />
          <UButton label="Move Now" icon="i-lucide-arrow-right-left" color="sky" class="rounded-xl font-bold" @click="confirmMove" />
        </div>
      </template>
    </UModal>
  </div>
</template>
