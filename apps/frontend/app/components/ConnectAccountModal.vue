<script setup lang="ts">
import type { StorageProvider } from '~/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const router = useRouter()
const accountsStore = useAccountsStore()

const isOpen = computed({
  get: () => props.open,
  set: (val) => emit('update:open', val)
})

const selectedProvider = ref<StorageProvider>('gdrive')
const label = ref('Google Drive Primary')
const email = ref('ekaprasetya2244@gmail.com')
const capacityGb = ref(15)
const isSubmitting = ref(false)

// Non-OAuth Credential Fields (S3, R2, B2)
const accessKeyId = ref('')
const secretAccessKey = ref('')
const bucketName = ref('')
const endpointOrRegion = ref('us-east-1')
const showSecret = ref(false)

interface ProviderDefinition {
  id: StorageProvider
  name: string
  desc: string
  icon: string
  tag: string
  isOAuth: boolean
  defaultCapacity: number
  defaultLabel: string
}

const providers: ProviderDefinition[] = [
  {
    id: 'gdrive',
    name: 'Google Drive',
    desc: 'Personal & Workspace',
    icon: 'i-simple-icons-googledrive',
    tag: 'OAuth 2.0',
    isOAuth: true,
    defaultCapacity: 15,
    defaultLabel: 'Google Drive Primary'
  },
  {
    id: 'onedrive',
    name: 'OneDrive',
    desc: 'Microsoft Graph',
    icon: 'i-simple-icons-microsoftonedrive',
    tag: 'OAuth 2.0',
    isOAuth: true,
    defaultCapacity: 100,
    defaultLabel: 'OneDrive Business'
  },
  {
    id: 'dropbox',
    name: 'Dropbox',
    desc: 'Personal & Team Share',
    icon: 'i-simple-icons-dropbox',
    tag: 'OAuth 2.0',
    isOAuth: true,
    defaultCapacity: 20,
    defaultLabel: 'Dropbox Team Share'
  },
  {
    id: 's3',
    name: 'Amazon S3',
    desc: 'Standard & Glacier Vault',
    icon: 'i-simple-icons-amazons3',
    tag: 'IAM Keys',
    isOAuth: false,
    defaultCapacity: 250,
    defaultLabel: 'AWS S3 Cold Bucket'
  },
  {
    id: 'r2',
    name: 'Cloudflare R2',
    desc: 'Zero-Egress Object Store',
    icon: 'i-simple-icons-cloudflare',
    tag: 'S3-Compatible',
    isOAuth: false,
    defaultCapacity: 50,
    defaultLabel: 'Cloudflare R2 Media Hot'
  },
  {
    id: 'b2',
    name: 'Backblaze B2',
    desc: 'High-Durability Cold Store',
    icon: 'i-lucide-hard-drive',
    tag: 'App Key ID',
    isOAuth: false,
    defaultCapacity: 100,
    defaultLabel: 'Backblaze B2 Archive'
  }
]

const currentProviderMeta = computed(() => {
  return providers.find(p => p.id === selectedProvider.value) || providers[0]
})

const presetCapacities = [
  { label: '15 GB', value: 15, tag: 'Free Tier' },
  { label: '50 GB', value: 50 },
  { label: '100 GB', value: 100, tag: 'Standard' },
  { label: '500 GB', value: 500 },
  { label: '1 TB', value: 1024 },
  { label: '5 TB', value: 5120 }
]

function selectProvider(p: StorageProvider) {
  selectedProvider.value = p
  const found = providers.find(item => item.id === p)
  if (found) {
    capacityGb.value = found.defaultCapacity
    label.value = found.defaultLabel
    if (p === 's3') {
      endpointOrRegion.value = 'us-east-1'
    } else if (p === 'r2') {
      endpointOrRegion.value = 'https://<account_id>.r2.cloudflarestorage.com'
    }
  }
}

