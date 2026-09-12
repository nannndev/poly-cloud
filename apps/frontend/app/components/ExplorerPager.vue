<script setup lang="ts">
/**
 * Kendali halaman untuk daftar file. Paging dikerjakan backend, jadi komponen ini
 * hanya menampilkan posisi dan meminta halaman lain — tak pernah memotong daftar
 * sendiri.
 */
const filesStore = useFilesStore()

const PER_PAGE_OPTIONS = [
  { value: 50, label: '50 / page' },
  { value: 100, label: '100 / page' },
  { value: 250, label: '250 / page' },
  { value: 500, label: '500 / page' }
]

const perPage = computed({
  get: () => filesStore.perPage,
  set: (value: number) => {
    filesStore.perPage = value
    void filesStore.applyFilters()
  }
})
</script>

<template>
  <div
    v-if="filesStore.totalFiles > 0"
    class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 px-4 py-3 border-t border-white/[0.06] bg-[#0e0e11]"
  >
    <span class="text-[11px] text-zinc-400 font-mono">
      Showing <strong class="text-zinc-200">{{ filesStore.pageStart }}–{{ filesStore.pageEnd }}</strong>
      of <strong class="text-zinc-200">{{ filesStore.totalFiles }}</strong> files
    </span>

    <div class="flex items-center gap-2">
      <USelect
        v-model="perPage"
        :items="PER_PAGE_OPTIONS"
        size="xs"
        class="w-32 rounded-xl bg-[#121215] border border-white/[0.08] text-zinc-200"
      />

      <div class="flex items-center gap-1">
        <UButton
          icon="i-lucide-chevron-left"
          color="neutral"
          variant="subtle"
          size="xs"
          square
          class="rounded-lg"
          :disabled="filesStore.page <= 1"
          aria-label="Previous page"
          @click="filesStore.goToPage(filesStore.page - 1)"
        />
        <span class="text-[11px] text-zinc-400 font-mono px-2 min-w-20 text-center">
          Page {{ filesStore.page }} / {{ filesStore.totalPages }}
        </span>
        <UButton
          icon="i-lucide-chevron-right"
          color="neutral"
          variant="subtle"
          size="xs"
          square
          class="rounded-lg"
          :disabled="filesStore.page >= filesStore.totalPages"
          aria-label="Next page"
          @click="filesStore.goToPage(filesStore.page + 1)"
        />
      </div>
    </div>
  </div>
</template>
