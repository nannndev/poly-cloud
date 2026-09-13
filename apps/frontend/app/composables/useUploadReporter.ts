/**
 * Unggah sekumpulan file lalu laporkan hasilnya sebagai toast.
 *
 * Dipakai bersama oleh modal upload dan area drop di halaman explorer — dua
 * pintu masuk yang sama-sama harus melaporkan kegagalan sebagian dengan cara
 * yang sama. Ditaruh di sini, bukan disalin ke keduanya, supaya pesan dan
 * perlakuan errornya tak pernah berbeda.
 */
export function useUploadReporter() {
  const filesStore = useFilesStore()
  const toast = useToast()

  /**
   * `targetFolderId` undefined berarti "pakai folder yang sedang dibuka" —
   * berbeda dari null, yang berarti root secara eksplisit.
   */
  async function uploadAndReport(fileList: File[], targetFolderId?: string | null) {
    if (fileList.length === 0) return []

    const folderId = targetFolderId === undefined
      ? filesStore.currentFolderId
      : targetFolderId

    // Backend yang memilih akun tujuan; UI tak mengirim preferensi akun.
    const results = await filesStore.uploadFiles(fileList, folderId)

    // Satu file gagal tak menghentikan sisanya, jadi keberhasilan sebagian
    // harus terlihat apa adanya — bukan dilaporkan sebagai sukses penuh.
    const failed = results.filter(r => r instanceof Error)
    if (failed.length > 0) {
      const ok = results.length - failed.length
      toast.add({
        title: `${failed.length} of ${results.length} file${results.length > 1 ? 's' : ''} failed to upload`,
        description: ok > 0
          ? `${ok} uploaded successfully. ${friendlyMessage(failed[0])}`
          : friendlyMessage(failed[0]),
        color: ok > 0 ? 'warning' : 'error'
      })
    } else {
      toast.add({
        title: fileList.length > 1 ? `${fileList.length} files uploaded` : 'File uploaded',
        description: 'The router picked the destination account with the most free space.',
        color: 'success'
      })
    }

    return results
  }

  return { uploadAndReport }
}
