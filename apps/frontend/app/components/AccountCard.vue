<script setup lang="ts">
import type { Account } from '~/types'
import type { DropdownMenuItem } from '@nuxt/ui'

const props = defineProps<{
  account: Account
}>()

const accountsStore = useAccountsStore()
const { formatBytes, formatDate, getProviderMeta } = useFormatters()

const providerMeta = computed(() => getProviderMeta(props.account.provider))

const usagePercent = computed(() => {
  if (!props.account.total_bytes) return 0
  return Math.round(((props.account.used_bytes || 0) / props.account.total_bytes) * 100)
})

const isSyncingThis = computed(() => accountsStore.isSyncing[props.account.id] || false)

const progressColor = computed(() => {
  if (props.account.status === 'needs_reconnect') return 'warning'
  if (usagePercent.value > 90) return 'error'
  if (usagePercent.value > 75) return 'warning'
  return 'primary'
})

const toast = useToast()
const isDisconnectOpen = ref(false)
const isDisconnecting = ref(false)
const isReconnecting = ref(false)

const menuItems = computed<DropdownMenuItem[][]>(() => [
  [
    {
      label: 'Sync Metadata Now',
      icon: 'i-lucide-refresh-cw',
      onSelect: () => { void handleSync() }
    },
    {
      label: 'Copy Account ID',
      icon: 'i-lucide-copy',
      onSelect: () => navigator.clipboard?.writeText(props.account.id)
    }
  ],
  [
    {
      label: 'Disconnect Account',
      icon: 'i-lucide-trash-2',
      color: 'error',
      onSelect: () => { isDisconnectOpen.value = true }
    }
  ]
])

async function handleSync() {
  try {
    const res = await accountsStore.syncAccount(props.account.id)
    toast.add({
      title: `${props.account.label} synced`,
      description: `${res.files_indexed} files indexed.`,
      color: 'success'
    })
  } catch (err) {
    toast.add({ title: 'Sync failed', description: friendlyMessage(err), color: 'error' })
  }
}

async function handleDisconnect() {
  isDisconnecting.value = true
  try {
    await accountsStore.removeAccount(props.account.id)
    isDisconnectOpen.value = false
    toast.add({ title: `${props.account.label} disconnected`, color: 'success' })
  } catch (err) {
    toast.add({ title: 'Could not disconnect account', description: friendlyMessage(err), color: 'error' })
  } finally {
    isDisconnecting.value = false
  }
}

// Token yang tak lagi berlaku hanya bisa dipulihkan lewat consent ulang.
async function handleReconnect() {
  isReconnecting.value = true
  try {
    const authUrl = await accountsStore.reconnectAccount(props.account.id)
    window.location.href = authUrl
  } catch (err) {
    toast.add({ title: 'Reconnect failed', description: friendlyMessage(err), color: 'error' })
    isReconnecting.value = false
  }
}
</script>

