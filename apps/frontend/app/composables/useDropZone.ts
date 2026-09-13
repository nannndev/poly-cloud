/**
 * Deteksi file yang diseret dari luar peramban ke seluruh halaman.
 *
 * Tiga hal yang membuat ini tak sesederhana memasang @dragover:
 *
 * 1. `dragenter`/`dragleave` ikut terpicu saat kursor melintasi elemen ANAK,
 *    jadi menghitungnya sebagai boolean membuat overlay berkedip tiap kali
 *    kursor melewati baris tabel. Karena itu dipakai penghitung kedalaman.
 *
 * 2. Tanpa `preventDefault` pada dragover, peramban akan membuka file yang
 *    dilepas sebagai halaman baru — seluruh aplikasi tergantikan oleh gambar.
 *
 * 3. Seretan dari DALAM aplikasi (baris tabel, teks terpilih) juga memicu event
 *    yang sama. Overlay unggah hanya boleh muncul untuk file dari sistem berkas,
 *    yang dikenali lewat `types` berisi "Files".
 */
export function useDropZone(onDrop: (files: File[]) => void) {
  const isDraggingFiles = ref(false)
  let depth = 0

  function carriesFiles(e: DragEvent) {
    return Array.from(e.dataTransfer?.types ?? []).includes('Files')
  }

  function onDragEnter(e: DragEvent) {
    if (!carriesFiles(e)) return
    depth++
    isDraggingFiles.value = true
  }

  function onDragOver(e: DragEvent) {
    if (!carriesFiles(e)) return
    // Wajib: tanpa ini peramban menolak drop dan membuka filenya sendiri.
    e.preventDefault()
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy'
  }

  function onDragLeave(e: DragEvent) {
    if (!carriesFiles(e)) return
    depth = Math.max(0, depth - 1)
    if (depth === 0) isDraggingFiles.value = false
  }

  function onDropFiles(e: DragEvent) {
    if (!carriesFiles(e)) return
    e.preventDefault()
    depth = 0
    isDraggingFiles.value = false

    const files = Array.from(e.dataTransfer?.files ?? [])
    if (files.length > 0) onDrop(files)
  }

  /**
   * Seretan yang berakhir di luar jendela tak pernah mengirim `drop` maupun
   * `dragleave` yang seimbang, jadi penghitungnya bisa tertinggal di atas nol
   * dan overlay tersangkut. `dragend` dan kepergian kursor dari dokumen
   * mengembalikannya ke keadaan bersih.
   */
  function reset() {
    depth = 0
    isDraggingFiles.value = false
  }

  onMounted(() => {
    window.addEventListener('dragenter', onDragEnter)
    window.addEventListener('dragover', onDragOver)
    window.addEventListener('dragleave', onDragLeave)
    window.addEventListener('drop', onDropFiles)
    window.addEventListener('dragend', reset)
    window.addEventListener('blur', reset)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('dragenter', onDragEnter)
    window.removeEventListener('dragover', onDragOver)
    window.removeEventListener('dragleave', onDragLeave)
    window.removeEventListener('drop', onDropFiles)
    window.removeEventListener('dragend', reset)
    window.removeEventListener('blur', reset)
  })

  return { isDraggingFiles }
}
