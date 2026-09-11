import { defineStore } from 'pinia'
import type { FileEntry, UploadJob, StorageProvider } from '~/types'

export const useFilesStore = defineStore('files', () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase || 'http://localhost:8080/api/v1'

  const files = ref<FileEntry[]>([
    {
      id: 'f-101',
      name: 'Q3_Financial_Forecast_Consolidated.xlsx',
      path: '/Documents/Finance',
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
      mime: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
      size_bytes: 3200000,
      modified_at: '2026-09-05T13:00:00Z',
      account_id: 'acc-1',
      account_label: 'Google Drive Primary',
      provider: 'gdrive',
      is_chunked: false
    }
  ])

  const searchQuery = ref('')
  const selectedAccountId = ref<string>('all')
  const selectedCategory = ref<string>('all')
  const viewMode = ref<'grid' | 'table'>('table')
  const sortBy = ref<'name' | 'size' | 'modified'>('modified')
  const sortDirection = ref<'asc' | 'desc'>('desc')
  const uploadJobs = ref<UploadJob[]>([])
  const isUploadModalOpen = ref(false)
  const isSearching = ref(false)

  const filteredFiles = computed(() => {
    return files.value.filter(file => {
      // Account filter
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

      // Search text
      if (searchQuery.value.trim()) {
        const query = searchQuery.value.toLowerCase()
        const matchName = file.name.toLowerCase().includes(query)
        const matchPath = file.path.toLowerCase().includes(query)
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

  function deleteFile(fileId: string) {
    files.value = files.value.filter(f => f.id !== fileId)
  }

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

        // Add to files list
        files.value.unshift({
          id: 'f-' + Math.random().toString(36).substring(2, 8),
          name: job.file_name,
          path: '/Uploads',
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

  function simulateUpload(fileData: { name: string; size: number; targetAccount?: { id: string; label: string; provider: StorageProvider } }) {
    const jobId = 'job-' + Math.random().toString(36).substring(2, 9)
    
    // Smart routing decision if not specified
    const target = fileData.targetAccount || {
      id: 'acc-2',
      label: 'OneDrive Business',
      provider: 'onedrive' as StorageProvider
    }

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
    files,
    searchQuery,
    selectedAccountId,
    selectedCategory,
    viewMode,
    sortBy,
    sortDirection,
    uploadJobs,
    isUploadModalOpen,
    isSearching,
    filteredFiles,
    deleteFile,
    moveFile,
    simulateUpload,
    pauseUpload,
    resumeUpload,
    cancelUpload
  }
})