async function handleConnect() {
  if (!label.value.trim()) {
    label.value = currentProviderMeta.value.defaultLabel
  }

  isSubmitting.value = true

  // Flow per doc 10 (Account Provisioning):
  if (currentProviderMeta.value.isOAuth) {
    // OAuth flow: Redirect to callback simulating backend-managed OAuth exchange
    await new Promise(resolve => setTimeout(resolve, 400))
    isSubmitting.value = false
    isOpen.value = false

    // Route to OAuth callback page with parameters
    router.push({
      path: '/connect/callback',
      query: {
        provider: selectedProvider.value,
        label: label.value,
        email: email.value || 'ekaprasetya2244@gmail.com',
        capacity: capacityGb.value.toString()
      }
    })
  } else {
    // Non-OAuth direct key provisioning:
    await new Promise(resolve => setTimeout(resolve, 800))

    const bucketIdentifier = bucketName.value.trim() || `${selectedProvider.value}-vault`
    accountsStore.addAccount({
      provider: selectedProvider.value,
      label: label.value,
      email: bucketIdentifier,
      total_bytes: capacityGb.value * 1024 * 1024 * 1024,
      provisioning_type: 'credentials'
    })

    isSubmitting.value = false
    isOpen.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="isOpen"
    :ui="{
      content: 'sm:max-w-2xl bg-[#0c0c0e] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
      body: 'p-6 space-y-5',
      footer: 'px-6 py-4 bg-[#09090b] border-t border-white/[0.06]'
    }"
  >
    <!-- Custom Modal Header -->
    <template #header>
      <div class="px-6 pt-5 pb-4 bg-[#111114] flex items-start justify-between gap-4 border-b border-white/[0.06]">
        <div class="flex items-start gap-3.5">
          <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-emerald-500/15 text-emerald-400 border border-emerald-500/25">
            <UIcon name="i-lucide-cloud-plus" class="size-6" />
          </div>
          <div>
            <h2 class="text-base font-bold tracking-tight text-white flex items-center gap-2">
              Connect Cloud Storage Account
              <UBadge
                :label="currentProviderMeta.isOAuth ? 'Backend OAuth' : 'Direct Key Provisioning'"
                color="emerald"
                variant="subtle"
                size="xs"
                class="rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[9px] font-medium"
              />
            </h2>
            <p class="text-xs text-zinc-400 mt-0.5 leading-relaxed max-w-lg">
              Dynamic provisioning injects credentials as an isolated <code class="text-emerald-400 font-mono text-[11px]">acc_&lt;uuid&gt;:</code> remote in the internal rclone daemon.
            </p>
          </div>
        </div>

        <button
          type="button"
          class="flex size-8 items-center justify-center rounded-xl bg-zinc-800/60 hover:bg-zinc-700 text-zinc-400 hover:text-white border border-white/[0.06] transition-colors cursor-pointer"
          @click="isOpen = false"
        >
          <UIcon name="i-lucide-x" class="size-4" />
        </button>
      </div>
    </template>

    <template #body>
      <div class="space-y-5">
        <!-- 1. Provider Selection Grid -->
        <div>
          <div class="flex items-center justify-between mb-2.5">
            <label class="text-[11px] font-bold uppercase tracking-wider text-zinc-400 flex items-center gap-1.5">
              <span>1. Select Storage Provider</span>
              <span class="text-emerald-400 font-normal">({{ providers.length }} supported)</span>
            </label>
            <span class="text-[11px] text-zinc-500 font-medium">Selected: <strong class="text-zinc-200">{{ currentProviderMeta.name }}</strong></span>
          </div>

          <div class="grid grid-cols-2 sm:grid-cols-3 gap-2.5">
            <button
              v-for="p in providers"
              :key="p.id"
              type="button"
              class="group relative flex flex-col justify-between p-3.5 rounded-2xl border text-left transition-all duration-200 cursor-pointer shadow-xs"
              :class="[
                selectedProvider === p.id
                  ? 'border-emerald-500/60 bg-emerald-500/10 ring-1 ring-emerald-500/30'
                  : 'border-white/[0.07] bg-[#121215] hover:border-white/[0.15] hover:bg-[#15151a]'
              ]"
              @click="selectProvider(p.id)"
            >
              <div class="flex items-start justify-between w-full mb-3">
                <div class="flex size-9 items-center justify-center rounded-xl bg-[#16161b] border border-white/[0.06] group-hover:scale-105 transition-transform">
                  <UIcon :name="p.icon" class="size-5 text-emerald-400" />
                </div>

                <div class="flex items-center gap-1">
                  <span
                    class="text-[9px] font-semibold px-1.5 py-0.5 rounded-md border"
                    :class="[
                      p.isOAuth
                        ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                        : 'bg-zinc-800/80 text-zinc-400 border-white/[0.06]'
                    ]"
                  >
                    {{ p.tag }}
                  </span>
                  <UIcon
                    v-if="selectedProvider === p.id"
                    name="i-lucide-check"
                    class="size-3.5 text-emerald-400 stroke-[3]"
                  />
                </div>
              </div>

              <div>
                <h4 class="text-xs font-bold text-zinc-200 leading-snug">{{ p.name }}</h4>
                <p class="text-[10px] text-zinc-500 line-clamp-1 mt-0.5">{{ p.desc }}</p>
              </div>
            </button>
          </div>
        </div>

        <!-- 2. Security & Provisioning Architecture Note (ADR-010 & Doc 10) -->
        <div class="flex items-start gap-3 p-3.5 rounded-2xl border border-emerald-500/20 bg-emerald-500/[0.05] text-xs">
          <UIcon name="i-lucide-shield-check" class="size-4 text-emerald-400 shrink-0 mt-0.5" />
          <div class="leading-relaxed">
            <span class="font-bold text-white">
              {{ currentProviderMeta.isOAuth ? 'Backend-Managed OAuth Architecture:' : 'Direct Key Provisioning Architecture:' }}
            </span>
            <span class="text-zinc-400 ml-1">
              {{ currentProviderMeta.isOAuth
                ? 'Platform OAuth credentials handle authentication. Tokens are stored encrypted (AES-256-GCM) in PostgreSQL and provisioned to the local rclone daemon via RC API.'
                : 'Access keys are stored encrypted and directly injected into the internal rclone daemon using config/create. No browser redirects.' }}
            </span>
          </div>
        </div>

        <!-- 3. Dynamic Configuration Fields (OAuth vs Non-OAuth) -->
        <div class="space-y-4">
          <!-- Common Field: Label -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <label class="block text-xs font-semibold text-zinc-300">
                Account Label / Alias
              </label>
              <UInput
                v-model="label"
                placeholder="e.g., GDrive Primary Work"
                icon="i-lucide-tag"
                size="md"
                class="w-full rounded-xl"
              />
              <span class="text-[10px] text-zinc-500 block">Display identifier shown in explorer and smart routing.</span>
            </div>

            <!-- If OAuth: Email field -->
            <div v-if="currentProviderMeta.isOAuth" class="space-y-1.5">
              <label class="block text-xs font-semibold text-zinc-300">
                Connected Account Email
              </label>
              <UInput
                v-model="email"
                placeholder="ekaprasetya2244@gmail.com"
                icon="i-lucide-mail"
                size="md"
                class="w-full rounded-xl"
              />
              <span class="text-[10px] text-zinc-500 block">Account owner address for identity tracking.</span>
            </div>

            <!-- If Non-OAuth S3/R2/B2: Bucket Name -->
            <div v-else class="space-y-1.5">
              <label class="block text-xs font-semibold text-zinc-300">
                Target Bucket Name
              </label>
              <UInput
                v-model="bucketName"
                placeholder="e.g., polycloud-vault-2026"
                icon="i-lucide-folder-archive"
                size="md"
                class="w-full rounded-xl"
              />
              <span class="text-[10px] text-zinc-500 block">Bucket or container name in cloud provider.</span>
            </div>
          </div>

          <!-- Non-OAuth Specific Fields (S3, R2, B2) -->
          <div v-if="!currentProviderMeta.isOAuth" class="p-4 rounded-2xl bg-[#121215] border border-white/[0.07] space-y-3.5">
            <div class="flex items-center justify-between text-xs font-semibold text-zinc-300">
              <span class="flex items-center gap-1.5">
                <UIcon name="i-lucide-key" class="size-3.5 text-amber-400" />
                {{ currentProviderMeta.id === 'b2' ? 'Backblaze B2 Application Credentials' : 'S3 / IAM API Credentials' }}
              </span>
              <span class="text-[10px] text-zinc-500 font-mono">Encrypted AES-256-GCM</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
              <!-- Key ID -->
              <div class="space-y-1">
                <label class="block text-[11px] font-medium text-zinc-400">
                  {{ currentProviderMeta.id === 'b2' ? 'Key ID' : 'Access Key ID' }}
                </label>
                <UInput
                  v-model="accessKeyId"
                  :placeholder="currentProviderMeta.id === 'b2' ? '002xxxxxxxxxxxxxxxx0001' : 'AKIAIOSFODNN7EXAMPLE'"
                  icon="i-lucide-fingerprint"
                  size="md"
                  class="w-full rounded-xl"
                />
              </div>

              <!-- Secret Key -->
              <div class="space-y-1">
                <div class="flex items-center justify-between">
                  <label class="block text-[11px] font-medium text-zinc-400">
                    {{ currentProviderMeta.id === 'b2' ? 'Application Key' : 'Secret Access Key' }}
                  </label>
                  <button
                    type="button"
                    class="text-[10px] text-emerald-400 hover:underline cursor-pointer"
                    @click="showSecret = !showSecret"
                  >
                    {{ showSecret ? 'Hide' : 'Show' }}
                  </button>
                </div>
                <UInput
                  v-model="secretAccessKey"
                  :type="showSecret ? 'text' : 'password'"
                  placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
                  icon="i-lucide-lock"
                  size="md"
                  class="w-full rounded-xl"
                />
              </div>
            </div>

            <!-- Region or Custom Endpoint -->
            <div class="space-y-1">
              <label class="block text-[11px] font-medium text-zinc-400">
                {{ currentProviderMeta.id === 'r2' ? 'Cloudflare Endpoint URL' : 'Region / Host' }}
              </label>
              <UInput
                v-model="endpointOrRegion"
                :placeholder="currentProviderMeta.id === 'r2' ? 'https://<account_id>.r2.cloudflarestorage.com' : 'us-east-1'"
                icon="i-lucide-globe"
                size="md"
                class="w-full rounded-xl"
              />
            </div>
          </div>

          <!-- 4. Storage Quota Allocation Slider & Presets -->
          <div class="space-y-2.5">
            <div class="flex items-center justify-between">
              <label class="text-xs font-semibold text-zinc-300 flex items-center gap-1.5">
                <span>Storage Quota Allocation</span>
              </label>
              <span class="font-mono text-xs font-bold text-emerald-400">
                {{ capacityGb >= 1024 ? `${(capacityGb / 1024).toFixed(1)} TB` : `${capacityGb} GB` }}
              </span>
            </div>

            <!-- Quick Preset Chips -->
            <div class="flex flex-wrap items-center gap-1.5">
              <button
                v-for="preset in presetCapacities"
                :key="preset.value"
                type="button"
                class="px-2.5 py-1 rounded-xl text-xs font-medium border transition-all duration-150 flex items-center gap-1.5 cursor-pointer shadow-xs"
                :class="[
                  capacityGb === preset.value
                    ? 'border-emerald-500/50 bg-emerald-500/15 text-emerald-300 font-bold'
                    : 'border-white/[0.08] bg-[#121215] text-zinc-400 hover:text-white hover:bg-[#16161b]'
                ]"
                @click="capacityGb = preset.value"
              >
                <span>{{ preset.label }}</span>
                <span v-if="preset.tag" class="text-[9px] opacity-75 uppercase">({{ preset.tag }})</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex items-center justify-between w-full">
        <span class="text-[11px] text-zinc-500 flex items-center gap-1.5">
          <UIcon name="i-lucide-cpu" class="size-3.5 text-emerald-400" />
          <span>Provisioning target: <code class="font-mono text-zinc-400">rclone rcd (localhost:5572)</code></span>
        </span>

        <div class="flex items-center gap-2.5">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            class="rounded-xl font-medium text-zinc-400 hover:text-white"
            @click="isOpen = false"
          />

          <button
            type="button"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold bg-emerald-600 hover:bg-emerald-500 text-white shadow-xs transition-all cursor-pointer disabled:opacity-50"
            :disabled="isSubmitting"
            @click="handleConnect"
          >
            <UIcon
              :name="isSubmitting ? 'i-lucide-loader-2' : currentProviderMeta.isOAuth ? 'i-lucide-external-link' : 'i-lucide-plus-circle'"
              class="size-4"
              :class="isSubmitting ? 'animate-spin' : ''"
            />
            <span>
              {{ isSubmitting
                ? 'Provisioning...'
                : currentProviderMeta.isOAuth
                ? `Authorize with ${currentProviderMeta.name}`
                : 'Validate & Provision Remote' }}
            </span>
          </button>
        </div>
      </div>
    </template>
  </UModal>
</template>