<template>
  <div
    class="relative flex flex-col justify-between p-5 rounded-3xl border transition-all duration-200 shadow-xs group overflow-hidden"
    :class="[
      account.status === 'needs_reconnect'
        ? 'border-amber-500/30 bg-[#1f1b14]'
        : 'border-white/[0.08] bg-[#141925] hover:border-primary-500/30 hover:bg-[#171c2a]'
    ]"
  >
    <div>
      <!-- Header: Provider Icon, Titles & Status -->
      <div class="flex items-start justify-between gap-3 mb-4">
        <div class="flex items-center gap-3 min-w-0">
          <div
            class="flex size-11 shrink-0 items-center justify-center rounded-2xl border border-white/[0.06] bg-[#1c2231] text-lg shadow-xs group-hover:scale-105 transition-transform"
          >
            <UIcon :name="providerMeta.icon" class="size-5" />
          </div>

          <div class="min-w-0 leading-tight">
            <h3 class="font-semibold text-sm text-zinc-100 truncate group-hover:text-primary-400 transition-colors">
              {{ account.label }}
            </h3>
            <div class="flex items-center gap-1.5 mt-1 flex-wrap">
              <span class="text-[11px] text-zinc-500 truncate">{{ providerMeta.name }}</span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-1 shrink-0">
          <UBadge
            v-if="account.status === 'active'"
            label="Connected"
            color="primary"
            variant="subtle"
            size="xs"
            class="rounded-lg px-2 py-0.5 font-medium text-[10px] bg-primary-500/10 text-primary-400 border border-primary-500/20"
          />
          <UBadge
            v-else-if="account.status === 'needs_reconnect'"
            label="Needs Reconnect"
            color="warning"
            variant="subtle"
            size="xs"
            class="rounded-lg px-2 py-0.5 font-medium text-[10px]"
          />
          <UBadge
            v-else-if="account.status === 'syncing'"
            label="Syncing"
            color="info"
            variant="subtle"
            size="xs"
            class="rounded-lg px-2 py-0.5 font-medium text-[10px]"
          />
          <UBadge
            v-else
            label="Error"
            color="error"
            variant="subtle"
            size="xs"
            class="rounded-lg px-2 py-0.5 font-medium text-[10px]"
          />

          <UDropdownMenu :items="menuItems" :content="{ align: 'end' }">
            <UButton
              icon="i-lucide-more-vertical"
              color="neutral"
              variant="ghost"
              size="xs"
              square
              class="rounded-xl text-zinc-400 hover:text-white"
            />
          </UDropdownMenu>
        </div>
      </div>

      <!-- Storage Metrics & Visual Gauge -->
      <div class="p-3.5 rounded-2xl border border-white/[0.06] bg-zinc-900/60 space-y-2.5">
        <!-- Sebagian provider (mis. S3) tak melaporkan kuota; jangan tampilkan 0% seolah kosong. -->
        <template v-if="account.total_bytes > 0">
          <div class="flex items-center justify-between text-xs">
            <div class="flex items-center gap-1.5 text-zinc-400">
              <UIcon name="i-lucide-database" class="size-3.5 text-primary-400" />
              <span>Capacity Utilized</span>
            </div>
            <span class="font-semibold font-mono text-zinc-200">{{ usagePercent }}%</span>
          </div>

          <UProgress
            :model-value="usagePercent"
            :color="progressColor"
            size="sm"
            class="rounded-full"
          />

          <div class="flex items-center justify-between text-[11px] font-mono">
            <span class="text-zinc-500">{{ formatBytes(account.used_bytes) }} used</span>
            <span class="text-primary-400 font-medium">{{ formatBytes(account.free_bytes) }} free</span>
          </div>
        </template>

        <div v-else class="flex items-center gap-2 text-xs text-zinc-400">
          <UIcon name="i-lucide-infinity" class="size-3.5 text-zinc-500" />
          <span>This provider does not report quota</span>
        </div>
      </div>
    </div>

    <!-- Card Footer -->
    <div class="mt-4 pt-3.5 border-t border-white/[0.06] flex items-center justify-between gap-2">
      <div class="flex items-center gap-1.5 text-[11px] text-muted">
        <UIcon name="i-lucide-refresh-cw" class="size-3" :class="[isSyncingThis ? 'animate-spin text-sky-500' : '']" />
        <span>{{ account.last_synced ? formatDate(account.last_synced) : 'Never synced' }}</span>
      </div>

      <div class="flex items-center gap-2">
        <UButton
          v-if="account.status === 'needs_reconnect'"
          label="Reconnect"
          icon="i-lucide-link"
          color="warning"
          variant="solid"
          size="xs"
          class="rounded-xl font-bold"
          :loading="isReconnecting"
          @click="handleReconnect"
        />
        <UButton
          v-else
          label="Sync Now"
          icon="i-lucide-refresh-cw"
          color="neutral"
          variant="subtle"
          size="xs"
          class="rounded-xl font-medium"
          :loading="isSyncingThis"
          @click="handleSync"
        />
      </div>
    </div>

    <!-- Mencabut akun menghapus remote rclone dan index-nya, tapi tak menyentuh
         file yang ada di provider. -->
    <UModal
      v-model:open="isDisconnectOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-3',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold flex items-center gap-2 text-rose-400">
            <UIcon name="i-lucide-unlink" class="size-5" />
            Disconnect Account
          </h3>
        </div>
      </template>

      <template #body>
        <p class="text-xs text-zinc-400 leading-relaxed">
          <strong class="text-white">{{ account.label }}</strong> will be detached from the
          platform: its tokens and file index are deleted. Files already stored at the
          provider <strong class="text-zinc-200">are left untouched</strong>.
        </p>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton label="Cancel" color="neutral" variant="ghost" class="rounded-xl" :disabled="isDisconnecting" @click="isDisconnectOpen = false" />
          <UButton label="Disconnect" color="error" class="rounded-xl font-bold" :loading="isDisconnecting" @click="handleDisconnect" />
        </div>
      </template>
    </UModal>
  </div>
</template>
