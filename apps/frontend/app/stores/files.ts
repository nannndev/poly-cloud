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
   * Filter & urutan yang dijalankan di klien atas hasil yang sudah dimuat.
   * Pencarian teks dan paging dikerjakan backend (query DB); yang di sini hanya
   * penyaringan ringan atas daftar yang sedang tampil.
   */
  const filteredFiles = computed(() => {
    let result = [...files.value]

    if (selectedAccountId.value !== 'all') {
      result = result.filter(f => f.account_id === selectedAccountId.value)
    }

    if (selectedCategory.value !== 'all') {
      result = result.filter(f => categoryOf(f) === selectedCategory.value)
    }

    const dir = sortDirection.value === 'asc' ? 1 : -1
    result.sort((a, b) => {
      if (sortBy.value === 'size') return (a.size_bytes - b.size_bytes) * dir
      if (sortBy.value === 'modified') {
        return ((new Date(a.modified_at || 0).getTime()) - (new Date(b.modified_at || 0).getTime())) * dir
      }
      return a.name.localeCompare(b.name) * dir
    })
    return result
  })

  /** Kategori untuk filter cepat di explorer. MIME dipercaya lebih dulu,
   *  ekstensi jadi cadangan bagi file yang providernya tak melaporkan tipe. */
  function categoryOf(file: FileEntry): string {
    const mime = (file.mime || '').toLowerCase()
    const name = file.name.toLowerCase()

    if (mime.startsWith('image/') || mime.startsWith('video/') || mime.startsWith('audio/')
      || /\.(mp4|webm|mov|mkv|png|jpe?g|gif|webp|svg|bmp|mp3|wav|flac|m4a)$/.test(name)) {
      return 'media'
    }
    if (mime.includes('zip') || mime.includes('tar') || mime.includes('gzip') || mime.includes('compressed')
      || /\.(zip|tar|gz|tgz|rar|7z|bz2)$/.test(name)) {
      return 'archives'
    }
    if (mime.includes('json') || mime.includes('yaml') || mime.includes('xml') || mime.includes('sql')
      || mime.includes('javascript') || mime.includes('x-sh') || mime.includes('text/x-')
      || /\.(ya?ml|toml|sql|ts|tsx|jsx?|go|py|rb|java|c|h|cpp|rs|php|sh|vue|css|html)$/.test(name)) {
      return 'code'
    }
    if (mime.includes('pdf') || mime.includes('word') || mime.includes('sheet') || mime.includes('presentation')
      || mime.includes('document') || mime.includes('csv') || mime.startsWith('text/')
      || /\.(pdf|docx?|xlsx?|pptx?|txt|log|md|csv|tsv)$/.test(name)) {
      return 'docs'
    }
    return 'other'
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

  async function fetchFiles() {
    const query: Record<string, any> = { per_page: 500 }
    if (currentFolderId.value) query.folder_id = currentFolderId.value
    const res = await api.get<FileListResult>('/files', query)
    files.value = res.items
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
   * Seluruh file milik user lintas folder — dipakai halaman statistik.
   * Endpoint search tanpa kata kunci mengembalikan semua baris index.
   */
  async function fetchAllFiles(): Promise<FileEntry[]> {
    const res = await api.get<FileListResult>('/files/search', { per_page: 500 })
    return res.items
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
      const query: Record<string, any> = { q, per_page: 500 }
      if (selectedAccountId.value !== 'all') query.account_id = selectedAccountId.value
      const res = await api.get<FileListResult>('/files/search', query)
      files.value = res.items
    } finally {
      isSearching.value = false
    }
  }

  async function navigateToFolder(folderId: string | null) {
    currentFolderId.value = folderId
    searchQuery.value = ''
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
    await fetchFiles()
  }

  async function moveFolder(folderId: string, newParentId: string | null) {
    await api.patch<FolderEntry>(`/folders/${folderId}`, { parent_id: newParentId })
    await fetchFolders()
    await fetchFiles()
  }

  async function deleteFolder(folderId: string, recursive = false) {
    await api.del(`/folders/${folderId}`, { recursive })

    // Kalau sedang berada di dalam folder yang dihapus, naik ke induknya.
    const deleted = folders.value.find(f => f.id === folderId)
    if (deleted && currentFolder.value?.path.startsWith(deleted.path)) {
      currentFolderId.value = deleted.parent_id
    }
    await fetchFolders()
    await fetchFiles()
    if (recursive) await accountsStore.fetchQuota().catch(() => null)
  }

  async function moveFileToFolder(fileId: string, targetFolderId: string | null) {
    const updated = await api.patch<FileEntry>(`/files/${fileId}`, { folder_id: targetFolderId })
    applyFileUpdate(updated)
    // File yang pindah keluar dari folder aktif tak lagi tampil di sini.
    if (!searchQuery.value.trim() && updated.folder_id !== currentFolderId.value) {
      files.value = files.value.filter(f => f.id !== fileId)
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
    files.value = files.value.filter(f => f.id !== fileId)
    await accountsStore.fetchQuota().catch(() => null)
  }

  /** Pindah fisik antar akun — transfer data nyata, beda dari pindah folder. */
  async function moveFile(fileId: string, destAccountId: string) {
    await api.post(`/files/${fileId}/move`, { dest_account_id: destAccountId })
    await fetchFiles()
    await Promise.all([
      accountsStore.fetchAccounts().catch(() => null),
      accountsStore.fetchQuota().catch(() => null)
    ])
  }

  /** URL unduh langsung; browser yang menstream, bukan JS. */
  function downloadUrl(fileId: string) {
    return `${api.baseURL}/files/${fileId}/download`
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

    const job: UploadJob = {
      id: jobId,
      file_name: file.name,
      size_bytes: file.size,
      bytes_uploaded: 0,
      progress: 0,
      status: 'routing',
      target_folder_id: targetFolderId,
      target_folder_path: folder?.path || '/'
    }
    uploadJobs.value = [job, ...uploadJobs.value]

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
