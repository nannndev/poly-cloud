<script setup lang="ts">
import type { BackendSettings, ApiKey, CreatedApiKey } from '~/types'

const config = useRuntimeConfig()
const accountsStore = useAccountsStore()
const filesStore = useFilesStore()
const api = useApi()
const toast = useToast()
const { formatBytes } = useFormatters()

const apiBase = computed(() => config.public.apiBase)

// Konfigurasi ini berasal dari environment backend saat proses start; mengubahnya
// butuh menyunting .env lalu me-restart, jadi halaman ini menampilkan, bukan menyimpan.
const settings = ref<BackendSettings | null>(null)

// Status kesehatan dibagi dengan indikator di sidebar — satu sumber, supaya
// keduanya tak pernah melaporkan hal yang berbeda.
const {
  status: backendStatus,
  latency: backendLatency,
  detail: healthDetail,
  isChecking: isTesting,
  check: testBackend
} = useBackendHealth()

const isReindexing = ref(false)

/** Sinkronkan ulang seluruh akun: baca kuota & isi terbaru dari tiap provider. */
async function triggerReindex() {
  isReindexing.value = true
  let failed = 0
  let indexed = 0
  for (const acc of accountsStore.accounts) {
    try {
      const res = await accountsStore.syncAccount(acc.id)
      indexed += res.files_indexed
    } catch {
      failed++
    }
  }
  await filesStore.loadAll().catch(() => null)
  isReindexing.value = false

  toast.add({
    title: failed > 0 ? `${failed} account${failed > 1 ? 's' : ''} failed to sync` : 'Index refreshed',
    description: `${indexed} files indexed across ${accountsStore.accounts.length} accounts.`,
    color: failed > 0 ? 'warning' : 'success'
  })
}

const strategyMeta: Record<string, { label: string; desc: string }> = {
  'most-free': {
    label: 'Most Free Space',
    desc: 'Files go to the account with the most free space, spreading usage evenly.'
  },
  'round-robin': {
    label: 'Round Robin',
    desc: 'Destination accounts are picked in turn, without regard to free space.'
  }
}

const activeStrategy = computed(() => {
  const key = settings.value?.routing_strategy || 'most-free'
  return { key, ...(strategyMeta[key] || { label: key, desc: 'Custom strategy.' }) }
})

// MCP & AI Assistant Integration state
const mcpTab = ref<'claude' | 'cursor' | 'http'>('claude')
const mcpCopied = ref(false)

const mcpClaudeConfig = computed(() => JSON.stringify({
  mcpServers: {
    "poly-cloud": {
      command: "go",
      args: ["run", "./apps/backend/cmd/mcp"],
      env: {
        POLYCLOUD_API_URL: apiBase.value || "http://localhost:8080"
      }
    }
  }
}, null, 2))

const mcpCursorEnv = computed(() => `POLYCLOUD_API_URL=${apiBase.value || 'http://localhost:8080'}`)
const mcpEndpointUrl = computed(() => `${apiBase.value || 'http://localhost:8080'}/mcp`)

async function copyMcpConfig(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    mcpCopied.value = true
    toast.add({
      title: 'Copied to clipboard',
      description: 'MCP configuration is ready to paste into your AI assistant.',
      color: 'success'
    })
    setTimeout(() => { mcpCopied.value = false }, 2000)
  } catch {
    // fallback
  }
}

const mcpTools = [
  { name: 'search_files', desc: 'Instant search across all clouds via PostgreSQL index', icon: 'i-lucide-search' },
  { name: 'list_files', desc: 'Explore virtual folders and indexed files in any directory', icon: 'i-lucide-folder-tree' },
  { name: 'read_file', desc: 'Download & stream file content directly into model context', icon: 'i-lucide-file-text' },
  { name: 'upload_file', desc: 'Auto-routed cloud upload to whichever account is emptiest', icon: 'i-lucide-upload-cloud' },
  { name: 'get_storage_quota', desc: 'Inspect live capacity breakdown across all cloud providers', icon: 'i-lucide-pie-chart' },
  { name: 'list_accounts', desc: 'Check connected cloud provider remotes and statuses', icon: 'i-lucide-cloud-cog' },
  { name: 'create_folder', desc: 'Create new virtual directories in the unified namespace', icon: 'i-lucide-folder-plus' },
  { name: 'delete_file', desc: 'Remove obsolete files from cloud storage safely', icon: 'i-lucide-trash-2' }
]

