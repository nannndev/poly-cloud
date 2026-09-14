<script setup lang="ts">
import type { CommandPaletteGroup, CommandPaletteItem, NavigationMenuItem } from '@nuxt/ui'

const accountsStore = useAccountsStore()
const filesStore = useFilesStore()
const { formatBytes } = useFormatters()
const { status: healthStatus, label: healthLabel, dotClass: healthDot, check: checkHealth } = useBackendHealth()
const colorMode = useColorMode()

const isSidebarOpen = ref(false)

function toggleColorMode() {
  const isDark = colorMode.value === 'dark' || (import.meta.client && document.documentElement.classList.contains('dark'))
  const next = isDark ? 'light' : 'dark'
  colorMode.preference = next
  if (import.meta.client) {
    if (next === 'light') {
      document.documentElement.classList.remove('dark')
      document.documentElement.classList.add('light')
    } else {
      document.documentElement.classList.remove('light')
      document.documentElement.classList.add('dark')
    }
  }
}

// Status engine dipakai di seluruh halaman, jadi diperiksa sekali di layout dan
// disegarkan berkala — indikator yang tak pernah berubah sama saja dengan hiasan.
onMounted(() => {
  void checkHealth()
  const timer = setInterval(() => { void checkHealth() }, 30_000)
  onBeforeUnmount(() => clearInterval(timer))
})

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
      class="bg-white/95 dark:bg-[#0d111a]/98 backdrop-blur-2xl border-r border-slate-200/80 dark:border-white/[0.07] shadow-xl transition-all duration-200"
      :ui="{
        header: 'p-3.5 border-b border-slate-200/80 dark:border-white/[0.06]',
        body: 'p-3 space-y-4',
        footer: 'p-2.5 border-t border-slate-200/80 dark:border-white/[0.06] bg-slate-50/80 dark:bg-[#111621]'
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
          class="bg-slate-100 dark:bg-[#1a1f2d] border border-slate-200 dark:border-white/[0.08] hover:border-primary-500/40 hover:bg-slate-200/60 dark:hover:bg-[#212736] rounded-xl text-xs py-2 shadow-xs text-slate-700 dark:text-zinc-300 transition-all"
        />

        <!-- Section Label -->
        <div v-if="!collapsed" class="px-1.5 pt-1 flex items-center justify-between">
          <span class="text-[10px] font-extrabold uppercase tracking-wider text-slate-500 dark:text-zinc-500">
            Storage Control Center
          </span>
          <UBadge
            :label="`${accountsStore.activeAccounts.length} Connected`"
            color="primary"
            variant="subtle"
            size="xs"
            class="text-[9px] px-1.5 py-0.5 font-semibold bg-primary-500/10 text-primary-600 dark:text-primary-400 border border-primary-500/20"
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
            class="rounded-xl font-bold shadow-xs bg-primary-600 hover:bg-primary-500 text-white transition-all cursor-pointer"
            @click="filesStore.isUploadModalOpen = true"
          />
        </div>

        <!-- Aggregate Storage Mini Widget -->
        <div
          v-if="!collapsed"
          class="mt-auto rounded-2xl border border-slate-200/80 dark:border-white/[0.08] bg-slate-100/80 dark:bg-[#151a27] p-3 space-y-2.5 shadow-sm"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1.5">
              <UIcon name="i-lucide-hard-drive" class="size-4 text-primary-500 dark:text-primary-400" />
              <span class="text-xs font-bold text-slate-800 dark:text-zinc-200">Unified Capacity</span>
            </div>
            <span class="text-[11px] font-bold text-primary-600 dark:text-primary-400">{{ accountsStore.usagePercent }}%</span>
          </div>

          <UProgress :model-value="accountsStore.usagePercent" color="primary" size="sm" />

          <div class="flex items-center justify-between text-[11px] text-slate-500 dark:text-zinc-400 font-mono">
            <span>{{ formatBytes(accountsStore.usedStorage) }} used</span>
            <span class="text-slate-800 dark:text-zinc-200 font-semibold">{{ formatBytes(accountsStore.totalStorage) }}</span>
          </div>
        </div>
      </template>

      <template #footer="{ collapsed }">
        <div class="space-y-1.5">
          <UserMenu :collapsed="collapsed" />

          <div v-if="!collapsed" class="flex items-center justify-between px-2 pt-1 border-t border-slate-200/70 dark:border-white/[0.05]">
            <!-- Status nyata dari /healthz; klik untuk memeriksa ulang. -->
            <button
              type="button"
              :title="`${healthLabel} — click to re-check`"
              class="flex items-center gap-1.5 text-[10px] text-slate-500 dark:text-zinc-500 font-mono cursor-pointer hover:text-slate-700 dark:hover:text-zinc-300 transition-colors"
              @click="checkHealth"
            >
              <span class="size-1.5 rounded-full" :class="healthDot" />
              <span>{{ healthLabel }}</span>
              <UIcon
                v-if="healthStatus !== 'online'"
                name="i-lucide-refresh-cw"
                class="size-3 text-slate-400 dark:text-zinc-500"
              />
            </button>

            <!-- Quick Theme Switcher Button -->
            <button
              type="button"
              :title="colorMode.value === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'"
              class="flex size-6 items-center justify-center rounded-lg text-slate-500 hover:text-slate-900 hover:bg-slate-200/60 dark:text-zinc-400 dark:hover:text-white dark:hover:bg-white/[0.08] transition-colors cursor-pointer"
              @click="toggleColorMode"
            >
              <UIcon
                :name="colorMode.value === 'dark' ? 'i-lucide-sun' : 'i-lucide-moon'"
                class="size-3.5"
              />
            </button>
          </div>
        </div>
      </template>
    </UDashboardSidebar>

    <!-- Global Search Palette -->
    <UDashboardSearch :groups="searchGroups" />

    <!-- Main Content Area: Responsive Light & Dark Canvas -->
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden bg-slate-50 dark:bg-[#0b0e14] text-slate-900 dark:text-slate-100 transition-colors duration-200">
      <slot />
    </div>

    <!-- Upload Modal Dialog -->
    <UploadModal />
  </UDashboardGroup>
</template>
