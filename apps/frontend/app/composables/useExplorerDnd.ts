import type { FileEntry, FolderEntry } from '~/types'

/** Apa yang sedang diseret di dalam explorer. */
interface DragPayload {
  kind: 'file' | 'folder'
  /** Id yang ikut berpindah. Untuk file bisa lebih dari satu (pilihan massal). */
  ids: string[]
  /** Ditampilkan pada penanda drop. */
  label: string
}

/**
 * Seret-lepas di dalam explorer: memindahkan file dan folder ke folder lain.
 *
 * Terpisah dari useDropZone, yang menangani berkas dari LUAR peramban. Keduanya
 * dibedakan lewat `dataTransfer.types`: seretan internal membawa tipe kustom di
 * bawah, seretan sistem berkas membawa "Files". Tanpa pemisahan itu, menyeret
 * satu baris tabel akan memunculkan overlay unggah.
 */
const MIME = 'application/x-polycloud-item'

export function useExplorerDnd() {
  const filesStore = useFilesStore()
  const toast = useToast()

  /** Sumber kebenaran selama seretan berlangsung. */
  const dragging = ref<DragPayload | null>(null)

  /** Folder yang sedang disorot sebagai calon tujuan; null = tak ada. */
  const dropTargetId = ref<string | null | undefined>(undefined)

  const isDragging = computed(() => dragging.value !== null)

  // ---- Sisi sumber ----

  function startFileDrag(event: DragEvent, file: FileEntry) {
    // Menyeret baris yang termasuk pilihan memindahkan SELURUH pilihan —
    // perilaku yang sama dengan penjelajah berkas mana pun. Menyeret baris di
    // luar pilihan hanya memindahkan baris itu.
    const selected = filesStore.selectedFileIds
    const inSelection = selected.includes(file.id)
    const ids = inSelection && selected.length > 0 ? [...selected] : [file.id]

    dragging.value = {
      kind: 'file',
      ids,
      label: ids.length > 1 ? `${ids.length} files` : file.name
    }
    markDrag(event)
  }

  function startFolderDrag(event: DragEvent, folder: FolderEntry) {
    dragging.value = { kind: 'folder', ids: [folder.id], label: folder.name }
    markDrag(event)
  }

  function markDrag(event: DragEvent) {
    if (!event.dataTransfer) return
    // Isi payload tak dibaca kembali — status sebenarnya ada di `dragging`.
    // Yang penting hanya TIPE-nya, sebagai penanda bahwa ini seretan internal.
    event.dataTransfer.setData(MIME, '1')
    event.dataTransfer.effectAllowed = 'move'
  }

  function endDrag() {
    dragging.value = null
    dropTargetId.value = undefined
  }

  // ---- Sisi tujuan ----

  function isInternalDrag(event: DragEvent) {
    return Array.from(event.dataTransfer?.types ?? []).includes(MIME)
  }

  /**
   * Apakah item yang sedang diseret boleh masuk ke folder ini?
   *
   * Sebuah folder tak boleh dipindahkan ke dalam dirinya sendiri maupun ke
   * turunannya — itu memutus pohon dan membuat cabangnya tak terjangkau.
   * Backend juga menolaknya, tapi menyaring di sini membuat tujuan terlarang
   * tak pernah tampak bisa dijatuhi.
   */
  function canDropInto(targetFolderId: string | null): boolean {
    const drag = dragging.value
    if (!drag) return false

    if (drag.kind === 'folder') {
      const movingId = drag.ids[0]!
      if (movingId === targetFolderId) return false

      const moving = filesStore.folders.find(f => f.id === movingId)
      if (!moving) return false

      // Induknya sekarang: menjatuhkan di situ tak mengubah apa pun.
      if (moving.parent_id === targetFolderId) return false

      if (targetFolderId !== null) {
        const target = filesStore.folders.find(f => f.id === targetFolderId)
        if (!target) return false
        const prefix = moving.path.endsWith('/') ? moving.path : `${moving.path}/`
        if (target.path === moving.path || target.path.startsWith(prefix)) return false
      }
      return true
    }

    // File: menjatuhkan ke folder yang sedang dibuka tak memindahkan apa pun.
    return targetFolderId !== filesStore.currentFolderId
  }

  function onDragOverFolder(event: DragEvent, targetFolderId: string | null) {
    if (!isInternalDrag(event) || !canDropInto(targetFolderId)) return
    event.preventDefault()
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'move'
    dropTargetId.value = targetFolderId
  }

  function onDragLeaveFolder(targetFolderId: string | null) {
    if (dropTargetId.value === targetFolderId) dropTargetId.value = undefined
  }

  function isDropTarget(targetFolderId: string | null) {
    return dropTargetId.value === targetFolderId && canDropInto(targetFolderId)
  }

  /** Kembalikan satu pemindahan ke keadaan sebelumnya. */
  async function undoMove(
    kind: 'file' | 'folder',
    origins: { fileId: string; folderId: string | null }[],
    label: string
  ) {
    try {
      if (kind === 'folder') {
        const o = origins[0]!
        await filesStore.moveFolder(o.fileId, o.folderId)
      } else {
        const res = await filesStore.restoreFilesToFolders(origins)
        if (res.failed.length > 0) {
          toast.add({
            title: `Could not fully undo — ${res.failed.length} item(s) stayed put`,
            description: friendlyMessage(res.failed[0]!.error),
            color: 'warning'
          })
          return
        }
      }
      toast.add({ title: `Move of ${label} undone`, color: 'neutral' })
    } catch (err) {
      toast.add({
        title: 'Could not undo the move',
        description: friendlyMessage(err),
        color: 'error'
      })
    }
  }

  /**
   * Jatuhkan ke folder tujuan. Mengembalikan true bila sesuatu benar-benar
   * dipindahkan, supaya pemanggil tahu apakah perlu menyegarkan tampilan.
   */
  async function dropInto(event: DragEvent, targetFolderId: string | null) {
    if (!isInternalDrag(event)) return false
    event.preventDefault()

    const drag = dragging.value
    const allowed = canDropInto(targetFolderId)
    endDrag()
    if (!drag || !allowed) return false

    const targetName = targetFolderId === null
      ? 'Root'
      : filesStore.folders.find(f => f.id === targetFolderId)?.name ?? 'folder'

    // Asal dicatat SEBELUM memindahkan — setelah pindah, informasi itu sudah
    // tertimpa dan tak ada lagi cara mengetahui ke mana harus dikembalikan.
    // Tiap item disimpan sendiri-sendiri: pada pilihan massal, file bisa berasal
    // dari folder yang berbeda-beda.
    const origins = drag.kind === 'folder'
      ? [{
          fileId: drag.ids[0]!,
          folderId: filesStore.folders.find(f => f.id === drag.ids[0])?.parent_id ?? null
        }]
      : drag.ids.map(id => ({
          fileId: id,
          folderId: filesStore.files.find(f => f.id === id)?.folder_id ?? null
        }))

    const undoAction = {
      label: 'Undo',
      color: 'neutral' as const,
      variant: 'outline' as const,
      onClick: (e: Event) => {
        e.stopPropagation()
        void undoMove(drag.kind, origins, drag.label)
      }
    }

    try {
      if (drag.kind === 'folder') {
        await filesStore.moveFolder(drag.ids[0]!, targetFolderId)
      } else if (drag.ids.length === 1) {
        await filesStore.moveFileToFolder(drag.ids[0]!, targetFolderId)
      } else {
        // Pilihan massal memakai jalur yang sudah melaporkan kegagalan sebagian.
        const res = await filesStore.moveSelectedToFolder(targetFolderId)
        if (res.failed.length > 0) {
          // Sebagian gagal: yang berhasil tetap bisa dibatalkan, tapi hanya
          // untuk item yang benar-benar berpindah.
          const movedIds = new Set(drag.ids.filter(
            id => !res.failed.some(f => f.id === id)))
          toast.add({
            title: `${res.failed.length} of ${res.ok + res.failed.length} could not be moved`,
            description: friendlyMessage(res.failed[0]!.error),
            color: res.ok > 0 ? 'warning' : 'error',
            actions: res.ok > 0
              ? [{
                  ...undoAction,
                  onClick: (e: Event) => {
                    e.stopPropagation()
                    void undoMove(
                      drag.kind,
                      origins.filter(o => movedIds.has(o.fileId)),
                      `${movedIds.size} file(s)`
                    )
                  }
                }]
              : undefined
          })
          return res.ok > 0
        }
      }

      toast.add({
        title: `${drag.label} moved to ${targetName}`,
        color: 'success',
        actions: [undoAction]
      })
      return true
    } catch (err) {
      toast.add({
        title: `Could not move ${drag.label}`,
        description: friendlyMessage(err),
        color: 'error'
      })
      return false
    }
  }

  return {
    isDragging,
    dragging,
    startFileDrag,
    startFolderDrag,
    endDrag,
    canDropInto,
    isDropTarget,
    onDragOverFolder,
    onDragLeaveFolder,
    dropInto
  }
}
