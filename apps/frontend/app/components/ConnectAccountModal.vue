<script setup lang="ts">
import type { StorageProvider } from '~/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const accountsStore = useAccountsStore()
const toast = useToast()

const isOpen = computed({
  get: () => props.open,
  set: (val) => emit('update:open', val)
})

const selectedProvider = ref<StorageProvider>('gdrive')
const label = ref('')
const isSubmitting = ref(false)
const submitError = ref('')

// Kredensial provider berbasis key (S3 / R2 / B2).
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
    defaultLabel: 'Google Drive'
  },
  {
    id: 'onedrive',
    name: 'OneDrive',
    desc: 'Microsoft Graph',
    icon: 'i-simple-icons-microsoftonedrive',
    tag: 'OAuth 2.0',
    isOAuth: true,
    defaultLabel: 'OneDrive'
  },
  {
    id: 'dropbox',
    name: 'Dropbox',
    desc: 'Personal & Business',
    icon: 'i-simple-icons-dropbox',
    tag: 'OAuth 2.0',
    isOAuth: true,
    defaultLabel: 'Dropbox'
  },
  {
    id: 's3',
    name: 'AWS S3',
    desc: 'S3 & compatible',
    icon: 'i-simple-icons-amazonwebservices',
    tag: 'Access Key',
    isOAuth: false,
    defaultLabel: 'S3 Bucket'
  },
  {
    id: 'r2',
    name: 'Cloudflare R2',
    desc: 'Zero egress',
    icon: 'i-simple-icons-cloudflare',
    tag: 'Access Key',
    isOAuth: false,
    defaultLabel: 'Cloudflare R2'
  },
  {
    id: 'b2',
    name: 'Backblaze B2',
    desc: 'Low-cost archive',
    icon: 'i-simple-icons-backblaze',
    tag: 'App Key',
    isOAuth: false,
    defaultLabel: 'Backblaze B2'
  }
]

const currentProviderMeta = computed(() =>
  providers.find(p => p.id === selectedProvider.value) || providers[0]!)

// Backend melaporkan provider OAuth mana yang kredensial platformnya sudah diisi
// di env; tanpa itu, tombol authorize pasti gagal (docs/10 §1).
const isProviderReady = computed(() => {
  const info = accountsStore.providerInfo(selectedProvider.value)
  return info ? info.configured : true
})

onMounted(() => {
  if (accountsStore.providers.length === 0) {
    accountsStore.fetchProviders().catch(() => null)
  }
})

watch(isOpen, (open) => {
  if (open) {
    submitError.value = ''
    if (!label.value) label.value = currentProviderMeta.value.defaultLabel
  }
})

function selectProvider(p: StorageProvider) {
  selectedProvider.value = p
  submitError.value = ''
  const found = providers.find(item => item.id === p)
  if (found) {
    label.value = found.defaultLabel
    if (p === 's3') {
      endpointOrRegion.value = 'us-east-1'
    } else if (p === 'r2') {
      endpointOrRegion.value = 'https://<account_id>.r2.cloudflarestorage.com'
    } else if (p === 'b2') {
      endpointOrRegion.value = ''
    }
  }
}

