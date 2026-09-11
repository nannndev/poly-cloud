<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const accountsStore = useAccountsStore()

const status = ref<'exchanging' | 'syncing' | 'completed'>('exchanging')
const provider = computed(() => (route.query.provider as string) || 'gdrive')

onMounted(async () => {
  // 1. Simulate code exchange
  await new Promise(resolve => setTimeout(resolve, 800))
  status.value = 'syncing'

  // 2. Simulate initial indexing
  await new Promise(resolve => setTimeout(resolve, 1200))
  status.value = 'completed'

  // 3. Redirect back to accounts
  setTimeout(() => {
    router.push('/accounts')
  }, 800)
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-background">
    <div class="max-w-md w-full p-8 rounded-3xl border border-default/70 bg-card shadow-lg text-center space-y-4">
      <div class="p-4 rounded-2xl bg-sky-500/10 text-sky-500 inline-flex mx-auto">
        <UIcon
          :name="status === 'completed' ? 'i-lucide-check-circle-2' : 'i-lucide-loader-2'"
          class="size-10"
          :class="[status !== 'completed' ? 'animate-spin' : 'text-emerald-500']"
        />
      </div>

      <div class="space-y-1">
        <h2 class="text-base font-bold text-highlighted">
          {{ status === 'exchanging' ? 'Verifying OAuth Token...' : status === 'syncing' ? 'Starting Metadata Index Synchronization...' : 'Connection Successful!' }}
        </h2>
        <p class="text-xs text-muted">
          Poly Cloud is connecting your {{ provider }} account and indexing file directories into the PostgreSQL database.
        </p>
      </div>

      <UProgress
        :model-value="status === 'exchanging' ? 35 : status === 'syncing' ? 75 : 100"
        :color="status === 'completed' ? 'success' : 'sky'"
        size="sm"
      />
    </div>
  </div>
</template>
