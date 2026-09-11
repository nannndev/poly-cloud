import type { StorageProvider } from '~/types'

export function useFormatters() {
  function formatBytes(bytes: number, decimals = 1): string {
    if (!bytes || bytes === 0) return '0 B'
    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return '-'
    const date = new Date(dateStr)
    if (isNaN(date.getTime())) return '-'
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date)
  }

  function getProviderMeta(provider: StorageProvider) {
    switch (provider) {
      case 'gdrive':
        return {
          name: 'Google Drive',
          icon: 'i-simple-icons-googledrive',
          color: 'emerald' as const,
          badgeColor: 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20'
        }
      case 'onedrive':
        return {
          name: 'OneDrive',
          icon: 'i-simple-icons-microsoftonedrive',
          color: 'sky' as const,
          badgeColor: 'text-sky-500 bg-sky-500/10 border-sky-500/20'
        }
      case 'dropbox':
        return {
          name: 'Dropbox',
          icon: 'i-simple-icons-dropbox',
          color: 'blue' as const,
          badgeColor: 'text-blue-500 bg-blue-500/10 border-blue-500/20'
        }
      case 's3':
        return {
          name: 'Amazon S3',
          icon: 'i-simple-icons-amazons3',
          color: 'amber' as const,
          badgeColor: 'text-amber-500 bg-amber-500/10 border-amber-500/20'
        }
      case 'r2':
        return {
          name: 'Cloudflare R2',
          icon: 'i-simple-icons-cloudflare',
          color: 'orange' as const,
          badgeColor: 'text-orange-500 bg-orange-500/10 border-orange-500/20'
        }
      case 'b2':
        return {
          name: 'Backblaze B2',
          icon: 'i-lucide-hard-drive',
          color: 'rose' as const,
          badgeColor: 'text-rose-500 bg-rose-500/10 border-rose-500/20'
        }
      default:
        return {
          name: 'Cloud Storage',
          icon: 'i-lucide-cloud',
          color: 'neutral' as const,
          badgeColor: 'text-zinc-500 bg-zinc-500/10 border-zinc-500/20'
        }
    }
  }

  function getFileIcon(mime: string | null, name: string) {
    const m = (mime || '').toLowerCase()
    const n = name.toLowerCase()

    if (m.includes('image') || n.endsWith('.png') || n.endsWith('.jpg') || n.endsWith('.jpeg') || n.endsWith('.svg') || n.endsWith('.webp')) {
      return { icon: 'i-lucide-image', color: 'text-purple-500' }
    }
    if (m.includes('video') || n.endsWith('.mp4') || n.endsWith('.mkv') || n.endsWith('.mov')) {
      return { icon: 'i-lucide-film', color: 'text-rose-500' }
    }
    if (m.includes('audio') || n.endsWith('.mp3') || n.endsWith('.wav')) {
      return { icon: 'i-lucide-music', color: 'text-pink-500' }
    }
    if (m.includes('pdf') || n.endsWith('.pdf')) {
      return { icon: 'i-lucide-file-text', color: 'text-red-500' }
    }
    if (m.includes('sheet') || n.endsWith('.xls') || n.endsWith('.xlsx') || n.endsWith('.csv')) {
      return { icon: 'i-lucide-file-spreadsheet', color: 'text-emerald-500' }
    }
    if (m.includes('word') || n.endsWith('.doc') || n.endsWith('.docx')) {
      return { icon: 'i-lucide-file-type', color: 'text-blue-500' }
    }
    if (m.includes('zip') || m.includes('tar') || m.includes('gzip') || n.endsWith('.zip') || n.endsWith('.gz')) {
      return { icon: 'i-lucide-file-archive', color: 'text-amber-500' }
    }
    if (n.endsWith('.sql') || n.endsWith('.db')) {
      return { icon: 'i-lucide-database', color: 'text-indigo-500' }
    }
    if (n.endsWith('.ts') || n.endsWith('.js') || n.endsWith('.json') || n.endsWith('.yaml') || n.endsWith('.yml')) {
      return { icon: 'i-lucide-file-code', color: 'text-cyan-500' }
    }
    return { icon: 'i-lucide-file', color: 'text-zinc-400' }
  }

  return {
    formatBytes,
    formatDate,
    getProviderMeta,
    getFileIcon
  }
}
