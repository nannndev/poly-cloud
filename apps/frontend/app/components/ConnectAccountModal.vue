<script setup lang="ts">
import type { StorageProvider } from '~/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const accountsStore = useAccountsStore()

const isOpen = computed({
  get: () => props.open,
  set: (val) => emit('update:open', val)
})

const selectedProvider = ref<StorageProvider>('gdrive')
const label = ref('Google Drive Primary')
const email = ref('')
const capacityGb = ref(15)
const isSubmitting = ref(false)

const providers = [
  {
    id: 'gdrive',
    name: 'Google Drive',
    desc: 'Personal & Google Workspace',
    icon: 'i-simple-icons-googledrive',
    tag: 'OAuth 2.0',
    color: 'emerald',
    borderColor: 'border-emerald-500/50 bg-emerald-500/[0.08] ring-emerald-500/30',
    iconColor: 'text-emerald-500',
    defaultCapacity: 15,
    defaultLabel: 'Google Drive Primary'
  },
  {
    id: 'onedrive',
    name: 'OneDrive',
    desc: 'Office 365 & SharePoint',
    icon: 'i-simple-icons-microsoftonedrive',
    tag: 'Microsoft Graph',
    color: 'sky',
    borderColor: 'border-sky-500/50 bg-sky-500/[0.08] ring-sky-500/30',
    iconColor: 'text-sky-500',
    defaultCapacity: 100,
    defaultLabel: 'OneDrive Business'
  },
  {
    id: 'dropbox',
    name: 'Dropbox',
    desc: 'Personal & Team Share',
    icon: 'i-simple-icons-dropbox',
    tag: 'Sync API',
    color: 'blue',
    borderColor: 'border-blue-500/50 bg-blue-500/[0.08] ring-blue-500/30',
    iconColor: 'text-blue-500',
    defaultCapacity: 20,
    defaultLabel: 'Dropbox Team Share'
  },
  {
    id: 's3',
    name: 'Amazon S3',
    desc: 'Standard, Glacier & Vault',
    icon: 'i-simple-icons-amazons3',
    tag: 'IAM Token',
    color: 'amber',
    borderColor: 'border-amber-500/50 bg-amber-500/[0.08] ring-amber-500/30',
    iconColor: 'text-amber-500',
    defaultCapacity: 250,
    defaultLabel: 'AWS S3 Cold Bucket'
  },
  {
    id: 'r2',
    name: 'Cloudflare R2',
    desc: 'Zero-Egress Object Storage',
    icon: 'i-simple-icons-cloudflare',
    tag: 'S3-Compatible',
    color: 'orange',
    borderColor: 'border-orange-500/50 bg-orange-500/[0.08] ring-orange-500/30',
    iconColor: 'text-orange-500',
    defaultCapacity: 50,
    defaultLabel: 'Cloudflare R2 Media Hot'
  },
  {
    id: 'b2',
    name: 'Backblaze B2',
    desc: 'High Durability Backup',
    icon: 'i-lucide-hard-drive',
    tag: 'Cold Storage',
    color: 'rose',
    borderColor: 'border-rose-500/50 bg-rose-500/[0.08] ring-rose-500/30',
    iconColor: 'text-rose-500',
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
  { label: '100 GB', value: 100, tag: 'Popular' },
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
  }
}

async function handleConnect() {
  if (!label.value.trim()) {
    label.value = currentProviderMeta.value.defaultLabel
  }

  isSubmitting.value = true
  await new Promise(resolve => setTimeout(resolve, 600))

  accountsStore.addAccount({
    provider: selectedProvider.value,
    label: label.value,
    email: email.value || `${selectedProvider.value}-user@polycloud.dev`,
    total_bytes: capacityGb.value * 1024 * 1024 * 1024
  })

  isSubmitting.value = false
  isOpen.value = false
}
</script>

