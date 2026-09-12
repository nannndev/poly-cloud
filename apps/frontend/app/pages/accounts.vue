<script setup lang="ts">
const route = useRoute()
const accountsStore = useAccountsStore()
const { formatBytes } = useFormatters()
const toast = useToast()

// ?connect=1 dipakai halaman callback untuk membuka modal setelah percobaan gagal.
const isConnectModalOpen = ref(route.query.connect === '1')
const selectedFilter = ref<'all' | 'active' | 'attention'>('all')
const isSyncingAll = ref(false)

const filteredAccounts = computed(() => {
  if (selectedFilter.value === 'active') {
    return accountsStore.accounts.filter(a => a.status === 'active')
  }
  if (selectedFilter.value === 'attention') {
    return accountsStore.accounts.filter(a => a.status !== 'active')
  }
  return accountsStore.accounts
})

const { refresh: reloadAccounts } = await useAsyncData(
  'accounts-page',
  () => accountsStore.loadAll().then(() => true),
  { server: false, default: () => false }
)

async function handleSyncAll() {
  isSyncingAll.value = true
  let failed = 0
  // Berurutan supaya provider tak dibanjiri permintaan sekaligus; satu akun
  // gagal tak boleh menghentikan sisanya.
  for (const acc of accountsStore.accounts) {
    try {
      await accountsStore.syncAccount(acc.id)
    } catch {
      failed++
    }
  }
  isSyncingAll.value = false

  if (failed > 0) {
    toast.add({ title: `${failed} account${failed > 1 ? 's' : ''} failed to sync`, color: 'warning' })
  } else if (accountsStore.accounts.length > 0) {
    toast.add({ title: 'All accounts synced', color: 'success' })
  }
}
</script>

