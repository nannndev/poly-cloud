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
  NO_ROOM: 'No connected account has enough free space for this file.',
  ACCOUNT_NEEDS_RECONNECT: 'This account needs to be reconnected — its access has expired.',
  PROVIDER_ERROR: 'The provider rejected this request.',
  NOT_FOUND: 'The requested item was not found.',
  RATE_LIMITED: 'The provider is rate limiting requests. Try again in a moment.',
  PATH_EXISTS: 'An item with that name already exists here.',
  FOLDER_NOT_EMPTY: 'This folder is not empty. Empty it first, or delete recursively.',
  INVALID_ARGUMENT: 'That request is not valid.',
  UNSUPPORTED: 'This operation is not supported yet.',
  UNAUTHORIZED: 'Session expired or unauthorized. Please sign in again.',
  INTERNAL: 'Something went wrong on the server.'
}

export function friendlyMessage(err: unknown): string {
  if (err instanceof ApiError) {
    return FRIENDLY[err.code] || err.message
  }
  if (err instanceof Error && err.message) {
    return err.message
  }
  return 'An unexpected error occurred.'
}

/**
 * useApi membungkus $fetch dengan base URL backend dan menerjemahkan body error
 * { error: { code, message } } jadi ApiError.
 */
export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase
  const token = useCookie<string | null>('polycloud_session')

  async function request<T>(path: string, opts: Parameters<typeof $fetch>[1] = {}): Promise<T> {
    try {
      const headers: Record<string, string> = {
        ...((opts.headers as Record<string, string>) || {})
      }
      if (token.value) {
        headers['Authorization'] = `Bearer ${token.value}`
      }

      return await $fetch<T>(path, {
        baseURL,
        ...opts,
        headers
      })
    } catch (err: any) {
      const body = err?.data as ApiErrorBody | undefined
      const status = err?.status ?? err?.statusCode ?? 0

      if (status === 401 && !path.includes('/auth/login')) {
        token.value = null
        if (import.meta.client) {
          navigateTo('/login')
        }
      }

      if (body?.error?.code) {
        throw new ApiError(body.error.code, body.error.message, status)
      }
      // Backend tak terjangkau / respons bukan JSON kontrak kita.
      throw new ApiError('INTERNAL', err?.message || 'Could not reach the server', status)
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
