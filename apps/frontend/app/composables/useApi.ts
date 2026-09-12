import type { ApiErrorBody, ApiErrorCode } from '~/types'

/**
 * ApiError membawa error code dari backend (docs/06 §Error Codes) supaya UI bisa
 * bereaksi spesifik — mis. NO_ROOM menawarkan pilih akun manual, PATH_EXISTS
 * menandai field nama, ACCOUNT_NEEDS_RECONNECT memunculkan tombol reconnect.
 */
export class ApiError extends Error {
  code: ApiErrorCode
  status: number

  constructor(code: ApiErrorCode, message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

/** Pesan yang layak dibaca user untuk tiap error code. */
const FRIENDLY: Record<ApiErrorCode, string> = {
  NO_ROOM: 'Tidak ada akun yang punya ruang cukup untuk file ini.',
  ACCOUNT_NEEDS_RECONNECT: 'Akun perlu dihubungkan ulang — izin aksesnya sudah tidak berlaku.',
  PROVIDER_ERROR: 'Provider menolak permintaan ini.',
  NOT_FOUND: 'Data yang diminta tidak ditemukan.',
  RATE_LIMITED: 'Provider sedang membatasi permintaan. Coba lagi sebentar lagi.',
  PATH_EXISTS: 'Sudah ada item dengan nama itu di lokasi ini.',
  FOLDER_NOT_EMPTY: 'Folder masih berisi. Hapus isinya dulu atau pakai hapus rekursif.',
  INVALID_ARGUMENT: 'Permintaan tidak valid.',
  UNSUPPORTED: 'Operasi ini belum didukung.',
  INTERNAL: 'Terjadi kesalahan di server.'
}

export function friendlyMessage(err: unknown): string {
  if (err instanceof ApiError) {
    return FRIENDLY[err.code] || err.message
  }
  if (err instanceof Error && err.message) {
    return err.message
  }
  return 'Terjadi kesalahan yang tidak terduga.'
}

/**
 * useApi membungkus $fetch dengan base URL backend dan menerjemahkan body error
 * { error: { code, message } } jadi ApiError.
 */
export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase

  async function request<T>(path: string, opts: Parameters<typeof $fetch>[1] = {}): Promise<T> {
    try {
      return await $fetch<T>(path, { baseURL, ...opts })
    } catch (err: any) {
      const body = err?.data as ApiErrorBody | undefined
      const status = err?.status ?? err?.statusCode ?? 0
      if (body?.error?.code) {
        throw new ApiError(body.error.code, body.error.message, status)
      }
      // Backend tak terjangkau / respons bukan JSON kontrak kita.
      throw new ApiError('INTERNAL', err?.message || 'Tidak bisa menghubungi server', status)
    }
  }

  return {
    baseURL,
    request,
    get: <T>(path: string, query?: Record<string, any>) =>
      request<T>(path, { method: 'GET', query }),
    post: <T>(path: string, body?: any, query?: Record<string, any>) =>
      request<T>(path, { method: 'POST', body, query }),
    patch: <T>(path: string, body?: any) =>
      request<T>(path, { method: 'PATCH', body }),
    del: <T>(path: string, query?: Record<string, any>) =>
      request<T>(path, { method: 'DELETE', query })
  }
}
