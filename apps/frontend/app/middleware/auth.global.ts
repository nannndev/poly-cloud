export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie<string | null>('polycloud_session')

  // Public route whitelist
  if (to.path === '/login') {
    if (token.value) {
      return navigateTo('/')
    }
    return
  }

  // Protect all dashboard routes
  if (!token.value) {
    return navigateTo('/login')
  }
})
