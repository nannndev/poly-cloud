<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const accountsStore = useAccountsStore()

type Phase = 'exchanging' | 'syncing' | 'done' | 'failed'

const phase = ref<Phase>('exchanging')
const errorMessage = ref('')
const accountLabel = ref('')

// Provider mengirim balik ?code&state; error consent datang sebagai ?error.
const code = computed(() => (route.query.code as string) || '')
const state = computed(() => (route.query.state as string) || '')
const providerError = computed(() => (route.query.error_description || route.query.error) as string | undefined)

const steps = [
  { id: 1, title: 'Exchanging authorization code', desc: 'Backend calls the provider token endpoint' },
  { id: 2, title: 'Storing encrypted token', desc: 'AES-256-GCM in the account_tokens table' },
  { id: 3, title: 'Provisioning rclone remote', desc: 'config/create under the acc_<uuid> identity' },
  { id: 4, title: 'Initial index sync', desc: 'Reading account contents into files_index' }
]

// Backend menyelesaikan langkah 1-3 dalam satu panggilan callback, lalu
// menjalankan sync di latar belakang. Stepper mencerminkan itu.
const currentStep = computed(() => {
  if (phase.value === 'exchanging') return 1
  if (phase.value === 'syncing') return 4
  return 4
})

const progressPercent = computed(() => {
  if (phase.value === 'exchanging') return 35
  if (phase.value === 'syncing') return 80
  return 100
})

onMounted(async () => {
  if (providerError.value) {
    phase.value = 'failed'
    errorMessage.value = `The provider rejected authorization: ${providerError.value}`
    return
  }
  if (!code.value || !state.value) {
    phase.value = 'failed'
    errorMessage.value = 'The callback URL is missing the code or state parameter.'
    return
  }

  try {
    const account = await accountsStore.completeOAuthConnect(code.value, state.value)
    accountLabel.value = account.label
    phase.value = 'syncing'

    // Backend memulai initial sync sendiri; muat ulang agar kuota & jumlah file
    // yang sudah terindeks muncul.
    await new Promise(resolve => setTimeout(resolve, 1200))
    await accountsStore.loadAll().catch(() => null)
    phase.value = 'done'

    setTimeout(() => router.push('/accounts'), 900)
  } catch (err) {
    phase.value = 'failed'
    errorMessage.value = friendlyMessage(err)
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-[#0b0e14] text-zinc-200">
    <div class="max-w-lg w-full p-8 rounded-3xl border border-white/[0.08] bg-[#0d111a] shadow-2xl space-y-6">
      <!-- Status utama -->
      <div class="flex flex-col items-center text-center space-y-3">
        <div
          class="flex size-14 items-center justify-center rounded-2xl border"
          :class="phase === 'failed'
            ? 'bg-red-500/15 text-red-400 border-red-500/25'
            : 'bg-primary-500/15 text-primary-400 border-primary-500/25'"
        >
          <UIcon
            :name="phase === 'failed' ? 'i-lucide-x' : phase === 'done' ? 'i-lucide-check' : 'i-lucide-loader-2'"
            class="size-7"
            :class="[phase === 'done' || phase === 'failed' ? 'stroke-[2.5]' : 'animate-spin']"
          />
        </div>

        <div>
          <h2 class="text-base font-bold text-white">
            {{ phase === 'failed'
              ? 'Could not connect the account'
              : phase === 'done'
              ? 'Account connected'
              : 'Connecting account...' }}
          </h2>
          <p v-if="phase !== 'failed'" class="text-xs text-zinc-400 mt-1 max-w-sm">
            The backend exchanges the authorization code, stores the encrypted token, then
            registers <strong class="text-primary-400">{{ accountLabel || 'this account' }}</strong>
            as a remote in the rclone daemon.
          </p>
          <p v-else class="text-xs text-zinc-400 mt-1 max-w-sm">{{ errorMessage }}</p>
        </div>
      </div>

      <!-- Stepper -->
      <div v-if="phase !== 'failed'" class="space-y-2.5 p-4 rounded-2xl bg-[#151a27] border border-white/[0.07]">
        <div
          v-for="step in steps"
          :key="step.id"
          class="flex items-start gap-3 text-xs transition-opacity duration-200"
          :class="[
            phase === 'done' || step.id < currentStep
              ? 'opacity-60 text-zinc-400'
              : step.id === currentStep
              ? 'opacity-100 text-white'
              : 'opacity-30 text-zinc-600'
          ]"
        >
          <div class="mt-0.5 shrink-0">
            <UIcon
              v-if="phase === 'done' || step.id < currentStep"
              name="i-lucide-check-circle-2"
              class="size-4 text-primary-400"
            />
            <UIcon
              v-else-if="step.id === currentStep"
              name="i-lucide-loader-2"
              class="size-4 text-primary-400 animate-spin"
            />
            <div
              v-else
              class="size-4 rounded-full border border-white/[0.15] flex items-center justify-center text-[9px] font-mono text-zinc-500"
            >
              {{ step.id }}
            </div>
          </div>

          <div class="min-w-0">
            <p class="font-medium text-[11px]" :class="step.id === currentStep && phase !== 'done' ? 'text-primary-300 font-semibold' : ''">
              {{ step.title }}
            </p>
            <p class="text-[10px] text-zinc-500 truncate mt-0.5">{{ step.desc }}</p>
          </div>
        </div>
      </div>

      <!-- Progress -->
      <div v-if="phase !== 'failed'" class="space-y-2">
        <div class="h-1.5 w-full bg-zinc-900 border border-white/[0.06] rounded-full overflow-hidden">
          <div
            class="h-full bg-primary-500 rounded-full transition-all duration-500"
            :style="{ width: `${progressPercent}%` }"
          />
        </div>
        <div class="flex items-center justify-between text-[10px] text-zinc-500 font-mono">
          <span>{{ phase === 'done' ? 'Done' : 'In progress' }}</span>
          <span class="text-primary-400">{{ progressPercent }}%</span>
        </div>
      </div>

      <!-- Failure actions -->
      <div v-else class="flex items-center justify-center gap-2.5">
        <UButton
          label="Back to Accounts"
          color="neutral"
          variant="ghost"
          class="rounded-xl font-medium"
          @click="router.push('/accounts')"
        />
        <UButton
          label="Try again"
          color="primary"
          class="rounded-xl font-bold"
          @click="router.push('/accounts?connect=1')"
        />
      </div>
    </div>
  </div>
</template>
