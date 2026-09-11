import { defineStore } from 'pinia'
import type { FileEntry, FolderEntry, UploadJob, StorageProvider } from '~/types'

export const useFilesStore = defineStore('files', () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase || 'http://localhost:8080/api/v1'

  // VFS Virtual Folders (DB-level hierarchy)
  const folders = ref<FolderEntry[]>([
    {
      id: 'fol-1',
      name: 'Documents',
      path: '/Documents',
      parent_id: null,
      created_at: '2026-09-01T10:00:00Z'
    },
    {
      id: 'fol-2',
      name: 'Finance',
      path: '/Documents/Finance',
      parent_id: 'fol-1',
      created_at: '2026-09-02T11:00:00Z'
    },
    {
      id: 'fol-3',
      name: 'CS',
      path: '/Documents/CS',
      parent_id: 'fol-1',
      created_at: '2026-09-03T12:00:00Z'
    },
    {
      id: 'fol-4',
      name: 'Design',
      path: '/Design',
      parent_id: null,
      created_at: '2026-09-04T09:00:00Z'
    },
    {
      id: 'fol-5',
      name: 'Marketing',
      path: '/Design/Marketing',
      parent_id: 'fol-4',
      created_at: '2026-09-04T10:30:00Z'
    },
    {
      id: 'fol-6',
      name: 'Architecture',
      path: '/Architecture',
      parent_id: null,
      created_at: '2026-09-05T08:00:00Z'
    },
    {
      id: 'fol-7',
      name: 'Docs',
      path: '/Architecture/Docs',
      parent_id: 'fol-6',
      created_at: '2026-09-05T08:30:00Z'
    },
    {
      id: 'fol-8',
      name: 'Backups',
      path: '/Backups',
      parent_id: null,
      created_at: '2026-09-06T07:00:00Z'
    },
    {
      id: 'fol-9',
      name: 'Database',
      path: '/Backups/Database',
      parent_id: 'fol-8',
      created_at: '2026-09-06T07:15:00Z'
    },
    {
      id: 'fol-10',
      name: 'Media',
      path: '/Media',
      parent_id: null,
      created_at: '2026-09-07T14:00:00Z'
    },
    {
      id: 'fol-11',
      name: 'Videos',
      path: '/Media/Videos',
      parent_id: 'fol-10',
      created_at: '2026-09-07T14:20:00Z'
    },
    {
      id: 'fol-12',
      name: 'Presentations',
      path: '/Presentations',
      parent_id: null,
      created_at: '2026-09-08T15:00:00Z'
    },
    {
      id: 'fol-13',
      name: 'DevOps',
      path: '/DevOps',
      parent_id: null,
      created_at: '2026-09-09T16:00:00Z'
    },
    {
      id: 'fol-14',
      name: 'Kube',
      path: '/DevOps/Kube',
      parent_id: 'fol-13',
      created_at: '2026-09-09T16:30:00Z'
    }
  ])

  // Files indexed across multi-cloud storage
  const files = ref<FileEntry[]>([
    {
      id: 'f-100',
      name: 'PolyCloud_Getting_Started_Guide.pdf',
      path: '/',
      virtual_path: '/PolyCloud_Getting_Started_Guide.pdf',
      folder_id: null,
      mime: 'application/pdf',
      size_bytes: 1420000,
      modified_at: '2026-09-11T09:00:00Z',
      account_id: 'acc-1',
      account_label: 'Google Drive Primary',
      provider: 'gdrive',
      is_chunked: false
    },
    {
      id: 'f-101',
      name: 'Q3_Financial_Forecast_Consolidated.xlsx',
      path: '/Documents/Finance',
      virtual_path: '/Documents/Finance/Q3_Financial_Forecast_Consolidated.xlsx',
      folder_id: 'fol-2',
      mime: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      size_bytes: 4820000, // ~4.8 MB
      modified_at: '2026-09-10T14:32:00Z',
      account_id: 'acc-2',
      account_label: 'OneDrive Business',
      provider: 'onedrive',
      is_chunked: false
    },
    {
      id: 'f-102',
      name: 'Brand_Identity_Assets_v4_Master.zip',
      path: '/Design/Marketing',
      virtual_path: '/Design/Marketing/Brand_Identity_Assets_v4_Master.zip',
      folder_id: 'fol-5',
      mime: 'application/zip',
      size_bytes: 284000000, // ~284 MB
      modified_at: '2026-09-09T18:10:00Z',
      account_id: 'acc-1',
      account_label: 'Google Drive Primary',
      provider: 'gdrive',
      is_chunked: true
    },
    {
      id: 'f-103',
      name: 'Enterprise_Architecture_Blueprint_2026.pdf',
      path: '/Architecture/Docs',
      virtual_path: '/Architecture/Docs/Enterprise_Architecture_Blueprint_2026.pdf',
      folder_id: 'fol-7',
      mime: 'application/pdf',
      size_bytes: 12500000, // 12.5 MB
      modified_at: '2026-09-08T09:15:00Z',
      account_id: 'acc-1',
      account_label: 'Google Drive Primary',
      provider: 'gdrive',
      is_chunked: false
    },
    {
      id: 'f-104',
      name: 'PostgreSQL_Dump_Production_Snapshot.sql.gz',
      path: '/Backups/Database',
      virtual_path: '/Backups/Database/PostgreSQL_Dump_Production_Snapshot.sql.gz',
      folder_id: 'fol-9',
      mime: 'application/gzip',
      size_bytes: 1420000000, // ~1.42 GB
      modified_at: '2026-09-11T03:00:00Z',
      account_id: 'acc-4',
      account_label: 'AWS S3 Cold Bucket (us-east-1)',
      provider: 's3',
      is_chunked: true
    },
    {
      id: 'f-105',
      name: 'Product_Demo_4K_60FPS_Showcase.mp4',
      path: '/Media/Videos',
      virtual_path: '/Media/Videos/Product_Demo_4K_60FPS_Showcase.mp4',
      folder_id: 'fol-11',
      mime: 'video/mp4',
      size_bytes: 840000000, // ~840 MB
      modified_at: '2026-09-07T11:45:00Z',
      account_id: 'acc-5',
      account_label: 'Cloudflare R2 Media Hot',
      provider: 'r2',
      is_chunked: true
    },
    {
      id: 'f-106',
      name: 'Team_Offsite_Keynote_Deck.pdf',
      path: '/Presentations',
      virtual_path: '/Presentations/Team_Offsite_Keynote_Deck.pdf',
      folder_id: 'fol-12',
      mime: 'application/pdf',
      size_bytes: 18900000, // ~18.9 MB
      modified_at: '2026-09-06T16:20:00Z',
      account_id: 'acc-3',
      account_label: 'Dropbox Team Share',
      provider: 'dropbox',
      is_chunked: false
    },
    {
      id: 'f-107',
      name: 'Docker_Compose_Cluster_Setup.yaml',
      path: '/DevOps/Kube',
      virtual_path: '/DevOps/Kube/Docker_Compose_Cluster_Setup.yaml',
      folder_id: 'fol-14',
      mime: 'text/yaml',
      size_bytes: 45200,
      modified_at: '2026-09-11T08:14:00Z',
      account_id: 'acc-2',
      account_label: 'OneDrive Business',
      provider: 'onedrive',
      is_chunked: false
    },
    {
      id: 'f-108',
      name: 'Customer_Success_Quarterly_Review.docx',
      path: '/Documents/CS',
      virtual_path: '/Documents/CS/Customer_Success_Quarterly_Review.docx',
      folder_id: 'fol-3',
      mime: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      size_bytes: 3200000,
      modified_at: '2026-09-05T13:00:00Z',
      account_id: 'acc-1',
      account_label: 'Google Drive Primary',
      provider: 'gdrive',
      is_chunked: false
    }
  ])

  // Current active VFS navigation location
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
  const isSearching = ref(false)

  // Current active folder object
  const currentFolder = computed(() => {
    if (!currentFolderId.value) return null
    return folders.value.find(f => f.id === currentFolderId.value) || null
  })

  // Materialized path of current directory
  const currentPath = computed(() => {
    return currentFolder.value ? currentFolder.value.path : '/'
  })

  // Breadcrumb path from Root down to current folder
  const breadcrumbs = computed(() => {
    const crumbs: { id: string | null; name: string; path: string }[] = [
      { id: null, name: 'Root', path: '/' }
    ]

    if (!currentFolder.value) return crumbs

    // Traverse ancestors up to root
    const trail: FolderEntry[] = []
    let curr: FolderEntry | null = currentFolder.value
    while (curr) {
      trail.unshift(curr)
      if (curr.parent_id) {
        curr = folders.value.find(f => f.id === curr!.parent_id) || null
      } else {
        curr = null
      }
    }

    trail.forEach(f => {
      crumbs.push({ id: f.id, name: f.name, path: f.path })
    })

    return crumbs
  })

  // Subfolders in current directory with dynamic item count
  const currentFolders = computed(() => {
    // If searching, hide direct subfolders or show matching folders
    const parentId = currentFolderId.value

    return folders.value
      .filter(f => {
        if (searchQuery.value.trim()) {
          return f.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                 f.path.toLowerCase().includes(searchQuery.value.toLowerCase())
        }
        return f.parent_id === parentId
      })
      .map(folder => {
        // Calculate item count (child folders + child files)
        const subCount = folders.value.filter(sub => sub.parent_id === folder.id).length
        const fileCount = files.value.filter(file => file.folder_id === folder.id).length
        return {
          ...folder,
          item_count: subCount + fileCount
        }
      })
      .sort((a, b) => a.name.localeCompare(b.name))
  })

  // Files in current directory or matching global search
  const filteredFiles = computed(() => {
    const isSearchActive = searchQuery.value.trim().length > 0
    const isCategoryFilterActive = selectedCategory.value !== 'all'

    return files.value.filter(file => {
      // If no search and no category filter, only show files in current directory
      if (!isSearchActive && !isCategoryFilterActive) {
        const fileFolderId = file.folder_id || null
        if (fileFolderId !== currentFolderId.value) {
          return false
        }
      }

      // Cloud account filter
      if (selectedAccountId.value !== 'all' && file.account_id !== selectedAccountId.value) {
        return false
      }

      // Category filter
      if (selectedCategory.value !== 'all') {
        const mime = (file.mime || '').toLowerCase()
        const name = file.name.toLowerCase()

        if (selectedCategory.value === 'docs') {
          const isDoc = mime.includes('pdf') || mime.includes('word') || mime.includes('sheet') || mime.includes('document') || name.endsWith('.txt') || name.endsWith('.md')
          if (!isDoc) return false
        } else if (selectedCategory.value === 'media') {
          const isMedia = mime.includes('image') || mime.includes('video') || mime.includes('audio') || name.endsWith('.mp4') || name.endsWith('.png') || name.endsWith('.jpg')
          if (!isMedia) return false
        } else if (selectedCategory.value === 'archives') {
          const isArchive = mime.includes('zip') || mime.includes('tar') || mime.includes('gz') || name.endsWith('.zip') || name.endsWith('.tar')
          if (!isArchive) return false
        } else if (selectedCategory.value === 'code') {
          const isCode = mime.includes('text') || mime.includes('json') || name.endsWith('.yaml') || name.endsWith('.sql') || name.endsWith('.ts')
          if (!isCode) return false
        }
      }

      // Search query filter (matches name, virtual path, account)
      if (isSearchActive) {
        const query = searchQuery.value.toLowerCase()
        const matchName = file.name.toLowerCase().includes(query)
        const matchPath = (file.virtual_path || file.path).toLowerCase().includes(query)
        const matchAccount = file.account_label.toLowerCase().includes(query)
        if (!matchName && !matchPath && !matchAccount) return false
      }

      return true
    }).sort((a, b) => {
      let comparison = 0
      if (sortBy.value === 'name') {
        comparison = a.name.localeCompare(b.name)
      } else if (sortBy.value === 'size') {
        comparison = a.size_bytes - b.size_bytes
      } else {
        comparison = new Date(a.modified_at).getTime() - new Date(b.modified_at).getTime()
      }
      return sortDirection.value === 'asc' ? comparison : -comparison
    })
  })

  // Hierarchical list of all folders for folder picker selection
  const allFoldersHierarchical = computed(() => {
    function buildTree(parentId: string | null, depth = 0): { id: string | null; name: string; path: string; depth: number }[] {
      const result: { id: string | null; name: string; path: string; depth: number }[] = []
      const children = folders.value.filter(f => f.parent_id === parentId).sort((a, b) => a.name.localeCompare(b.name))
      for (const child of children) {
        result.push({
          id: child.id,
          name: child.name,
          path: child.path,
          depth
        })
        result.push(...buildTree(child.id, depth + 1))
      }
      return result
    }

    return [
      { id: null, name: 'Root (/)', path: '/', depth: 0 },
      ...buildTree(null, 1)
    ]
  })

  // Navigation actions
  function navigateToFolder(folderId: string | null) {
    currentFolderId.value = folderId
    searchQuery.value = ''
  }

  function navigateUp() {
    if (!currentFolder.value) return
    currentFolderId.value = currentFolder.value.parent_id
  }

  // VFS Folder Operations (DB Transactions)
  function createFolder(name: string, parentId: string | null = currentFolderId.value): FolderEntry {
    const trimmedName = name.trim().replace(/^\/+|\/+$/g, '')
    const parent = folders.value.find(f => f.id === parentId)
    const newPath = parent ? `${parent.path}/${trimmedName}` : `/${trimmedName}`

    // Check if path exists
    const exists = folders.value.some(f => f.path.toLowerCase() === newPath.toLowerCase())
    if (exists) {
      throw new Error(`A folder named "${trimmedName}" already exists at this path.`)
    }

    const newFolder: FolderEntry = {
      id: 'fol-' + Math.random().toString(36).substring(2, 9),
      parent_id: parentId,
      name: trimmedName,
      path: newPath,
      created_at: new Date().toISOString()
    }

    folders.value.push(newFolder)
    return newFolder
  }

  function renameFolder(folderId: string, newName: string) {
    const folder = folders.value.find(f => f.id === folderId)
    if (!folder) return

    const trimmed = newName.trim().replace(/^\/+|\/+$/g, '')
    const oldPath = folder.path
    const parent = folders.value.find(f => f.id === folder.parent_id)
    const newPath = parent ? `${parent.path}/${trimmed}` : `/${trimmed}`

    folder.name = trimmed
    folder.path = newPath

    // Cascade materialized path update to all descendant folders
    folders.value.forEach(f => {
      if (f.path.startsWith(oldPath + '/')) {
        f.path = f.path.replace(oldPath, newPath)
      }
    })

    // Cascade materialized path update to all files under this folder hierarchy
    files.value.forEach(file => {
      if (file.path.startsWith(oldPath)) {
        file.path = file.path.replace(oldPath, newPath)
        if (file.virtual_path) {
          file.virtual_path = file.virtual_path.replace(oldPath, newPath)
        }
      }
    })
  }

  function deleteFolder(folderId: string, recursive: boolean = true) {
    const folderToDelete = folders.value.find(f => f.id === folderId)
    if (!folderToDelete) return

    const targetPath = folderToDelete.path

    if (recursive) {
      // Remove all descendant files
      files.value = files.value.filter(file => !file.path.startsWith(targetPath))
      // Remove all descendant folders including target
      folders.value = folders.value.filter(f => !f.path.startsWith(targetPath))
    } else {
      // Delete single folder only if empty
      const hasChildFolders = folders.value.some(f => f.parent_id === folderId)
      const hasChildFiles = files.value.some(f => f.folder_id === folderId)
      if (hasChildFolders || hasChildFiles) {
        throw new Error('Folder is not empty. Use recursive delete.')
      }
      folders.value = folders.value.filter(f => f.id !== folderId)
    }

    // Reset current location if we were inside deleted folder
    if (currentFolder.value && currentFolder.value.path.startsWith(targetPath)) {
      currentFolderId.value = folderToDelete.parent_id
    }
  }

  // VFS File Move (Instant Virtual Organization - Pure DB, No Physical Remote Transfer)
  function moveFileToFolder(fileId: string, targetFolderId: string | null) {
    const file = files.value.find(f => f.id === fileId)
    if (!file) return

    const targetFolder = targetFolderId ? folders.value.find(f => f.id === targetFolderId) : null
    const destPath = targetFolder ? targetFolder.path : '/'

    file.folder_id = targetFolderId
    file.path = destPath
    file.virtual_path = destPath === '/' ? `/${file.name}` : `${destPath}/${file.name}`
  }

  function renameFile(fileId: string, newName: string) {
    const file = files.value.find(f => f.id === fileId)
    if (!file) return

    file.name = newName
    file.virtual_path = file.path === '/' ? `/${newName}` : `${file.path}/${newName}`
  }

  // File Deletion (deletes from index & provider blocks)
  function deleteFile(fileId: string) {
    files.value = files.value.filter(f => f.id !== fileId)
  }

  // Physical Cross-Provider Migration (POST /files/{id}/move)
  function moveFile(fileId: string, destAccount: { id: string; label: string; provider: StorageProvider }) {
    const file = files.value.find(f => f.id === fileId)
    if (file) {
      file.account_id = destAccount.id
      file.account_label = destAccount.label
      file.provider = destAccount.provider
      file.modified_at = new Date().toISOString()
    }
  }

  const uploadIntervals = new Map<string, any>()

  function runUploadProgress(job: UploadJob, target: { id: string; label: string; provider: StorageProvider }) {
    job.status = 'uploading'
    job.speed_mbps = Number((1.5 + Math.random() * 2.8).toFixed(1))

    const interval = setInterval(() => {
      if (job.status === 'paused') {
        clearInterval(interval)
        return
      }

      if (job.progress < 100) {
        const step = Math.floor(Math.random() * 8) + 5
        job.progress = Math.min(100, job.progress + step)
        job.bytes_uploaded = Math.min(job.size_bytes, Math.floor((job.progress / 100) * job.size_bytes))
        job.speed_mbps = Number((1.2 + Math.random() * 3.2).toFixed(1))
      }

      if (job.progress >= 100) {
        clearInterval(interval)
        uploadIntervals.delete(job.id)
        job.progress = 100
        job.bytes_uploaded = job.size_bytes
        job.status = 'completed'

        // Determine mime
        const name = job.file_name.toLowerCase()
        let mime = 'application/octet-stream'
        if (name.endsWith('.xlsx') || name.endsWith('.xls')) mime = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        else if (name.endsWith('.pdf')) mime = 'application/pdf'
        else if (name.endsWith('.png')) mime = 'image/png'
        else if (name.endsWith('.jpg') || name.endsWith('.jpeg')) mime = 'image/jpeg'
        else if (name.endsWith('.zip')) mime = 'application/zip'
        else if (name.endsWith('.mp4')) mime = 'video/mp4'
        else if (name.endsWith('.csv')) mime = 'text/csv'

        const targetFolder = job.target_folder_id ? folders.value.find(f => f.id === job.target_folder_id) : null
        const destPath = targetFolder ? targetFolder.path : (job.target_folder_path || '/')
        const vPath = destPath === '/' ? `/${job.file_name}` : `${destPath}/${job.file_name}`

        // Add to files list
        files.value.unshift({
          id: 'f-' + Math.random().toString(36).substring(2, 8),
          name: job.file_name,
          path: destPath,
          virtual_path: vPath,
          folder_id: job.target_folder_id ?? null,
          mime,
          size_bytes: job.size_bytes,
          modified_at: new Date().toISOString(),
          account_id: target.id,
          account_label: target.label,
          provider: target.provider,
          is_chunked: job.size_bytes > 100 * 1024 * 1024
        })
      }
    }, 280)

    uploadIntervals.set(job.id, interval)
  }

  function simulateUpload(fileData: {
    name: string
    size: number
    targetAccount?: { id: string; label: string; provider: StorageProvider }
    targetFolderId?: string | null
    targetFolderPath?: string
  }) {
    const jobId = 'job-' + Math.random().toString(36).substring(2, 9)

    // Smart routing decision if not specified
    const target = fileData.targetAccount || {
      id: 'acc-2',
      label: 'OneDrive Business',
      provider: 'onedrive' as StorageProvider
    }

    const folderId = fileData.targetFolderId !== undefined ? fileData.targetFolderId : currentFolderId.value
    const folderPath = fileData.targetFolderPath || currentPath.value

    const newJob: UploadJob = {
      id: jobId,
      file_name: fileData.name,
      size_bytes: fileData.size,
      bytes_uploaded: 0,
      progress: 0,
      status: 'routing',
      target_account_id: target.id,
      target_account_label: target.label,
      target_provider: target.provider,
      target_folder_id: folderId,
      target_folder_path: folderPath,
      routing_strategy: 'most-free-capacity',
      speed_mbps: 2.4
    }

    uploadJobs.value.unshift(newJob)

    // Simulate smart routing brief delay -> then progress
    setTimeout(() => {
      runUploadProgress(newJob, target)
    }, 400)

    return newJob
  }

  function pauseUpload(jobId: string) {
    const job = uploadJobs.value.find(j => j.id === jobId)
    if (job && job.status === 'uploading') {
      job.status = 'paused'
      const interval = uploadIntervals.get(jobId)
      if (interval) {
        clearInterval(interval)
        uploadIntervals.delete(jobId)
      }
    }
  }

  function resumeUpload(jobId: string) {
    const job = uploadJobs.value.find(j => j.id === jobId)
    if (job && job.status === 'paused') {
      const target = {
        id: job.target_account_id || 'acc-2',
        label: job.target_account_label || 'OneDrive Business',
        provider: job.target_provider || 'onedrive'
      }
      runUploadProgress(job, target)
    }
  }

  function cancelUpload(jobId: string) {
    const interval = uploadIntervals.get(jobId)
    if (interval) {
      clearInterval(interval)
      uploadIntervals.delete(jobId)
    }
    uploadJobs.value = uploadJobs.value.filter(j => j.id !== jobId)
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
    isSearching,
    filteredFiles,
    navigateToFolder,
    navigateUp,
    createFolder,
    renameFolder,
    deleteFolder,
    moveFileToFolder,
    renameFile,
    deleteFile,
    moveFile,
    simulateUpload,
    pauseUpload,
    resumeUpload,
    cancelUpload
  }
})
