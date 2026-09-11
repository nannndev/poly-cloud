<script setup lang="ts">
import type { StorageProvider } from '~/types'

const route = useRoute()
const router = useRouter()
const accountsStore = useAccountsStore()

const currentStep = ref<1 | 2 | 3 | 4>(1)
const provider = computed(() => (route.query.provider as StorageProvider) || 'gdrive')
const label = computed(() => (route.query.label as string) || 'Google Drive Primary')
const email = computed(() => (route.query.email as string) || 'ekaprasetya2244@gmail.com')
const capacityGb = computed(() => Number(route.query.capacity) || 15)

const steps = [
  { id: 1, title: 'Validating OAuth State & Exchanging Code', desc: 'Communicating with provider OAuth token endpoint' },
  { id: 2, title: 'Encrypting Token Credentials (AES-256-GCM)', desc: 'Securing access and refresh tokens in PostgreSQL' },
  { id: 3, title: 'Provisioning rclone Daemon Remote', desc: 'Calling rclone rcd config/create with acc_<uuid> identity' },
  { id: 4, title: 'Initial Metadata Synchronization Complete', desc: 'Populating files_index and starting smart routing' }
]

onMounted(async () => {
  // Step 1: Exchange code
  await new Promise(resolve => setTimeout(resolve, 700))
  currentStep.value = 2

  // Step 2: Encrypt token
  await new Promise(resolve => setTimeout(resolve, 600))
  currentStep.value = 3

  // Step 3: Provision rclone daemon
  await new Promise(resolve => setTimeout(resolve, 800))
  currentStep.value = 4

  // Add the newly provisioned account into the Pinia store
  accountsStore.addAccount({
    provider: provider.value,
    label: label.value,
    email: email.value,
    total_bytes: capacityGb.value * 1024 * 1024 * 1024,
    provisioning_type: 'oauth'
  })

  // Step 4: Redirect to accounts page
  setTimeout(() => {
    router.push('/accounts')
  }, 1000)
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-[#09090b] text-zinc-200">
    <div class="max-w-lg w-full p-8 rounded-3xl border border-white/[0.08] bg-[#0c0c0e] shadow-2xl space-y-6">
      <!-- Top Status Icon -->
      <div class="flex flex-col items-center text-center space-y-3">
        <div class="flex size-14 items-center justify-center rounded-2xl bg-emerald-500/15 text-emerald-400 border border-emerald-500/25">
          <UIcon
            :name="currentStep === 4 ? 'i-lucide-check' : 'i-lucide-loader-2'"
            class="size-7"
            :class="[currentStep !== 4 ? 'animate-spin' : 'stroke-[2.5]']"
          />
        </div>

        <div>
          <h2 class="text-base font-bold text-white">
            {{ currentStep === 4 ? 'Account Provisioned Successfully!' : 'Provisioning Cloud Account...' }}
          </h2>
          <p class="text-xs text-zinc-400 mt-1 max-w-sm">
            Backend-managed OAuth is registering <strong class="text-emerald-400">{{ label }}</strong> into the Poly Cloud platform and configuring the rclone daemon.
          </p>
        </div>
      </div>

      <!-- Provisioning Stepper Tracker -->
      <div class="space-y-2.5 p-4 rounded-2xl bg-[#121215] border border-white/[0.07]">
        <div
          v-for="step in steps"
          :key="step.id"
          class="flex items-start gap-3 text-xs transition-opacity duration-200"
          :class="[
            step.id < currentStep
              ? 'opacity-60 text-zinc-400'
              : step.id === currentStep
              ? 'opacity-100 text-white'
              : 'opacity-30 text-zinc-600'
          ]"
        >
          <div class="mt-0.5 shrink-0">
            <UIcon
              v-if="step.id < currentStep"
              name="i-lucide-check-circle-2"
              class="size-4 text-emerald-400"
            />
            <UIcon
              v-else-if="step.id === currentStep"
              name="i-lucide-loader-2"
              class="size-4 text-emerald-400 animate-spin"
            />
            <div
              v-else
              class="size-4 rounded-full border border-white/[0.15] flex items-center justify-center text-[9px] font-mono text-zinc-500"
            >
              {{ step.id }}
            </div>
          </div>

          <div class="min-w-0">
            <p class="font-medium text-[11px]" :class="step.id === currentStep ? 'text-emerald-300 font-semibold' : ''">
              {{ step.title }}
            </p>
            <p class="text-[10px] text-zinc-500 truncate mt-0.5">{{ step.desc }}</p>
          </div>
        </div>
      </div>

      <!-- Progress Bar -->
      <div class="space-y-2">
        <div class="h-1.5 w-full bg-zinc-900 border border-white/[0.06] rounded-full overflow-hidden">
          <div
            class="h-full bg-emerald-500 rounded-full transition-all duration-500"
            :style="{ width: `${(currentStep / 4) * 100}%` }"
          />
        </div>
        <div class="flex items-center justify-between text-[10px] text-zinc-500 font-mono">
          <span>Provisioning Step {{ currentStep }} of 4</span>
          <span class="text-emerald-400">{{ (currentStep / 4) * 100 }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>
