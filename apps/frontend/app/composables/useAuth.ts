export interface AuthUser {
  id: string
  email: string
}

export interface LoginResponse {
  token: string
  user: AuthUser
}

export function useAuth() {
  const api = useApi()
  const token = useCookie<string | null>('polycloud_session', {
    maxAge: 60 * 60 * 24 * 7, // 7 hari
    path: '/',
    sameSite: 'lax'
  })

  const user = useCookie<AuthUser | null>('polycloud_user', {
    maxAge: 60 * 60 * 24 * 7,
    path: '/',
    sameSite: 'lax'
  })

  const isAuthenticated = computed(() => !!token.value)

  async function login(email: string, pass: string) {
    const res = await api.post<LoginResponse>('/auth/login', {
      email,
      password: pass
    })
    token.value = res.token
    user.value = res.user
    return res
  }

  async function logout() {
    try {
      await api.post('/auth/logout')
    } catch {
      // abaikan error network saat logout
    }
    token.value = null
    user.value = null
    await navigateTo('/login')
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout
  }
}