// Developer API Keys State
const apiKeys = ref<ApiKey[]>([])
const isLoadingKeys = ref(false)
const isCreateKeyOpen = ref(false)
const isSecretRevealOpen = ref(false)
const isSubmittingKey = ref(false)
const newKeyName = ref('')
const newKeyScopes = ref<string[]>(['read', 'write'])
const createdSecretKey = ref<CreatedApiKey | null>(null)
const copiedSecret = ref(false)
const deletingKeyId = ref<string | null>(null)

async function loadApiKeys() {
  isLoadingKeys.value = true
  try {
    const res = await api.get<ApiKey[]>('/api-keys')
    apiKeys.value = res || []
  } catch {
    apiKeys.value = []
  } finally {
    isLoadingKeys.value = false
  }
}

async function handleCreateKey() {
  if (!newKeyName.value.trim()) return
  isSubmittingKey.value = true
  try {
    const res = await api.post<CreatedApiKey>('/api-keys', {
      name: newKeyName.value.trim(),
      scopes: newKeyScopes.value
    })
    createdSecretKey.value = res
    isCreateKeyOpen.value = false
    newKeyName.value = ''
    newKeyScopes.value = ['read', 'write']
    isSecretRevealOpen.value = true
    await loadApiKeys()
    toast.add({
      title: 'API Key Created',
      description: `Key "${res.name}" is ready for use.`,
      color: 'success'
    })
  } catch (err: any) {
    toast.add({
      title: 'Failed to create key',
      description: err?.message || 'Error occurred while creating key',
      color: 'error'
    })
  } finally {
    isSubmittingKey.value = false
  }
}

async function handleDeleteKey(key: ApiKey) {
  deletingKeyId.value = key.id
  try {
    await api.del(`/api-keys/${key.id}`)
    apiKeys.value = apiKeys.value.filter(k => k.id !== key.id)
    toast.add({
      title: 'API Key Revoked',
      description: `"${key.name}" has been revoked permanently.`,
      color: 'success'
    })
  } catch (err: any) {
    toast.add({
      title: 'Failed to revoke key',
      description: err?.message || 'Error occurred while revoking key',
      color: 'error'
    })
  } finally {
    deletingKeyId.value = null
  }
}

async function copySecretKey() {
  if (!createdSecretKey.value?.key) return
  try {
    await navigator.clipboard.writeText(createdSecretKey.value.key)
    copiedSecret.value = true
    toast.add({
      title: 'Secret Key Copied',
      description: 'API key copied to clipboard.',
      color: 'success'
    })
    setTimeout(() => { copiedSecret.value = false }, 2000)
  } catch {
    // fallback
  }
}

