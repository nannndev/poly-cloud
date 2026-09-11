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
  return 'emerald'
})

const menuItems = computed<DropdownMenuItem[][]>(() => [
  [
    {
      label: 'Sync Metadata Now',
      icon: 'i-lucide-refresh-cw',
      onSelect: () => accountsStore.syncAccount(props.account.id)
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
      onSelect: () => accountsStore.removeAccount(props.account.id)
    }
  ]
])
</script>

<template>
  <div
    class="relative flex flex-col justify-between p-5 rounded-3xl border transition-all duration-200 shadow-xs group overflow-hidden"
    :class="[
      account.status === 'needs_reconnect'
        ? 'border-amber-500/30 bg-[#161411]'
        : 'border-white/[0.08] bg-[#111114] hover:border-emerald-500/30 hover:bg-[#131317]'
    ]"
  >
    <div>
      <!-- Header: Provider Icon, Titles & Status -->
      <div class="flex items-start justify-between gap-3 mb-4">
        <div class="flex items-center gap-3 min-w-0">
          <div
            class="flex size-11 shrink-0 items-center justify-center rounded-2xl border border-white/[0.06] bg-[#16161b] text-lg shadow-xs group-hover:scale-105 transition-transform"
          >
            <UIcon :name="providerMeta.icon" class="size-5" />
          </div>

          <div class="min-w-0 leading-tight">
            <h3 class="font-semibold text-sm text-zinc-100 truncate group-hover:text-emerald-400 transition-colors">
              {{ account.label }}
            </h3>
            <p class="text-[11px] text-zinc-500 truncate mt-0.5 font-mono">
              {{ account.email || providerMeta.name }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-1 shrink-0">
          <UBadge
            v-if="account.status === 'active'"
            label="Connected"
            color="emerald"
            variant="subtle"
            size="xs"
            class="rounded-lg px-2 py-0.5 font-medium text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
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
        <div class="flex items-center justify-between text-xs">
          <div class="flex items-center gap-1.5 text-zinc-400">
            <UIcon name="i-lucide-database" class="size-3.5 text-emerald-400" />
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
          <span class="text-zinc-500">{{ formatBytes(account.used_bytes || 0) }} used</span>
          <span class="text-emerald-400 font-medium">{{ formatBytes(account.free_bytes || 0) }} free</span>
        </div>
      </div>
    </div>

    <!-- Card Footer -->
    <div class="mt-4 pt-3.5 border-t border-white/[0.06] flex items-center justify-between gap-2">
      <div class="flex items-center gap-1.5 text-[11px] text-muted">
        <UIcon name="i-lucide-refresh-cw" class="size-3" :class="[isSyncingThis ? 'animate-spin text-sky-500' : '']" />
        <span>{{ formatDate(account.last_synced || '') }}</span>
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
          @click="accountsStore.reconnectAccount(account.id)"
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
          @click="accountsStore.syncAccount(account.id)"
        />
      </div>
    </div>
  </div>
</template>
