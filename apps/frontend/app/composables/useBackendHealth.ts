/**
 * Status /healthz milik backend, dibagi seluruh aplikasi lewat useState supaya
 * sidebar dan halaman settings melaporkan hal yang sama, bukan dua penilaian
 * terpisah. Backend menjawab 503 saat DB tak terjangkau — respons itu tetap
 * membawa rincian per dependensi, jadi diperlakukan sebagai "degraded".
 */
export type HealthState = 'checking' | 'online' | 'degraded' | 'offline'

export function useBackendHealth() {
  const config = useRuntimeConfig()
  const healthUrl = computed(() =>
    config.public.apiBase.replace(/\/api\/v1$/, '') + '/healthz')

  const status = useState<HealthState>('backend-health-status', () => 'checking')
  const latency = useState<number | null>('backend-health-latency', () => null)
  const detail = useState<Record<string, string>>('backend-health-detail', () => ({}))
  const isChecking = useState('backend-health-checking', () => false)

  async function check() {
    isChecking.value = true
    status.value = 'checking'
    const start = performance.now()
    try {
      const res = await $fetch<Record<string, string>>(healthUrl.value, { timeout: 5000 })
      latency.value = Math.round(performance.now() - start)
      detail.value = res
      status.value = res.status === 'ok' ? 'online' : 'degraded'
    } catch (err: any) {
      const body = err?.data as Record<string, string> | undefined
      latency.value = Math.round(performance.now() - start)
      if (body?.status) {
        detail.value = body
        status.value = 'degraded'
      } else {
        detail.value = {}
        status.value = 'offline'
      }
    } finally {
      isChecking.value = false
    }
  }

  const label = computed(() => {
    switch (status.value) {
      case 'online': return 'Engine online'
      case 'degraded': return 'Engine degraded'
      case 'offline': return 'Engine offline'
      default: return 'Checking engine'
    }
  })

  const dotClass = computed(() => {
    switch (status.value) {
      case 'online': return 'bg-emerald-400'
      case 'degraded': return 'bg-amber-400'
      case 'offline': return 'bg-rose-500'
      default: return 'bg-zinc-500 animate-pulse'
    }
  })

  return { status, latency, detail, isChecking, healthUrl, label, dotClass, check }
}