<template>
  <div class="flex flex-1 min-w-0 min-h-0 w-full">
    <UDashboardPanel id="accounts-panel">
      <template #header>
        <UDashboardNavbar title="Connected Cloud Accounts" :ui="{ right: 'gap-2.5' }">
          <template #leading>
            <UDashboardSidebarCollapse />
          </template>

          <template #title>
            <div class="flex items-center gap-2">
              <h1 class="font-bold text-base text-highlighted">Connected Cloud Accounts</h1>
              <UBadge
                :label="`${accountsStore.accounts.length} Connected Accounts`"
                color="primary"
                variant="subtle"
                size="xs"
                class="rounded-lg font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
              />
            </div>
          </template>

          <template #right>
            <UButton
              label="Sync All"
              icon="i-lucide-refresh-cw"
              color="neutral"
              variant="outline"
              class="rounded-xl text-xs font-medium border-white/[0.08] text-zinc-300 hover:text-white"
              :loading="isSyncingAll"
              @click="handleSyncAll"
            />
            <UButton
              label="Connect New Account"
              icon="i-lucide-plus"
              color="primary"
              variant="solid"
              class="rounded-xl font-semibold shadow-xs bg-emerald-600 hover:bg-emerald-500 text-white px-3.5 transition-all cursor-pointer"
              @click="isConnectModalOpen = true"
            />
          </template>
        </UDashboardNavbar>
      </template>

      <template #body>
        <div class="space-y-6 p-1">
          <!-- Summary Hero Card: Clean Matte Black with Soft Emerald Accent -->
          <div
            class="relative flex flex-col md:flex-row items-start md:items-center justify-between gap-5 p-6 rounded-3xl border border-white/[0.08] bg-[#111114] shadow-xs"
          >
            <div class="flex items-center gap-4">
              <div class="flex size-13 shrink-0 items-center justify-center rounded-2xl bg-emerald-500/15 text-emerald-400 border border-emerald-500/25">
                <UIcon name="i-lucide-cloud-cog" class="size-6" />
              </div>
              <div class="space-y-0.5">
                <div class="flex items-center gap-2">
                  <h3 class="font-bold text-base text-zinc-100">Multi-Provider Storage Mesh</h3>
                  <UBadge label="Zero Vendor Lock-in" color="primary" variant="subtle" size="xs" class="rounded-md bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[10px]" />
                </div>
                <p class="text-xs text-zinc-400 max-w-xl leading-relaxed">
                  Poly Cloud aggregates Google Drive, OneDrive, Dropbox, and AWS S3 into a single unified virtual drive with automated smart routing and deduplication.
                </p>
              </div>
            </div>

            <div class="flex items-center gap-6 text-xs shrink-0 self-stretch md:self-auto justify-between md:justify-end border-t md:border-t-0 border-white/[0.06] pt-3 md:pt-0">
              <div class="text-left md:text-right">
                <span class="text-zinc-500 block text-[10px] uppercase font-bold tracking-wider">Aggregated Total</span>
                <span class="font-mono text-base font-bold text-zinc-200">{{ formatBytes(accountsStore.totalStorage) }}</span>
              </div>
              <div class="text-left md:text-right">
                <span class="text-zinc-500 block text-[10px] uppercase font-bold tracking-wider">Free Capacity</span>
                <span class="font-mono text-base font-bold text-emerald-400">{{ formatBytes(accountsStore.freeStorage) }}</span>
              </div>
            </div>
          </div>

          <!-- Filter Pills Toolbar -->
          <div class="flex items-center justify-between border-b border-white/[0.06] pb-3.5">
            <div class="flex items-center gap-1.5">
              <UButton
                label="All Providers"
                :badge="`${accountsStore.accounts.length}`"
                size="xs"
                :color="selectedFilter === 'all' ? 'primary' : 'neutral'"
                :variant="selectedFilter === 'all' ? 'solid' : 'ghost'"
                class="rounded-xl font-medium"
                @click="selectedFilter = 'all'"
              />
              <UButton
                label="Active & Synced"
                :badge="`${accountsStore.activeAccounts.length}`"
                size="xs"
                :color="selectedFilter === 'active' ? 'primary' : 'neutral'"
                :variant="selectedFilter === 'active' ? 'solid' : 'ghost'"
                class="rounded-xl font-medium"
                @click="selectedFilter = 'active'"
              />
              <UButton
                label="Needs Attention"
                :badge="`${accountsStore.accounts.length - accountsStore.activeAccounts.length}`"
                size="xs"
                :color="selectedFilter === 'attention' ? 'warning' : 'neutral'"
                :variant="selectedFilter === 'attention' ? 'solid' : 'ghost'"
                class="rounded-xl font-medium"
                @click="selectedFilter = 'attention'"
              />
            </div>

            <span class="text-xs text-zinc-500 hidden sm:inline-block">
              {{ accountsStore.activeAccounts.length }} of {{ accountsStore.accounts.length }} accounts operational
            </span>
          </div>

          <!-- Backend tak terjangkau tak boleh tampak seperti "belum ada akun". -->
          <StateNotice
            v-if="accountsStore.isLoading && accountsStore.accounts.length === 0"
            variant="loading"
            title="Loading connected accounts..."
          />

          <StateNotice
            v-else-if="accountsStore.loadError && accountsStore.accounts.length === 0"
            variant="error"
            title="Could not load your accounts"
            :description="accountsStore.loadError"
            @retry="reloadAccounts"
          />

          <!-- Accounts Grid -->
          <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            <AccountCard
              v-for="acc in filteredAccounts"
              :key="acc.id"
              :account="acc"
            />

            <!-- Connect Account CTA Card -->
            <button
              type="button"
              class="flex flex-col items-center justify-center p-8 rounded-3xl border-2 border-dashed border-white/[0.1] hover:border-emerald-500/40 bg-[#111114] hover:bg-[#141418] transition-all text-center min-h-[220px] group cursor-pointer"
              @click="isConnectModalOpen = true"
            >
              <div class="flex size-11 items-center justify-center rounded-2xl bg-zinc-800 text-zinc-400 group-hover:bg-emerald-500/15 group-hover:text-emerald-400 transition-all mb-3 group-hover:scale-105">
                <UIcon name="i-lucide-cloud-upload" class="size-5" />
              </div>
              <h4 class="font-semibold text-xs text-zinc-200 group-hover:text-emerald-400 transition-colors">
                Connect New Cloud Account
              </h4>
              <p class="text-[11px] text-zinc-500 mt-1 max-w-xs leading-normal">
                Google Drive, OneDrive, Dropbox, AWS S3, Cloudflare R2, Backblaze B2
              </p>
            </button>
          </div>
        </div>
      </template>
    </UDashboardPanel>

    <!-- Connect Account Modal -->
    <ConnectAccountModal v-model:open="isConnectModalOpen" />
  </div>
</template>