async function handleConnect() {
  submitError.value = ''
  const finalLabel = label.value.trim() || currentProviderMeta.value.defaultLabel
  isSubmitting.value = true

  try {
    if (currentProviderMeta.value.isOAuth) {
      // Backend membangun consent URL memakai OAuth app platform, lalu browser
      // diarahkan ke provider. Sisanya diselesaikan halaman callback.
      const authUrl = await accountsStore.startOAuthConnect(selectedProvider.value, finalLabel)
      window.location.href = authUrl
      return
    }

    // Provider berbasis key: akun langsung jadi tanpa redirect.
    const fields: Record<string, string> = { bucket: bucketName.value.trim() }
    if (selectedProvider.value === 'b2') {
      fields.account = accessKeyId.value.trim()
      fields.key = secretAccessKey.value.trim()
    } else {
      fields.access_key_id = accessKeyId.value.trim()
      fields.secret_access_key = secretAccessKey.value.trim()
      if (selectedProvider.value === 'r2') {
        fields.endpoint = endpointOrRegion.value.trim()
      } else if (endpointOrRegion.value.trim()) {
        // Nilai berupa URL diperlakukan sebagai endpoint kustom (MinIO dsb.),
        // selain itu sebagai region AWS.
        const value = endpointOrRegion.value.trim()
        if (value.startsWith('http://') || value.startsWith('https://')) {
          fields.endpoint = value
        } else {
          fields.region = value
        }
      }
    }

    const account = await accountsStore.connectWithKeys(selectedProvider.value, finalLabel, fields)
    toast.add({
      title: 'Account connected',
      description: `${account.label} is ready. Run a sync to index its contents.`,
      color: 'success'
    })
    resetForm()
    isOpen.value = false
    await accountsStore.fetchQuota().catch(() => null)
  } catch (err) {
    submitError.value = friendlyMessage(err)
  } finally {
    isSubmitting.value = false
  }
}

function resetForm() {
  accessKeyId.value = ''
  secretAccessKey.value = ''
  bucketName.value = ''
  label.value = ''
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
                color="primary"
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
                placeholder="e.g. Work GDrive"
                icon="i-lucide-tag"
                size="md"
                class="w-full rounded-xl"
              />
              <span class="text-[10px] text-zinc-500 block">Shown in the explorer and on file badges.</span>
            </div>

            <!-- Key-based providers: bucket is required, since the root of an object
                 storage remote is a list of buckets, not a place to put files. -->
            <div v-if="!currentProviderMeta.isOAuth" class="space-y-1.5">
              <label class="block text-xs font-semibold text-zinc-300">
                Bucket Name
              </label>
              <UInput
                v-model="bucketName"
                placeholder="e.g. polycloud-vault"
                icon="i-lucide-folder-archive"
                size="md"
                class="w-full rounded-xl"
              />
              <span class="text-[10px] text-zinc-500 block">Destination bucket; every object is stored inside it.</span>
            </div>
          </div>

          <!-- Key-based provider credentials (S3, R2, B2) -->
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

          <!-- Warning: OAuth provider whose platform credentials are not set -->
          <div
            v-if="currentProviderMeta.isOAuth && !isProviderReady"
            class="flex items-start gap-3 p-3.5 rounded-2xl border border-amber-500/25 bg-amber-500/[0.06] text-xs"
          >
            <UIcon name="i-lucide-triangle-alert" class="size-4 text-amber-400 shrink-0 mt-0.5" />
            <div class="leading-relaxed text-zinc-300">
              <span class="font-bold text-amber-300">Provider not configured.</span>
              The {{ currentProviderMeta.name }} OAuth credentials are missing from the backend
              environment, so authorization will be rejected. Register an OAuth app with the
              provider, then set its client id &amp; secret in <code class="font-mono text-zinc-400">.env</code>.
            </div>
          </div>

          <!-- Error from the backend -->
          <div
            v-if="submitError"
            class="flex items-start gap-3 p-3.5 rounded-2xl border border-red-500/25 bg-red-500/[0.06] text-xs"
          >
            <UIcon name="i-lucide-circle-alert" class="size-4 text-red-400 shrink-0 mt-0.5" />
            <span class="leading-relaxed text-zinc-300">{{ submitError }}</span>
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
            :disabled="isSubmitting || (currentProviderMeta.isOAuth && !isProviderReady)"
            @click="handleConnect"
          >
            <UIcon
              :name="isSubmitting ? 'i-lucide-loader-2' : currentProviderMeta.isOAuth ? 'i-lucide-external-link' : 'i-lucide-plus-circle'"
              class="size-4"
              :class="isSubmitting ? 'animate-spin' : ''"
            />
            <span>
              {{ isSubmitting
                ? 'Connecting...'
                : currentProviderMeta.isOAuth
                ? `Authorize with ${currentProviderMeta.name}`
                : 'Provision Remote' }}
            </span>
          </button>
        </div>
      </div>
    </template>
  </UModal>
</template>
