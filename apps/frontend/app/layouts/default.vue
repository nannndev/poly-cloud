<script setup lang="ts">
import type { CommandPaletteGroup, CommandPaletteItem, NavigationMenuItem } from '@nuxt/ui'

const route = useRoute()
const accountsStore = useAccountsStore()
const filesStore = useFilesStore()
const { formatBytes } = useFormatters()

const isSidebarOpen = ref(false)

const navLinks = computed<NavigationMenuItem[]>(() => [
  {
    label: 'Unified File Explorer',
    icon: 'i-lucide-folder-archive',
    to: '/',
    exact: true,
    onSelect: () => { isSidebarOpen.value = false }
  },
  {
    label: 'Connected Accounts',
    icon: 'i-lucide-cloud-cog',
    to: '/accounts',
    badge: `${accountsStore.accounts.length}`,
    onSelect: () => { isSidebarOpen.value = false }
  },
  {
    label: 'Storage & Quota',
    icon: 'i-lucide-pie-chart',
    to: '/quota',
    badge: `${accountsStore.usagePercent}%`,
    onSelect: () => { isSidebarOpen.value = false }
  },
  {
    label: 'Smart Routing & Settings',
    icon: 'i-lucide-sliders-horizontal',
    to: '/settings',
    onSelect: () => { isSidebarOpen.value = false }
  }
])

const searchItems = computed<CommandPaletteItem[]>(() => [
  ...navLinks.value.map(item => ({
    id: String(item.to || item.label),
    label: String(item.label),
    icon: item.icon,
    to: item.to
  })),
  {
    id: 'act-upload',
    label: 'Upload New File to Multi-Cloud',
    icon: 'i-lucide-upload-cloud',
    onSelect: () => { filesStore.isUploadModalOpen = true }
  },
  {
    id: 'act-add-account',
    label: 'Connect New Cloud Provider (GDrive, OneDrive, S3, R2)',
    icon: 'i-lucide-cloud-upload',
    to: '/accounts'
  },
  {
    id: 'act-sync-all',
    label: 'Resync All Cloud Metadata Catalogs',
    icon: 'i-lucide-refresh-cw',
    to: '/accounts'
  }
])

const searchGroups = computed<CommandPaletteGroup<CommandPaletteItem>[]>(() => [
  {
    id: 'navigation',
    label: 'Poly Cloud Navigation',
    items: searchItems.value
  }
])
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar
      id="poly-sidebar"
      v-model:open="isSidebarOpen"
      collapsible
      resizable
      class="bg-[#0c0c0e]/98 backdrop-blur-2xl border-r border-white/[0.07] shadow-xl transition-all duration-200"
      :ui="{
        header: 'p-3.5 border-b border-white/[0.06]',
        body: 'p-3 space-y-4',
        footer: 'p-2.5 border-t border-white/[0.06] bg-[#0e0e11]'
      }"
    >
      <template #header="{ collapsed }">
        <AppBrand :compact="collapsed" />
      </template>

      <template #default="{ collapsed }">
        <!-- Quick Search Modal Button -->
        <UDashboardSearchButton
          :collapsed="collapsed"
          label="Search files or modules (⌘K)..."
          class="bg-[#141418] border border-white/[0.08] hover:border-emerald-500/40 hover:bg-[#18181d] rounded-xl text-xs py-2 shadow-xs text-zinc-300 transition-all"
        />

        <!-- Section Label -->
        <div v-if="!collapsed" class="px-1.5 pt-1 flex items-center justify-between">
          <span class="text-[10px] font-extrabold uppercase tracking-wider text-zinc-500">
            Storage Control Center
          </span>
          <UBadge
            :label="`${accountsStore.activeAccounts.length} Connected`"
            color="primary"
            variant="subtle"
            size="xs"
            class="text-[9px] px-1.5 py-0.5 font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
          />
        </div>

        <!-- Main Navigation Links -->
        <UNavigationMenu
          :collapsed="collapsed"
          :items="navLinks"
          orientation="vertical"
          tooltip
          popover
          class="space-y-1"
        />

        <!-- Quick Upload Action Button -->
        <div v-if="!collapsed" class="pt-2">
          <UButton
            icon="i-lucide-upload-cloud"
            label="Upload File"
            block
            color="primary"
            variant="solid"
            class="rounded-xl font-bold shadow-xs bg-emerald-600 hover:bg-emerald-500 text-white transition-all cursor-pointer"
            @click="filesStore.isUploadModalOpen = true"
          />
        </div>

        <!-- Aggregate Storage Mini Widget -->
        <div
          v-if="!collapsed"
          class="mt-auto rounded-2xl border border-white/[0.08] bg-[#121215] p-3 space-y-2.5 shadow-sm"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1.5">
              <UIcon name="i-lucide-hard-drive" class="size-4 text-emerald-400" />
              <span class="text-xs font-bold text-zinc-200">Unified Capacity</span>
            </div>
            <span class="text-[11px] font-bold text-emerald-400">{{ accountsStore.usagePercent }}%</span>
          </div>

          <UProgress :model-value="accountsStore.usagePercent" color="primary" size="sm" />

          <div class="flex items-center justify-between text-[11px] text-zinc-400 font-mono">
            <span>{{ formatBytes(accountsStore.usedStorage) }} used</span>
            <span class="text-zinc-200 font-semibold">{{ formatBytes(accountsStore.totalStorage) }}</span>
          </div>
        </div>
      </template>

      <template #footer="{ collapsed }">
        <div class="space-y-1.5">
          <UserMenu :collapsed="collapsed" />
          <div v-if="!collapsed" class="px-2 pt-1 border-t border-white/[0.05] flex items-center justify-between text-[10px] text-zinc-500 font-mono select-none">
            <span class="flex items-center gap-1.5">
              <span class="size-1.5 rounded-full bg-emerald-400" />
              Engine Online
            </span>
            <span class="px-1.5 py-0.5 rounded bg-zinc-800/60 text-zinc-400 border border-white/[0.06] text-[9px] font-medium">v1.0.4</span>
          </div>
        </div>
      </template>
    </UDashboardSidebar>

    <!-- Global Search Palette -->
    <UDashboardSearch :groups="searchGroups" />

    <!-- Main Content Area: Pure Matte Deep Black -->
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden bg-[#09090b]">
      <slot />
    </div>

    <!-- Upload Modal Dialog -->
    <UploadModal />
  </UDashboardGroup>
</template>