function formatDate(iso: string | null) {
  if (!iso) return 'Never'
  const d = new Date(iso)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

await useAsyncData('settings-page', async () => {
  await Promise.all([
    accountsStore.loadAll().catch(() => null),
    api.get<BackendSettings>('/settings').then(res => { settings.value = res }).catch(() => null),
    loadApiKeys().catch(() => null),
    testBackend()
  ])
  return true
}, { server: false, default: () => false })
</script>

<template>
  <div class="flex flex-1 min-w-0 min-h-0 w-full">
    <UDashboardPanel id="settings-panel">
      <template #header>
        <UDashboardNavbar title="Smart Routing & Engine Settings" :ui="{ right: 'gap-2.5' }">
          <template #leading>
            <UDashboardSidebarCollapse />
          </template>

          <template #title>
            <div class="flex items-center gap-2.5">
              <h1 class="font-bold text-base text-highlighted">
                Smart Routing & Engine Settings
              </h1>
              <UBadge
                label="System Architecture"
                color="primary"
                variant="subtle"
                size="xs"
                class="rounded-lg font-medium bg-primary-500/10 text-primary-400 border border-primary-500/20"
              />
            </div>
          </template>

          <template #right>
            <UButton
              label="Re-check"
              icon="i-lucide-refresh-cw"
              color="neutral"
              variant="outline"
              class="rounded-xl font-semibold shadow-xs px-4 transition-all cursor-pointer"
              :loading="isTesting"
              @click="testBackend"
            />
          </template>
        </UDashboardNavbar>
      </template>

      <template #body>
        <!-- Centered responsive container that gracefully fills wide monitors -->
        <div class="w-full max-w-6xl mx-auto space-y-6 py-2 pb-16">
          <!-- Row 1: Backend Connection & Maintenance (2 Columns) -->
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
            <!-- Backend Connection (2 Cols) -->
            <div class="lg:col-span-2 p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-5">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-default/60">
                <div class="flex items-start gap-3.5">
                  <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-sky-500/10 text-sky-500 border border-sky-500/20">
                    <UIcon name="i-lucide-server" class="size-6" />
                  </div>
                  <div>
                    <h3 class="font-bold text-sm text-highlighted">
                      Go Backend API & Rclone Engine Connectivity
                    </h3>
                    <p class="text-xs text-muted mt-0.5 leading-relaxed">
                      REST communication between Nuxt 4 frontend, rclone backend engine, and PostgreSQL database.
                    </p>
                  </div>
                </div>

                <div class="flex items-center gap-2 shrink-0">
                  <div
                    class="flex items-center gap-2 px-3 py-1 rounded-xl border text-xs font-semibold"
                    :class="{
                      'border-primary-500/30 bg-primary-500/10 text-primary-500': backendStatus === 'online',
                      'border-amber-500/30 bg-amber-500/10 text-amber-500': backendStatus === 'degraded',
                      'border-rose-500/30 bg-rose-500/10 text-rose-500': backendStatus === 'offline',
                      'border-default/70 bg-elevated/40 text-muted': backendStatus === 'checking'
                    }"
                  >
                    <span
                      class="size-2 rounded-full"
                      :class="{
                        'bg-primary-500 animate-pulse': backendStatus === 'online',
                        'bg-amber-500': backendStatus === 'degraded',
                        'bg-rose-500': backendStatus === 'offline',
                        'bg-zinc-500 animate-pulse': backendStatus === 'checking'
                      }"
                    />
                    <span>
                      {{ backendStatus === 'online' ? 'Connected'
                        : backendStatus === 'degraded' ? 'Partially degraded'
                        : backendStatus === 'offline' ? 'Unreachable'
                        : 'Checking...' }}
                    </span>
                  </div>

                  <UButton
                    icon="i-lucide-refresh-cw"
                    color="neutral"
                    variant="ghost"
                    size="xs"
                    square
                    class="rounded-xl"
                    :loading="isTesting"
                    @click="testBackend"
                  />
                </div>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div class="sm:col-span-2 space-y-1.5">
                  <label class="block text-xs font-bold text-highlighted">API Base Endpoint</label>
                  <div class="h-10 px-3.5 rounded-xl border border-default/70 bg-elevated/40 flex items-center text-xs font-mono text-muted truncate">
                    {{ apiBase }}
                  </div>
                  <span class="text-[10px] text-muted block">
                    Set via <code class="font-mono">NUXT_PUBLIC_API_BASE</code> at build time.
                  </span>
                </div>

                <div class="space-y-1.5">
                  <label class="block text-xs font-bold text-highlighted">Latency</label>
                  <div class="h-10 px-3.5 rounded-xl border border-default/70 bg-elevated/40 flex items-center justify-between text-xs">
                    <span class="text-muted flex items-center gap-1.5 font-medium">
                      <UIcon name="i-lucide-activity" class="size-3.5 text-sky-500" />
                      /healthz
                    </span>
                    <span class="font-mono font-bold text-sm" :class="backendStatus === 'offline' ? 'text-rose-500' : 'text-primary-500'">
                      {{ backendLatency === null ? '—' : `${backendLatency} ms` }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Per-dependency health detail -->
              <div v-if="Object.keys(healthDetail).length > 0" class="flex flex-wrap gap-2 pt-1">
                <span
                  v-for="(value, key) in healthDetail"
                  :key="key"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg border text-[11px] font-mono"
                  :class="value === 'ok'
                    ? 'border-primary-500/25 bg-primary-500/[0.07] text-primary-400'
                    : 'border-amber-500/25 bg-amber-500/[0.07] text-amber-400'"
                >
                  <UIcon :name="value === 'ok' ? 'i-lucide-check' : 'i-lucide-triangle-alert'" class="size-3" />
                  {{ key }}: {{ value }}
                </span>
              </div>
            </div>

            <!-- Maintenance & Cache Actions (1 Col) -->
            <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-4 flex flex-col justify-between">
              <div>
                <div class="flex items-center gap-2.5 mb-1">
                  <UIcon name="i-lucide-wrench" class="size-5 text-rose-500" />
                  <h3 class="font-bold text-sm text-highlighted">Catalog & Cache</h3>
                </div>
                <p class="text-xs text-muted leading-relaxed">
                  Maintenance operations for multi-cloud metadata catalog and database indices.
                </p>
              </div>

              <div class="space-y-2.5">
                <UButton
                  label="Resync All Accounts"
                  icon="i-lucide-refresh-cw"
                  color="info"
                  variant="subtle"
                  block
                  class="rounded-xl text-xs font-bold justify-start"
                  :loading="isReindexing"
                  :disabled="accountsStore.accounts.length === 0"
                  @click="triggerReindex"
                />
                <p class="text-[10px] text-muted leading-relaxed">
                  Re-reads each account's quota and contents from the provider, then refreshes
                  the database index. Your virtual folder layout is left untouched.
                </p>
              </div>
            </div>
          </div>

          <!-- Row: Model Context Protocol (MCP) & AI Integration -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-6">
            <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4 pb-4 border-b border-default/60">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-violet-500/10 text-violet-400 border border-violet-500/20">
                  <UIcon name="i-lucide-bot" class="size-6" />
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <h3 class="font-bold text-sm text-highlighted">
                      Model Context Protocol (MCP) Server
                    </h3>
                    <UBadge
                      label="AI Agent Ready"
                      color="primary"
                      variant="subtle"
                      size="xs"
                      class="rounded-lg font-mono text-[10px] bg-violet-500/10 text-violet-400 border border-violet-500/20"
                    />
                  </div>
                  <p class="text-xs text-muted mt-0.5 leading-relaxed max-w-2xl">
                    Connect Claude Desktop, Cursor, Antigravity, or custom agents directly to Poly Cloud. AI models can search, read, and write files across all your connected clouds via 1 standard protocol.
                  </p>
                </div>
              </div>

              <div class="flex items-center gap-2 shrink-0">
                <div class="flex items-center gap-2 px-3 py-1 rounded-xl border border-emerald-500/30 bg-emerald-500/10 text-emerald-400 text-xs font-semibold">
                  <span class="size-2 rounded-full bg-emerald-400 animate-pulse" />
                  <span>JSON-RPC 2.0 Ready</span>
                </div>
              </div>
            </div>

            <!-- Client Config Tabs & Code Snippet -->
            <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
              <div class="lg:col-span-7 space-y-3">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-1.5 p-1 rounded-xl bg-elevated/40 border border-default/60 text-xs">
                    <button
                      type="button"
                      class="px-3 py-1.5 rounded-lg font-medium transition-colors cursor-pointer"
                      :class="mcpTab === 'claude' ? 'bg-primary-500 text-white font-semibold shadow-xs' : 'text-muted hover:text-highlighted'"
                      @click="mcpTab = 'claude'"
                    >
                      Claude Desktop
                    </button>
                    <button
                      type="button"
                      class="px-3 py-1.5 rounded-lg font-medium transition-colors cursor-pointer"
                      :class="mcpTab === 'cursor' ? 'bg-primary-500 text-white font-semibold shadow-xs' : 'text-muted hover:text-highlighted'"
                      @click="mcpTab = 'cursor'"
                    >
                      Cursor IDE
                    </button>
                    <button
                      type="button"
                      class="px-3 py-1.5 rounded-lg font-medium transition-colors cursor-pointer"
                      :class="mcpTab === 'http' ? 'bg-primary-500 text-white font-semibold shadow-xs' : 'text-muted hover:text-highlighted'"
                      @click="mcpTab = 'http'"
                    >
                      HTTP Endpoint
                    </button>
                  </div>

                  <UButton
                    :label="mcpCopied ? 'Copied' : 'Copy Config'"
                    :icon="mcpCopied ? 'i-lucide-check' : 'i-lucide-copy'"
                    color="neutral"
                    variant="outline"
                    size="xs"
                    class="rounded-xl font-semibold cursor-pointer"
                    @click="copyMcpConfig(mcpTab === 'claude' ? mcpClaudeConfig : mcpTab === 'cursor' ? mcpCursorEnv : mcpEndpointUrl)"
                  />
                </div>

                <!-- Claude Tab -->
                <div v-if="mcpTab === 'claude'" class="space-y-2">
                  <div class="p-4 rounded-2xl border border-default/70 bg-[#0d111a] font-mono text-[12px] leading-relaxed text-zinc-300 overflow-x-auto">
                    <pre><code>{{ mcpClaudeConfig }}</code></pre>
                  </div>
                  <p class="text-[11px] text-muted">
                    Paste this snippet into your <code class="font-mono text-highlighted">claude_desktop_config.json</code> under <code class="font-mono">mcpServers</code>.
                  </p>
                </div>

                <!-- Cursor Tab -->
                <div v-if="mcpTab === 'cursor'" class="space-y-3">
                  <div class="p-4 rounded-2xl border border-default/70 bg-[#0d111a] space-y-3">
                    <div>
                      <span class="text-[11px] text-muted block mb-1">Command</span>
                      <code class="block font-mono text-[12px] text-primary-400 bg-black/40 p-2 rounded-lg">go run ./apps/backend/cmd/mcp</code>
                    </div>
                    <div>
                      <span class="text-[11px] text-muted block mb-1">Environment</span>
                      <code class="block font-mono text-[12px] text-emerald-400 bg-black/40 p-2 rounded-lg">{{ mcpCursorEnv }}</code>
                    </div>
                  </div>
                  <p class="text-[11px] text-muted">
                    Add in Cursor Settings &rarr; Features &rarr; MCP &rarr; Add New MCP Server.
                  </p>
                </div>

                <!-- HTTP Tab -->
                <div v-if="mcpTab === 'http'" class="space-y-2">
                  <div class="p-4 rounded-2xl border border-default/70 bg-[#0d111a] space-y-2">
                    <div class="flex items-center gap-2">
                      <span class="px-2 py-0.5 rounded bg-primary-500/20 text-primary-400 font-mono text-[10px] font-bold">POST</span>
                      <code class="font-mono text-[12px] text-highlighted">{{ mcpEndpointUrl }}</code>
                    </div>
                    <p class="text-[11px] text-muted leading-relaxed">
                      Standard JSON-RPC 2.0 endpoint for remote agents or containerized AI assistants.
                    </p>
                  </div>
                </div>
              </div>

              <!-- Exposed MCP Tools List (5 Cols) -->
              <div class="lg:col-span-5 space-y-2.5">
                <span class="text-[11px] font-bold uppercase tracking-wider text-muted block">
                  8 Available AI Tools
                </span>
                <div class="grid grid-cols-1 gap-2">
                  <div
                    v-for="tool in mcpTools"
                    :key="tool.name"
                    class="p-2.5 rounded-xl border border-default/60 bg-elevated/20 flex items-start gap-2.5"
                  >
                    <div class="p-1.5 rounded-lg bg-primary-500/10 text-primary-400 border border-primary-500/20 shrink-0">
                      <UIcon :name="tool.icon" class="size-3.5" />
                    </div>
                    <div class="min-w-0">
                      <div class="font-mono font-bold text-xs text-highlighted truncate">{{ tool.name }}</div>
                      <p class="text-[11px] text-muted line-clamp-1 leading-snug">{{ tool.desc }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Row: Developer API Keys (Personal Access Tokens) -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-6">
            <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4 pb-4 border-b border-default/60">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
                  <UIcon name="i-lucide-key" class="size-6" />
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <h3 class="font-bold text-sm text-highlighted">
                      Developer API Keys
                    </h3>
                    <UBadge
                      label="REST & MCP Gateway"
                      color="warning"
                      variant="subtle"
                      size="xs"
                      class="rounded-lg font-mono text-[10px]"
                    />
                  </div>
                  <p class="text-xs text-muted mt-0.5 leading-relaxed max-w-2xl">
                    Generate personal access tokens to authenticate external scripts, CLI utilities, and autonomous AI agents using <code class="font-mono text-highlighted">Authorization: Bearer &lt;key&gt;</code>.
                  </p>
                </div>
              </div>

              <div class="flex items-center gap-2 shrink-0">
                <UButton
                  label="Generate New Key"
                  icon="i-lucide-plus"
                  color="primary"
                  class="rounded-xl text-xs font-bold"
                  @click="isCreateKeyOpen = true"
                />
              </div>
            </div>

            <!-- Active Keys Table / Empty State -->
            <div v-if="apiKeys.length === 0" class="p-8 rounded-2xl border border-dashed border-default/80 bg-elevated/10 text-center space-y-2">
              <div class="size-10 rounded-full bg-default/10 text-muted mx-auto flex items-center justify-center">
                <UIcon name="i-lucide-key-round" class="size-5" />
              </div>
              <p class="text-xs font-medium text-highlighted">No Developer API Keys</p>
              <p class="text-[11px] text-muted max-w-sm mx-auto">
                You haven't generated any API tokens yet. Create one to use Poly Cloud programmatically from your own tools.
              </p>
            </div>

            <div v-else class="space-y-3">
              <div
                v-for="k in apiKeys"
                :key="k.id"
                class="p-4 rounded-2xl border border-default/70 bg-elevated/20 flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:border-default transition-colors"
              >
                <div class="space-y-1.5 min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="font-bold text-xs text-highlighted truncate">{{ k.name }}</span>
                    <code class="font-mono text-[11px] text-primary-400 bg-black/40 px-2 py-0.5 rounded-lg border border-white/[0.04]">
                      {{ k.key_prefix }}
                    </code>
                    <div class="flex items-center gap-1">
                      <UBadge
                        v-for="s in k.scopes"
                        :key="s"
                        :label="s"
                        color="neutral"
                        variant="subtle"
                        size="xs"
                        class="rounded-md font-mono text-[9px] uppercase"
                      />
                    </div>
                  </div>
                  <div class="flex items-center gap-4 text-[11px] text-muted font-mono">
                    <span>Created: {{ formatDate(k.created_at) }}</span>
                    <span>Last used: <strong :class="k.last_used_at ? 'text-emerald-400' : 'text-muted'">{{ formatDate(k.last_used_at) }}</strong></span>
                  </div>
                </div>

                <div class="flex items-center gap-2 shrink-0">
                  <UButton
                    label="Revoke"
                    icon="i-lucide-trash-2"
                    color="error"
                    variant="ghost"
                    size="xs"
                    class="rounded-xl text-[11px]"
                    :loading="deletingKeyId === k.id"
                    @click="handleDeleteKey(k)"
                  />
                </div>
              </div>
            </div>

            <!-- Quick Example Snippet Box -->
            <div class="p-4 rounded-2xl border border-default/60 bg-black/40 space-y-2 font-mono text-xs">
              <div class="flex items-center justify-between text-muted text-[11px]">
                <span>Example: Accessing REST API with Bearer Token</span>
                <span class="text-primary-400">cURL</span>
              </div>
              <pre class="overflow-x-auto text-zinc-300 p-2 rounded-lg bg-black/50 border border-white/[0.04]"><code>curl http://localhost:8080/api/v1/files \
  -H "Authorization: Bearer plc_live_..."</code></pre>
            </div>
          </div>

          <!-- Backend runtime config: displayed, not editable -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm space-y-5">
            <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-indigo-500/10 text-indigo-500 border border-indigo-500/20">
                  <UIcon name="i-lucide-route" class="size-6" />
                </div>
                <div>
                  <h3 class="font-bold text-sm text-highlighted">Engine Configuration</h3>
                  <p class="text-xs text-muted mt-0.5 leading-relaxed max-w-xl">
                    The backend reads these values from the environment at startup. Changing them
                    means editing <code class="font-mono">.env</code> and restarting the service —
                    not from this page.
                  </p>
                </div>
              </div>
            </div>

            <div v-if="settings" class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Active routing strategy -->
              <div class="p-5 rounded-2xl border border-sky-500/30 bg-sky-500/[0.06] space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Smart Routing</span>
                  <code class="text-[10px] font-mono text-muted">ROUTING_STRATEGY</code>
                </div>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-check-circle-2" class="size-4 text-sky-500 shrink-0" />
                  <span class="font-bold text-sm text-highlighted">{{ activeStrategy.label }}</span>
                </div>
                <p class="text-[11px] text-muted leading-relaxed">{{ activeStrategy.desc }}</p>
              </div>

              <!-- Storage model -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Storage Model</span>
                  <UBadge :label="`Model ${settings.storage_model}`" color="primary" variant="subtle" size="xs" class="rounded-lg font-mono text-[10px]" />
                </div>
                <p class="text-[11px] text-muted leading-relaxed">
                  Each file is stored whole on a single account (whole-file).
                  <template v-if="!settings.chunking">
                    Splitting a file across accounts (Model B) is not available yet, so a file
                    larger than the roomiest account's free space is rejected rather than split.
                  </template>
                </p>
              </div>

              <!-- Physical folder -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-muted">Object Folder</span>
                  <code class="text-[10px] font-mono text-muted">RCLONE_BASE_DIR</code>
                </div>
                <code class="block font-mono text-sm text-primary-400">{{ settings.remote_base_dir }}/</code>
                <p class="text-[11px] text-muted leading-relaxed">
                  Every object lands flat in this folder on each account. The folder structure
                  you see lives in the database, not at the provider.
                </p>
              </div>

              <!-- Transfer -->
              <div class="p-5 rounded-2xl border border-default/70 bg-elevated/20 space-y-2">
                <span class="text-[11px] font-bold uppercase tracking-wider text-muted block">Transfer</span>
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-shield-check" class="size-4 text-primary-500 shrink-0" />
                  <span class="font-bold text-sm text-highlighted">Stream-through</span>
                </div>
                <p class="text-[11px] text-muted leading-relaxed">
                  Uploads and downloads stream directly between the browser and the provider;
                  the server keeps no copy of the file.
                </p>
              </div>
            </div>

            <div v-else class="p-5 rounded-2xl border border-amber-500/25 bg-amber-500/[0.06] text-xs text-muted">
              Backend configuration could not be read. Check the API connection.
            </div>
          </div>

          <!-- Capacity summary -->
          <div class="p-6 rounded-3xl border border-default/80 bg-card shadow-sm">
            <div class="flex flex-wrap items-center justify-between gap-4">
              <div class="flex items-start gap-3.5">
                <div class="flex size-11 shrink-0 items-center justify-center rounded-2xl bg-primary-500/10 text-primary-500 border border-primary-500/20">
                  <UIcon name="i-lucide-database" class="size-6" />
                </div>
                <div>
                  <h3 class="font-bold text-sm text-highlighted">Aggregate Capacity</h3>
                  <p class="text-xs text-muted mt-0.5">
                    {{ accountsStore.accounts.length }} accounts connected,
                    {{ accountsStore.activeAccounts.length }} active.
                  </p>
                </div>
              </div>

              <div class="flex items-center gap-6 font-mono text-xs">
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Used</span>
                  <span class="text-highlighted font-bold text-sm">{{ formatBytes(accountsStore.usedStorage) }}</span>
                </div>
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Free</span>
                  <span class="text-primary-500 font-bold text-sm">{{ formatBytes(accountsStore.freeStorage) }}</span>
                </div>
                <div>
                  <span class="text-muted block text-[10px] uppercase tracking-wider">Total</span>
                  <span class="text-highlighted font-bold text-sm">{{ formatBytes(accountsStore.totalStorage) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </UDashboardPanel>

    <!-- Modal: Generate API Key -->
    <UModal
      v-model:open="isCreateKeyOpen"
      :ui="{
        content: 'sm:max-w-md bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold flex items-center gap-2 text-white">
            <UIcon name="i-lucide-key" class="size-5 text-amber-400" />
            Generate Developer API Key
          </h3>
        </div>
      </template>

      <template #body>
        <div class="space-y-4">
          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-zinc-300">Key Name</label>
            <UInput
              v-model="newKeyName"
              placeholder="e.g. Cursor AI Assistant / Backup Script"
              class="w-full"
              autofocus
              @keydown.enter="handleCreateKey"
            />
          </div>

          <div class="space-y-2">
            <label class="text-xs font-semibold text-zinc-300 block">Permissions / Scopes</label>
            <div class="flex items-center gap-4 text-xs text-zinc-300">
              <label class="inline-flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  value="read"
                  checked
                  disabled
                  class="rounded bg-black/40 border-white/20 text-primary-500 focus:ring-0"
                >
                <span>Read (Search, List, Download)</span>
              </label>
              <label class="inline-flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  value="write"
                  checked
                  disabled
                  class="rounded bg-black/40 border-white/20 text-primary-500 focus:ring-0"
                >
                <span>Write (Upload, Delete, Folders)</span>
              </label>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-2 w-full">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            class="rounded-xl"
            :disabled="isSubmittingKey"
            @click="isCreateKeyOpen = false"
          />
          <UButton
            label="Generate Key"
            color="primary"
            class="rounded-xl font-bold"
            :loading="isSubmittingKey"
            :disabled="!newKeyName.trim()"
            @click="handleCreateKey"
          />
        </div>
      </template>
    </UModal>

    <!-- Modal: Reveal Generated Secret Key -->
    <UModal
      v-model:open="isSecretRevealOpen"
      :ui="{
        content: 'sm:max-w-lg bg-[#0d111a] border border-white/[0.09] rounded-3xl shadow-2xl p-0 overflow-hidden text-zinc-200',
        body: 'p-6 space-y-4',
        footer: 'px-6 py-4 bg-[#0b0e14] border-t border-white/[0.06]'
      }"
    >
      <template #header>
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-base font-bold flex items-center gap-2 text-emerald-400">
            <UIcon name="i-lucide-shield-check" class="size-5" />
            API Key Created Successfully
          </h3>
        </div>
      </template>

      <template #body>
        <div class="space-y-4">
          <div class="p-3.5 rounded-2xl bg-amber-500/10 border border-amber-500/20 text-amber-300 text-xs flex items-start gap-2.5 leading-relaxed">
            <UIcon name="i-lucide-alert-triangle" class="size-4 shrink-0 mt-0.5" />
            <span>
              <strong>Copy this token immediately!</strong> For your security, this raw key will never be shown again once you close this dialog.
            </span>
          </div>

          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-zinc-400">Secret Token</label>
            <div class="flex items-center gap-2">
              <UInput
                :model-value="createdSecretKey?.key"
                readonly
                class="font-mono text-xs flex-1 text-primary-400"
              />
              <UButton
                :label="copiedSecret ? 'Copied' : 'Copy'"
                :icon="copiedSecret ? 'i-lucide-check' : 'i-lucide-copy'"
                color="primary"
                class="rounded-xl shrink-0 font-semibold"
                @click="copySecretKey"
              />
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end w-full">
          <UButton
            label="I Have Copied My Key"
            color="primary"
            class="rounded-xl font-bold"
            @click="isSecretRevealOpen = false"
          />
        </div>
      </template>
    </UModal>
  </div>
</template>
