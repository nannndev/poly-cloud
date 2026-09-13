import { defineStore } from 'pinia'
import type {
  FileEntry,
  FileListResult,
  FolderEntry,
  StorageProvider,
  UploadJob,
  UploadResponse
} from '~/types'

export const useFilesStore = defineStore('files', () => {
  const api = useApi()
  const accountsStore = useAccountsStore()

  const folders = ref<FolderEntry[]>([])
  const files = ref<FileEntry[]>([])

  const currentFolderId = ref<string | null>(null)
  const searchQuery = ref('')
  const selectedAccountId = ref<string>('all')
  const selectedCategory = ref<string>('all')
  const viewMode = ref<'grid' | 'table'>('table')
  const sortBy = ref<'name' | 'size' | 'modified'>('name')
  const sortDirection = ref<'asc' | 'desc'>('asc')

  // Paging dikerjakan backend (query DB). Tanpa ini, daftar terpotong diam-diam
  // di baris ke-500 dan sisanya tak pernah terlihat.
  const page = ref(1)
  const perPage = ref(100)
  const totalFiles = ref(0)

  // Pilihan untuk aksi massal. Disimpan sebagai id, bukan objek file, supaya
  // tetap sahih setelah daftar dimuat ulang.
  const selectedFileIds = ref<string[]>([])

  const uploadJobs = ref<UploadJob[]>([])
  const isUploadModalOpen = ref(false)
  const isNewFolderModalOpen = ref(false)

  const isLoading = ref(false)
  const isSearching = ref(false)
  const loadError = ref<string | null>(null)

  // ---- Turunan untuk tampilan ----

  const currentFolder = computed(() =>
    currentFolderId.value ? folders.value.find(f => f.id === currentFolderId.value) || null : null)

  const currentPath = computed(() => currentFolder.value?.path || '/')

  const breadcrumbs = computed(() => {
    const crumbs: { id: string | null; name: string; path: string }[] = [
      { id: null, name: 'Root', path: '/' }
    ]
    if (!currentFolder.value) return crumbs

    // Telusuri ke atas lewat parent_id (adjacency = sumber kebenaran hierarki).
    const trail: FolderEntry[] = []
    let cursor: FolderEntry | undefined = currentFolder.value
    while (cursor) {
      trail.unshift(cursor)
      const parentId: string | null = cursor.parent_id
      cursor = parentId ? folders.value.find(f => f.id === parentId) : undefined
    }
    for (const f of trail) {
      crumbs.push({ id: f.id, name: f.name, path: f.path })
    }
    return crumbs
  })

  const currentFolders = computed(() => {
    // Saat mencari, folder disembunyikan: hasil pencarian bersifat lintas-folder.
    if (searchQuery.value.trim()) return []
    return folders.value
      .filter(f => f.parent_id === currentFolderId.value)
      .sort((a, b) => a.name.localeCompare(b.name))
      .map(folder => ({
        ...folder,
        subfolder_count: folders.value.filter(sub => sub.parent_id === folder.id).length
      }))
  })

  /** Daftar folder rata + kedalaman, untuk dropdown pilih folder tujuan. */
  const allFoldersHierarchical = computed(() => {
    function build(parentId: string | null, depth = 0): { id: string | null; name: string; path: string; depth: number }[] {
      const out: { id: string | null; name: string; path: string; depth: number }[] = []
      const children = folders.value
        .filter(f => f.parent_id === parentId)
        .sort((a, b) => a.name.localeCompare(b.name))
      for (const child of children) {
        out.push({ id: child.id, name: child.name, path: child.path, depth })
        out.push(...build(child.id, depth + 1))
      }
      return out
    }
    return [{ id: null, name: 'Root', path: '/', depth: 0 }, ...build(null)]
  })

  /**
   * Daftar yang dirender. Penyaringan, urutan, dan paging seluruhnya dikerjakan
   * backend atas index DB, jadi di sini tak ada penyaringan ulang — menyaring
   * lagi di klien hanya akan membuang baris dari halaman yang sedang tampil dan
   * membuat jumlah per halaman tampak tak konsisten dengan `totalFiles`.
   */
  const filteredFiles = computed(() => files.value)

  const totalPages = computed(() =>
    Math.max(1, Math.ceil(totalFiles.value / perPage.value)))

  const pageStart = computed(() =>
    totalFiles.value === 0 ? 0 : (page.value - 1) * perPage.value + 1)

  const pageEnd = computed(() =>
    Math.min(totalFiles.value, (page.value - 1) * perPage.value + files.value.length))

  /**
   * Kategori explorer dipetakan ke parameter `type` backend, yang dicocokkan
   * sebagai `mime ilike %...%`. Satu kategori bisa butuh beberapa pencocokan
   * (mis. media = image/video/audio) sementara backend hanya menerima satu pola,
   * jadi kategori multi-pola dikirim per pola dan digabung di sini.
   */
  const CATEGORY_MIME: Record<string, string[]> = {
    docs: ['pdf', 'word', 'sheet', 'presentation', 'document', 'text/', 'csv'],
    media: ['image/', 'video/', 'audio/'],
    archives: ['zip', 'tar', 'gzip', 'compressed', 'bzip', '7z', 'rar'],
    code: ['json', 'yaml', 'xml', 'sql', 'javascript', 'x-sh', 'text/x-', 'typescript']
  }

  // ---- Muat data ----

  async function fetchFolders() {
    // Muat seluruh pohon: hierarki dipakai untuk breadcrumb & dropdown tujuan,
    // dan jumlah folder virtual per user tergolong kecil.
    const roots = await api.get<FolderEntry[]>('/folders')
    const all: FolderEntry[] = [...roots]
    const queue = [...roots]
    while (queue.length > 0) {
      const parent = queue.shift()!
      const children = await api.get<FolderEntry[]>('/folders', { parent_id: parent.id })
      all.push(...children)
      queue.push(...children)
    }
    folders.value = all
  }

  /** Parameter penyaringan & urutan yang sama untuk list maupun search. */
  function baseQuery(): Record<string, any> {
    const query: Record<string, any> = {
      page: page.value,
      per_page: perPage.value,
      sort: sortBy.value === 'name'
        ? (sortDirection.value === 'asc' ? 'name' : 'name_desc')
        : `${sortBy.value}${sortDirection.value === 'asc' ? '' : '_desc'}`
    }
    if (selectedAccountId.value !== 'all') query.account_id = selectedAccountId.value
    const patterns = CATEGORY_MIME[selectedCategory.value]
    if (patterns) query.type = patterns.join(',')
    return query
  }

  async function fetchFiles() {
    const query = baseQuery()
    if (currentFolderId.value) query.folder_id = currentFolderId.value
    const res = await api.get<FileListResult>('/files', query)
    files.value = res.items
    totalFiles.value = res.total
  }

  async function loadAll() {
    isLoading.value = true
    loadError.value = null
    try {
      await fetchFolders()
      await fetchFiles()
    } catch (err) {
      loadError.value = friendlyMessage(err)
      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Seluruh file milik user lintas folder — dipakai halaman statistik, yang
   * menjumlah ukuran per kategori dan karenanya butuh setiap baris, bukan satu
   * halaman. Backend membatasi per_page di 500, jadi halaman berikutnya diambil
   * sampai `total` terpenuhi.
   */
  async function fetchAllFiles(): Promise<FileEntry[]> {
    const out: FileEntry[] = []
    let current = 1
    for (;;) {
      const res = await api.get<FileListResult>('/files/search', { page: current, per_page: 500 })
      out.push(...res.items)
      if (out.length >= res.total || res.items.length === 0) return out
      current++
    }
  }

  /** Pencarian lintas akun & folder, dijalankan backend atas index DB. */
  async function runSearch() {
    const q = searchQuery.value.trim()
    if (!q) {
      await fetchFiles()
      return
    }
    isSearching.value = true
    try {
      const res = await api.get<FileListResult>('/files/search', { ...baseQuery(), q })
      files.value = res.items
      totalFiles.value = res.total
    } finally {
      isSearching.value = false
    }
  }

  /**
   * Muat ulang daftar aktif — pencarian bila sedang mencari, isi folder bila tidak.
   * Menghapus isi halaman terakhir bisa membuat `page` melewati halaman terakhir
   * yang tersisa; kalau itu terjadi, mundur satu halaman dan muat lagi, supaya
   * daftar tak tampak kosong padahal masih ada isinya.
   */
  async function refreshList() {
    const load = () => (searchQuery.value.trim() ? runSearch() : fetchFiles())
    await load()
    while (page.value > 1 && files.value.length === 0 && totalFiles.value > 0) {
      page.value = Math.min(page.value - 1, totalPages.value)
      await load()
    }
  }

  /** Pindah halaman; nomor di luar rentang diabaikan. */
  async function goToPage(next: number) {
    if (next < 1 || next > totalPages.value || next === page.value) return
    page.value = next
    clearSelection()
    await refreshList()
  }

  /** Filter/urutan berubah → daftar kembali ke halaman pertama. */
  async function applyFilters() {
    page.value = 1
    clearSelection()
    await refreshList()
  }

  async function navigateToFolder(folderId: string | null) {
    currentFolderId.value = folderId
    searchQuery.value = ''
    page.value = 1
    clearSelection()
    await fetchFiles()
  }

  async function navigateUp() {
    await navigateToFolder(currentFolder.value?.parent_id ?? null)
  }

  // ---- Operasi VFS (transaksi DB murni di backend, tak menyentuh provider) ----

  async function createFolder(name: string, parentId: string | null = currentFolderId.value) {
    const folder = await api.post<FolderEntry>('/folders', { name: name.trim(), parent_id: parentId })
    folders.value = [...folders.value, folder]
    return folder
  }

  async function renameFolder(folderId: string, newName: string) {
    await api.patch<FolderEntry>(`/folders/${folderId}`, { name: newName.trim() })
    // Rename menggeser path seluruh turunan; muat ulang agar cache path akurat.
    await fetchFolders()
    await refreshList()
  }

  /**
   * Pindahkan folder ke induk lain. Backend menolak memindahkan folder ke dalam
   * turunannya sendiri; UI menyaring pilihan itu lebih dulu agar tak sampai ke
   * sana, tapi penjagaan sebenarnya tetap di backend.
   */
  async function moveFolder(folderId: string, newParentId: string | null) {
    await api.patch<FolderEntry>(`/folders/${folderId}`, { parent_id: newParentId })
    await fetchFolders()
    await refreshList()
  }

  async function deleteFolder(folderId: string, recursive = false) {
    await api.del(`/folders/${folderId}`, { recursive })

    // Kalau sedang berada di dalam folder yang dihapus, naik ke induknya.
    const deleted = folders.value.find(f => f.id === folderId)
    if (deleted && currentFolder.value?.path.startsWith(deleted.path)) {
      currentFolderId.value = deleted.parent_id
      page.value = 1
    }
    await fetchFolders()
    await refreshList()
    if (recursive) await accountsStore.fetchQuota().catch(() => null)
  }

  async function moveFileToFolder(fileId: string, targetFolderId: string | null) {
    const updated = await api.patch<FileEntry>(`/files/${fileId}`, { folder_id: targetFolderId })
    // File yang pindah keluar dari folder aktif tak lagi termasuk halaman ini.
    if (!searchQuery.value.trim() && updated.folder_id !== currentFolderId.value) {
      await refreshList()
    } else {
      applyFileUpdate(updated)
    }
  }

  async function renameFile(fileId: string, newName: string) {
    const updated = await api.patch<FileEntry>(`/files/${fileId}`, { name: newName.trim() })
    applyFileUpdate(updated)
  }

  function applyFileUpdate(updated: FileEntry) {
    const idx = files.value.findIndex(f => f.id === updated.id)
    if (idx !== -1) files.value[idx] = updated
  }

  // ---- Operasi yang menyentuh provider ----

  async function deleteFile(fileId: string) {
    await api.del(`/files/${fileId}`)
    // Baris hilang dari index, jadi halaman ini digeser isi halaman berikutnya —
    // muat ulang alih-alih menyunting daftar lokal, supaya `total` tetap benar.
    await refreshList()
    await accountsStore.fetchQuota().catch(() => null)
  }

  /** Pindah fisik antar akun — transfer data nyata, beda dari pindah folder. */
  async function moveFile(fileId: string, destAccountId: string) {
    await api.post(`/files/${fileId}/move`, { dest_account_id: destAccountId })
    await refreshList()
    await Promise.all([
      accountsStore.fetchAccounts().catch(() => null),
      accountsStore.fetchQuota().catch(() => null)
    ])
  }

  /** URL unduh langsung; browser yang menstream, bukan JS. */
  function downloadUrl(fileId: string) {
    return `${api.baseURL}/files/${fileId}/download`
  }

  // ---- Pilihan & aksi massal ----

  const selectedFiles = computed(() =>
    files.value.filter(f => selectedFileIds.value.includes(f.id)))

  const selectedBytes = computed(() =>
    selectedFiles.value.reduce((sum, f) => sum + f.size_bytes, 0))

  const allVisibleSelected = computed(() =>
    files.value.length > 0 && files.value.every(f => selectedFileIds.value.includes(f.id)))

  function isSelected(fileId: string) {
    return selectedFileIds.value.includes(fileId)
  }

  function toggleSelection(fileId: string) {
    selectedFileIds.value = selectedFileIds.value.includes(fileId)
      ? selectedFileIds.value.filter(id => id !== fileId)
      : [...selectedFileIds.value, fileId]
  }

  /** Pilih rentang dari jangkar terakhir — perilaku shift-click daftar berkas. */
  function selectRange(fromId: string, toId: string) {
    const ids = files.value.map(f => f.id)
    const start = ids.indexOf(fromId)
    const end = ids.indexOf(toId)
    if (start === -1 || end === -1) return
    const slice = ids.slice(Math.min(start, end), Math.max(start, end) + 1)
    const merged = new Set([...selectedFileIds.value, ...slice])
    selectedFileIds.value = ids.filter(id => merged.has(id))
  }

  function toggleSelectAll() {
    selectedFileIds.value = allVisibleSelected.value ? [] : files.value.map(f => f.id)
  }

  function clearSelection() {
    selectedFileIds.value = []
  }

  /**
   * Jalankan satu operasi atas tiap file terpilih, berurutan. Satu kegagalan tak
   * menghentikan sisanya; yang gagal dikembalikan agar UI bisa melaporkannya.
   */
  async function runBulk(
    ids: string[],
    op: (fileId: string) => Promise<unknown>
  ): Promise<{ ok: number; failed: { id: string; error: unknown }[] }> {
    const failed: { id: string; error: unknown }[] = []
    let ok = 0
    for (const id of ids) {
      try {
        await op(id)
        ok++
      } catch (error) {
        failed.push({ id, error })
      }
    }
    return { ok, failed }
  }

  /** Hapus seluruh file terpilih — objek fisik di provider ikut terhapus. */
  async function deleteSelected() {
    const ids = [...selectedFileIds.value]
    const res = await runBulk(ids, id => api.del(`/files/${id}`))
    clearSelection()
    await refreshList()
    await accountsStore.fetchQuota().catch(() => null)
    return res
  }

  /** Pindahkan seluruh file terpilih ke satu folder virtual (transaksi DB murni). */
  async function moveSelectedToFolder(targetFolderId: string | null) {
    const ids = [...selectedFileIds.value]
    const res = await runBulk(ids, id =>
      api.patch<FileEntry>(`/files/${id}`, { folder_id: targetFolderId }))
    clearSelection()
    await refreshList()
    return res
  }

  /** Migrasikan seluruh file terpilih ke akun lain — transfer data nyata. */
  async function migrateSelected(destAccountId: string) {
    const ids = [...selectedFileIds.value]
    const res = await runBulk(ids, id =>
      api.post(`/files/${id}/move`, { dest_account_id: destAccountId }))
    clearSelection()
    await refreshList()
    await Promise.all([
      accountsStore.fetchAccounts().catch(() => null),
      accountsStore.fetchQuota().catch(() => null)
    ])
    return res
  }

  // ---- Upload ----

  const uploadControllers = new Map<string, AbortController>()

  function randomId() {
    return Math.random().toString(36).slice(2, 10) + Date.now().toString(36)
  }

  /**
   * Unggah satu file: buka SSE lebih dulu memakai job id yang kita tentukan,
   * lalu kirim byte-nya. Backend memilih akun tujuan (smart routing) — UI tak
   * menentukannya, hanya menampilkan hasil keputusan router.
   */
  async function uploadFile(file: File, targetFolderId: string | null = currentFolderId.value) {
    const jobId = randomId()
    const folder = targetFolderId ? folders.value.find(f => f.id === targetFolderId) : null

    uploadJobs.value = [{
      id: jobId,
      file_name: file.name,
      size_bytes: file.size,
      bytes_uploaded: 0,
      progress: 0,
      status: 'routing',
      target_folder_id: targetFolderId,
      target_folder_path: folder?.path || '/'
    }, ...uploadJobs.value]

    // Ambil kembali entri dari ref supaya yang dimutasi adalah proxy reaktifnya,
    // bukan objek mentah yang tadi dibuat. Memutasi objek mentah tak terpantau
    // Vue: kartu upload membeku di "uploading" sampai sesuatu yang lain memicu
    // render — persis kenapa job sebelumnya baru tampak selesai saat file
    // berikutnya diunggah.
    const job = uploadJobs.value.find(j => j.id === jobId)!

    const events = subscribeUploadProgress(jobId, job)
    const controller = new AbortController()
    uploadControllers.set(jobId, controller)

    try {
      const query: Record<string, any> = { name: file.name, job_id: jobId }
      if (targetFolderId) query.folder_id = targetFolderId

      job.status = 'uploading'
      const res = await api.request<UploadResponse>('/files/upload', {
        method: 'POST',
        query,
        body: file,
        headers: { 'Content-Type': file.type || 'application/octet-stream' },
        signal: controller.signal
      })

      job.status = 'completed'
      job.progress = 100
      job.bytes_uploaded = file.size
      job.target_account_id = res.account_id
      job.target_account_label = res.account_label
      job.routing_strategy = res.routed_by
      job.target_provider = res.file.provider as StorageProvider

      // Tampilkan file baru bila memang milik folder yang sedang dibuka.
      if (!searchQuery.value.trim() && res.file.folder_id === currentFolderId.value) {
        files.value = [res.file, ...files.value]
      }
      await accountsStore.fetchQuota().catch(() => null)
      return res
    } catch (err: any) {
      // Pembatalan oleh user bukan kegagalan yang perlu dilaporkan sebagai error.
      if (controller.signal.aborted) {
        job.status = 'cancelled'
      } else {
        job.status = 'error'
        job.error_message = friendlyMessage(err)
      }
      throw err
    } finally {
      events.close()
      uploadControllers.delete(jobId)
    }
  }

  /** Ikuti progres nyata dari backend (docs/06 §Events). */
  function subscribeUploadProgress(jobId: string, job: UploadJob) {
    if (typeof EventSource === 'undefined') {
      return { close: () => {} }
    }
    const source = new EventSource(`${api.baseURL}/events/uploads/${jobId}`)

    source.addEventListener('progress', (ev) => {
      try {
        const data = JSON.parse((ev as MessageEvent).data)
        job.bytes_uploaded = data.bytes
        const total = data.total || job.size_bytes
        if (total > 0) {
          job.progress = Math.min(99, Math.round((data.bytes / total) * 100))
        }
      } catch {
        // Event rusak diabaikan: progres berikutnya akan menyusul.
      }
    })
    source.addEventListener('done', () => source.close())
    source.addEventListener('error', () => source.close())

    return { close: () => source.close() }
  }

  async function uploadFiles(fileList: File[], targetFolderId: string | null = currentFolderId.value) {
    const results = []
    for (const file of fileList) {
      // Berurutan, bukan paralel: tiap upload menahan satu koneksi stream ke
      // provider, dan router perlu kuota terbaru untuk keputusan berikutnya.
      results.push(await uploadFile(file, targetFolderId).catch(err => err))
    }
    return results
  }

  function cancelUpload(jobId: string) {
    uploadControllers.get(jobId)?.abort()
  }

  function dismissJob(jobId: string) {
    uploadJobs.value = uploadJobs.value.filter(j => j.id !== jobId)
  }

  function clearFinishedJobs() {
    uploadJobs.value = uploadJobs.value.filter(
      j => j.status !== 'completed' && j.status !== 'cancelled')
  }

  return {
    folders,
    files,
    currentFolderId,
    currentFolder,
    currentPath,
    breadcrumbs,
    currentFolders,
    allFoldersHierarchical,
    searchQuery,
    selectedAccountId,
    selectedCategory,
    viewMode,
    sortBy,
    sortDirection,
    page,
    perPage,
    totalFiles,
    totalPages,
    pageStart,
    pageEnd,
    selectedFileIds,
    selectedFiles,
    selectedBytes,
    allVisibleSelected,
    uploadJobs,
    isUploadModalOpen,
    isNewFolderModalOpen,
    isLoading,
    isSearching,
    loadError,
    filteredFiles,
    loadAll,
    fetchFolders,
    fetchFiles,
    refreshList,
    goToPage,
    applyFilters,
    isSelected,
    toggleSelection,
    selectRange,
    toggleSelectAll,
    clearSelection,
    deleteSelected,
    moveSelectedToFolder,
    migrateSelected,
    runSearch,
    fetchAllFiles,
    navigateToFolder,
    navigateUp,
    createFolder,
    renameFolder,
    moveFolder,
    deleteFolder,
    moveFileToFolder,
    renameFile,
    deleteFile,
    moveFile,
    downloadUrl,
    uploadFile,
    uploadFiles,
    cancelUpload,
    dismissJob,
    clearFinishedJobs
  }
})