<template>
  <UModal
    v-model:open="isOpen"
    :ui="{
      content: 'sm:max-w-2xl bg-card/95 backdrop-blur-2xl border border-default/80 rounded-3xl shadow-2xl p-0 overflow-hidden',
      body: 'p-6 space-y-6',
      footer: 'px-6 py-4 bg-elevated/40 border-t border-default/60'
    }"
  >
    <!-- Custom Modal Header -->
    <template #header>
      <div class="px-6 pt-6 pb-2 flex items-start justify-between gap-4">
        <div class="flex items-start gap-3.5">
          <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-sky-500 to-indigo-600 text-white shadow-lg shadow-sky-500/25">
            <UIcon name="i-lucide-cloud-upload" class="size-6" />
          </div>
          <div>
            <h2 class="text-lg font-black tracking-tight text-highlighted">
              Connect Cloud Storage Account
            </h2>
            <p class="text-xs text-muted mt-1 leading-relaxed max-w-lg">
              Integrate your cloud storage account into Poly Cloud unified index. Automatic metadata synchronization via rclone engine without duplicating local files.
            </p>
          </div>
        </div>

        <UButton
          icon="i-lucide-x"
          color="neutral"
          variant="ghost"
          size="sm"
          square
          class="rounded-xl"
          @click="isOpen = false"
        />
      </div>
    </template>

    <template #body>
      <div class="space-y-5">
        <!-- 1. Provider Selection Grid -->
        <div>
          <div class="flex items-center justify-between mb-2.5">
            <label class="text-xs font-bold uppercase tracking-wider text-muted/90 flex items-center gap-1.5">
              <span>1. Select Cloud Provider</span>
              <span class="text-sky-500 font-normal">({{ providers.length }} supported)</span>
            </label>
            <span class="text-[11px] text-muted font-medium">Selected: <strong class="text-highlighted">{{ currentProviderMeta.name }}</strong></span>
          </div>

          <div class="grid grid-cols-2 sm:grid-cols-3 gap-2.5">
            <button
              v-for="p in providers"
              :key="p.id"
              type="button"
              class="group relative flex flex-col justify-between p-3.5 rounded-2xl border text-left transition-all duration-200 cursor-pointer"
              :class="[
                selectedProvider === p.id
                  ? `${p.borderColor} ring-2 shadow-md`
                  : 'border-default/70 bg-card hover:border-default/90 hover:bg-elevated/50'
              ]"
              @click="selectProvider(p.id as StorageProvider)"
            >
              <div class="flex items-start justify-between w-full mb-3">
                <div class="flex size-9 items-center justify-center rounded-xl bg-elevated/80 group-hover:scale-105 transition-transform">
                  <UIcon :name="p.icon" class="size-5" :class="p.iconColor" />
                </div>
                
                <div class="flex items-center gap-1">
                  <span class="text-[9px] font-semibold px-1.5 py-0.5 rounded-md bg-elevated/80 text-muted">
                    {{ p.tag }}
                  </span>
                  <UIcon
                    v-if="selectedProvider === p.id"
                    name="i-lucide-check-circle-2"
                    class="size-4"
                    :class="p.iconColor"
                  />
                </div>
              </div>

              <div>
                <h4 class="text-xs font-bold text-highlighted leading-snug">{{ p.name }}</h4>
                <p class="text-[10px] text-muted line-clamp-1 mt-0.5">{{ p.desc }}</p>
              </div>
            </button>
          </div>
        </div>

        <!-- 2. Provider Context Banner -->
        <div class="flex items-center gap-3 p-3.5 rounded-2xl border border-sky-500/30 bg-sky-500/[0.06]">
          <UIcon name="i-lucide-shield-check" class="size-5 text-sky-500 shrink-0" />
          <div class="text-xs leading-relaxed">
            <span class="font-bold text-highlighted">End-to-End Token Encryption:</span>
            <span class="text-muted ml-1">
              OAuth credentials and API tokens are encrypted with AES-256-GCM. Poly Cloud never stores provider account passwords.
            </span>
          </div>
        </div>

        <!-- 3. Account Configuration Fields -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <!-- Label -->
          <div class="space-y-1.5">
            <label class="block text-xs font-bold text-highlighted">
              Account Alias / Label
            </label>
            <UInput
              v-model="label"
              placeholder="e.g., GDrive Primary Work"
              icon="i-lucide-tag"
              size="md"
              class="w-full rounded-xl"
            />
            <span class="text-[10px] text-muted block">Display name shown across unified file explorer.</span>
          </div>

          <!-- Email / Identifier -->
          <div class="space-y-1.5">
            <label class="block text-xs font-bold text-highlighted">
              User Email / Bucket Identifier
            </label>
            <UInput
              v-model="email"
              placeholder="name@company.com or bucket-name"
              icon="i-lucide-mail"
              size="md"
              class="w-full rounded-xl"
            />
            <span class="text-[10px] text-muted block">Used for session identity or bucket ARN mapping.</span>
          </div>
        </div>

        <!-- 4. Capacity Presets & Slider -->
        <div class="space-y-2.5">
          <div class="flex items-center justify-between">
            <label class="text-xs font-bold text-highlighted flex items-center gap-1.5">
              <span>Storage Capacity Allocation</span>
            </label>
            <span class="font-mono text-xs font-bold text-sky-500">
              {{ capacityGb >= 1024 ? `${(capacityGb / 1024).toFixed(1)} TB` : `${capacityGb} GB` }}
            </span>
          </div>

          <!-- Quick Preset Chips -->
          <div class="flex flex-wrap items-center gap-1.5">
            <button
              v-for="preset in presetCapacities"
              :key="preset.value"
              type="button"
              class="px-2.5 py-1 rounded-xl text-xs font-medium border transition-all duration-150 flex items-center gap-1.5"
              :class="[
                capacityGb === preset.value
                  ? 'border-sky-500 bg-sky-500 text-white shadow-xs font-bold'
                  : 'border-default/60 bg-elevated/30 text-muted hover:text-highlighted hover:bg-elevated/70'
              ]"
              @click="capacityGb = preset.value"
            >
              <span>{{ preset.label }}</span>
              <span v-if="preset.tag" class="text-[9px] opacity-80 uppercase">({{ preset.tag }})</span>
            </button>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <div class="flex items-center justify-between w-full">
        <span class="text-[11px] text-muted flex items-center gap-1">
          <UIcon name="i-lucide-check-circle-2" class="size-3.5 text-emerald-500" />
          Model A Whole-File & Smart Routing Enabled
        </span>

        <div class="flex items-center gap-2.5">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            class="rounded-xl font-semibold"
            @click="isOpen = false"
          />
          <UButton
            label="Authorize & Connect Account"
            icon="i-lucide-link"
            color="sky"
            variant="solid"
            class="rounded-xl font-bold shadow-md shadow-sky-500/25 px-4"
            :loading="isSubmitting"
            @click="handleConnect"
          />
        </div>
      </div>
    </template>
  </UModal>
</template>
