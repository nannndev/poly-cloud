<script setup lang="ts">
import { REPO_URL } from '~/config/site'

type TabKey = 'docker' | 'providers' | 'production' | 'baremetal'

const activeTab = ref<TabKey>('docker')

const copiedKey = ref<string | null>(null)
let timer: ReturnType<typeof setTimeout> | undefined

async function copyCode(key: string, text: string) {
  try {
    await navigator.clipboard.writeText(text)
    copiedKey.value = key
    clearTimeout(timer)
    timer = setTimeout(() => { copiedKey.value = null }, 2000)
  } catch {
    // Clipboard API fallback
  }
}

onBeforeUnmount(() => clearTimeout(timer))

// Snippets
const dockerClone = `git clone ${REPO_URL}
cd poly-cloud
cp .env.example .env`

const dockerKeys = `# Generate 256-bit random key for token encryption (AES-256-GCM)
openssl rand -hex 32

# Generate random password to secure rclone daemon
openssl rand -hex 24`

const dockerUp = `# Build and start all 4 services (Frontend, Backend, PostgreSQL, rclone)
make dev

# Or run detached in background:
docker compose up -d --build`

const dockerOps = `# View real-time logs from all containers
make logs

# Stop services gracefully (data in Postgres volume remains intact)
make down

# Run backend unit tests inside Go container (no Go installation needed)
make test`

const caddySnippet = `# /etc/caddy/Caddyfile
cloud.yourdomain.com {
    reverse_proxy localhost:3000
}

api.cloud.yourdomain.com {
    reverse_proxy localhost:8080
}`

const envDomainSnippet = `# In .env:
FRONTEND_URL=https://cloud.yourdomain.com
BACKEND_URL=https://api.cloud.yourdomain.com`

const pgBackupSnippet = `# Backup your database to a local SQL file:
docker compose exec -T postgres pg_dump -U polycloud polycloud > backup_$(date +%F).sql

# Restore if needed:
cat backup.sql | docker compose exec -T postgres psql -U polycloud polycloud`

const baremetalDb = `# 1. Create database and apply schema
createuser -s polycloud
createdb -O polycloud polycloud
psql -d polycloud -c "ALTER USER polycloud PASSWORD 'polycloud';"
psql -d polycloud -f apps/backend/migrations/001_init.sql`

const baremetalRclone = `# 2. Launch rclone in RC (remote control) daemon mode
rclone rcd --rc-web-gui=false \\
  --rc-addr=127.0.0.1:5572 \\
  --rc-user=polycloud \\
  --rc-pass="$RCLONE_RC_PASS"`

const baremetalApp = `# 3. Run Go backend API
cd apps/backend && go run ./cmd/server

# 4. In another terminal, run Nuxt frontend
cd apps/frontend && npm install && npm run dev`
</script>

<template>
  <section id="install" class="border-b border-white/[0.06] scroll-mt-16 bg-ink-900/40">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <!-- Section Header -->
      <div class="max-w-3xl">
        <div class="inline-flex items-center gap-2 rounded-full border border-accent-400/25 bg-accent-500/10 px-3 py-1 text-[11px] font-semibold text-accent-300">
          <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="4 17 10 11 4 5" /><line x1="12" y1="19" x2="20" y2="19" />
          </svg>
          Self-Hosting Tutorial &amp; Setup
        </div>

        <h2 class="mt-4 text-2xl font-semibold tracking-tight text-white md:text-4xl">
          Run Poly Cloud on your own infrastructure
        </h2>
        <p class="mt-3 text-[15px] leading-relaxed text-zinc-400">
          Deploy on your home server, homelab, Raspberry Pi, or VPS in minutes.
          Zero SaaS vendor lock-in, zero monthly subscription fees, and 100% control over your credentials and files.
        </p>
      </div>

      <!-- Tab Navigation -->
      <div class="mt-10 flex flex-wrap gap-2 border-b border-white/[0.08] pb-4">
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-xl px-4 py-2.5 text-[13px] font-semibold transition-all cursor-pointer"
          :class="activeTab === 'docker' ? 'bg-accent-600 text-white shadow-xs' : 'bg-ink-850 text-zinc-400 border border-white/[0.06] hover:text-white hover:border-white/20'"
          @click="activeTab = 'docker'"
        >
          <svg viewBox="0 0 24 24" class="size-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
          </svg>
          1. Docker Compose (Quickstart)
        </button>

        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-xl px-4 py-2.5 text-[13px] font-semibold transition-all cursor-pointer"
          :class="activeTab === 'providers' ? 'bg-accent-600 text-white shadow-xs' : 'bg-ink-850 text-zinc-400 border border-white/[0.06] hover:text-white hover:border-white/20'"
          @click="activeTab = 'providers'"
        >
          <svg viewBox="0 0 24 24" class="size-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z" />
          </svg>
          2. Connecting Storage Providers
        </button>

        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-xl px-4 py-2.5 text-[13px] font-semibold transition-all cursor-pointer"
          :class="activeTab === 'production' ? 'bg-accent-600 text-white shadow-xs' : 'bg-ink-850 text-zinc-400 border border-white/[0.06] hover:text-white hover:border-white/20'"
          @click="activeTab = 'production'"
        >
          <svg viewBox="0 0 24 24" class="size-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
            <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
            <line x1="6" y1="6" x2="6.01" y2="6" />
            <line x1="6" y1="18" x2="6.01" y2="18" />
          </svg>
          3. Production &amp; Domain Setup
        </button>

        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-xl px-4 py-2.5 text-[13px] font-semibold transition-all cursor-pointer"
          :class="activeTab === 'baremetal' ? 'bg-accent-600 text-white shadow-xs' : 'bg-ink-850 text-zinc-400 border border-white/[0.06] hover:text-white hover:border-white/20'"
          @click="activeTab = 'baremetal'"
        >
          <svg viewBox="0 0 24 24" class="size-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="4 17 10 11 4 5" /><line x1="12" y1="19" x2="20" y2="19" />
          </svg>
          4. Bare-Metal (Without Docker)
        </button>
      </div>

      <!-- Tab 1: Docker Compose Walkthrough -->
      <div v-if="activeTab === 'docker'" class="mt-8 space-y-8">
        <div class="grid gap-6 lg:grid-cols-3">
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-5">
            <span class="font-mono text-xs text-accent-400">PREREQUISITE</span>
            <h4 class="mt-1 text-sm font-semibold text-white">Docker &amp; Compose</h4>
            <p class="mt-1 text-xs text-zinc-400">Docker Desktop on macOS/Windows, or Docker Engine + Docker Compose Plugin on Linux.</p>
          </div>
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-5">
            <span class="font-mono text-xs text-accent-400">SERVICES BUNDLED</span>
            <h4 class="mt-1 text-sm font-semibold text-white">4 Isolated Containers</h4>
            <p class="mt-1 text-xs text-zinc-400">Go 1.25 API, Nuxt 4 Frontend, PostgreSQL 16 DB, and rclone daemon over internal bridge network.</p>
          </div>
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-5">
            <span class="font-mono text-xs text-accent-400">LOCAL ENDPOINTS</span>
            <h4 class="mt-1 text-sm font-semibold text-white">Ports 3000 &amp; 8080</h4>
            <p class="mt-1 text-xs text-zinc-400">Web dashboard at <code class="font-mono text-zinc-300">localhost:3000</code> and REST API at <code class="font-mono text-zinc-300">localhost:8080</code>.</p>
          </div>
        </div>

        <!-- Step by Step -->
        <div class="space-y-6">
          <!-- Step 1 -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 overflow-hidden">
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3 bg-ink-800/40">
              <div class="flex items-center gap-3">
                <span class="grid size-6 place-items-center rounded-full bg-accent-500/20 text-accent-300 font-mono text-xs font-bold">1</span>
                <span class="text-sm font-semibold text-white">Clone repository &amp; create .env file</span>
              </div>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-[11px] font-medium text-zinc-400 transition-colors hover:bg-white/[0.06] hover:text-white"
                @click="copyCode('clone', dockerClone)"
              >
                {{ copiedKey === 'clone' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <pre class="overflow-x-auto p-5 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ dockerClone }}</code></pre>
          </div>

          <!-- Step 2 -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 overflow-hidden">
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3 bg-ink-800/40">
              <div class="flex items-center gap-3">
                <span class="grid size-6 place-items-center rounded-full bg-accent-500/20 text-accent-300 font-mono text-xs font-bold">2</span>
                <span class="text-sm font-semibold text-white">Generate cryptographic security keys</span>
              </div>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-[11px] font-medium text-zinc-400 transition-colors hover:bg-white/[0.06] hover:text-white"
                @click="copyCode('keys', dockerKeys)"
              >
                {{ copiedKey === 'keys' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <div class="p-5 space-y-3">
              <p class="text-xs leading-relaxed text-zinc-400">
                Run the commands below and paste the outputs into your <code class="font-mono text-accent-300">.env</code> file.
                <strong class="text-zinc-200">TOKEN_ENC_KEY</strong> encrypts your cloud OAuth tokens at rest using AES-256-GCM.
                <strong class="text-zinc-200">RCLONE_RC_PASS</strong> protects the internal rclone storage engine.
              </p>
              <pre class="overflow-x-auto rounded-xl bg-ink-950 p-4 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ dockerKeys }}</code></pre>
              <div class="rounded-xl border border-amber-500/20 bg-amber-500/[0.06] p-3 text-[12px] text-amber-200/90">
                ⚠️ <strong>Important:</strong> Keep your <code class="font-mono">TOKEN_ENC_KEY</code> safe. Changing it after connecting accounts will make existing saved tokens unreadable.
              </div>
            </div>
          </div>

          <!-- Step 3 -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 overflow-hidden">
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3 bg-ink-800/40">
              <div class="flex items-center gap-3">
                <span class="grid size-6 place-items-center rounded-full bg-accent-500/20 text-accent-300 font-mono text-xs font-bold">3</span>
                <span class="text-sm font-semibold text-white">Build &amp; spin up all services</span>
              </div>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-[11px] font-medium text-zinc-400 transition-colors hover:bg-white/[0.06] hover:text-white"
                @click="copyCode('up', dockerUp)"
              >
                {{ copiedKey === 'up' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <div class="p-5 space-y-3">
              <pre class="overflow-x-auto rounded-xl bg-ink-950 p-4 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ dockerUp }}</code></pre>
              <div class="grid gap-3 sm:grid-cols-2 pt-1 text-xs">
                <div class="flex items-center justify-between rounded-xl border border-white/[0.06] bg-ink-900 p-3">
                  <span class="text-zinc-400">Dashboard UI</span>
                  <a href="http://localhost:3000" target="_blank" rel="noopener" class="font-mono text-accent-400 hover:underline">http://localhost:3000 ↗</a>
                </div>
                <div class="flex items-center justify-between rounded-xl border border-white/[0.06] bg-ink-900 p-3">
                  <span class="text-zinc-400">API Healthcheck</span>
                  <a href="http://localhost:8080/healthz" target="_blank" rel="noopener" class="font-mono text-accent-400 hover:underline">http://localhost:8080/healthz ↗</a>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 4: Maintenance -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 overflow-hidden">
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3 bg-ink-800/40">
              <div class="flex items-center gap-3">
                <span class="grid size-6 place-items-center rounded-full bg-accent-500/20 text-accent-300 font-mono text-xs font-bold">4</span>
                <span class="text-sm font-semibold text-white">Daily management &amp; tests</span>
              </div>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1 text-[11px] font-medium text-zinc-400 transition-colors hover:bg-white/[0.06] hover:text-white"
                @click="copyCode('ops', dockerOps)"
              >
                {{ copiedKey === 'ops' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <pre class="overflow-x-auto p-5 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ dockerOps }}</code></pre>
          </div>
        </div>
      </div>

      <!-- Tab 2: Providers Setup -->
      <div v-if="activeTab === 'providers'" class="mt-8 space-y-6">
        <p class="text-sm text-zinc-400 leading-relaxed max-w-3xl">
          Poly Cloud aggregates multiple providers simultaneously. You can mix free tier accounts from Google Drive, OneDrive, and Dropbox, or connect high-capacity S3 / R2 buckets.
        </p>

        <div class="grid gap-6 md:grid-cols-2">
          <!-- S3 Compatible -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-6 space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-white">S3-Compatible (R2, MinIO, AWS, B2)</h3>
              <span class="rounded-lg bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 text-[11px] font-semibold text-emerald-400">Zero .env Setup</span>
            </div>
            <p class="text-xs text-zinc-400 leading-relaxed">
              Works straight out of the box without restarting the server or adding API keys in <code class="font-mono text-zinc-300">.env</code>.
            </p>
            <ol class="space-y-2.5 text-xs text-zinc-300 list-decimal list-inside leading-relaxed">
              <li>Open Poly Cloud Web UI at <code class="font-mono text-accent-300">localhost:3000</code>.</li>
              <li>Navigate to <span class="font-medium text-white">Connected Accounts</span> → <span class="font-medium text-white">Add Account</span>.</li>
              <li>Select <span class="font-semibold text-white">S3 Compatible</span>.</li>
              <li>Enter your Bucket Name, Endpoint URL (e.g. Cloudflare R2 or MinIO), Access Key, and Secret Key.</li>
              <li>Poly Cloud verifies connection with rclone and mounts the bucket instantly!</li>
            </ol>
          </div>

          <!-- Google Drive -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-6 space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-white">Google Drive</h3>
              <span class="rounded-lg bg-accent-500/10 border border-accent-400/20 px-2 py-0.5 text-[11px] font-semibold text-accent-300">OAuth 2.0</span>
            </div>
            <p class="text-xs text-zinc-400 leading-relaxed">
              Use your own Google Cloud OAuth app credentials to stay independent.
            </p>
            <ol class="space-y-2.5 text-xs text-zinc-300 list-decimal list-inside leading-relaxed">
              <li>Go to <a href="https://console.cloud.google.com" target="_blank" rel="noopener" class="text-accent-400 hover:underline">Google Cloud Console</a> &amp; enable the <strong class="text-white">Google Drive API</strong>.</li>
              <li>Create credentials → <strong class="text-white">OAuth client ID</strong> (Web application).</li>
              <li>Add Authorized redirect URI:
                <div class="mt-1 font-mono text-[11px] bg-ink-950 p-2 rounded-lg text-accent-300 break-all select-all">
                  http://localhost:8080/api/v1/auth/callback/gdrive
                </div>
              </li>
              <li>Put <code class="font-mono text-zinc-300">GDRIVE_CLIENT_ID</code> and <code class="font-mono text-zinc-300">GDRIVE_CLIENT_SECRET</code> in <code class="font-mono text-zinc-300">.env</code>.</li>
              <li>Restart backend and connect through the web dashboard consent screen.</li>
            </ol>
          </div>

          <!-- OneDrive -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-6 space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-white">Microsoft OneDrive</h3>
              <span class="rounded-lg bg-accent-500/10 border border-accent-400/20 px-2 py-0.5 text-[11px] font-semibold text-accent-300">OAuth 2.0</span>
            </div>
            <p class="text-xs text-zinc-400 leading-relaxed">
              Connect personal OneDrive or work/school Microsoft accounts.
            </p>
            <ol class="space-y-2.5 text-xs text-zinc-300 list-decimal list-inside leading-relaxed">
              <li>Register an app in <a href="https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps" target="_blank" rel="noopener" class="text-accent-400 hover:underline">Microsoft Entra ID / Azure Portal</a>.</li>
              <li>Add Web platform redirect URI:
                <div class="mt-1 font-mono text-[11px] bg-ink-950 p-2 rounded-lg text-accent-300 break-all select-all">
                  http://localhost:8080/api/v1/auth/callback/onedrive
                </div>
              </li>
              <li>Add <code class="font-mono text-zinc-300">ONEDRIVE_CLIENT_ID</code> and <code class="font-mono text-zinc-300">ONEDRIVE_CLIENT_SECRET</code> to your <code class="font-mono text-zinc-300">.env</code>.</li>
              <li>Click Connect OneDrive in your Poly Cloud UI.</li>
            </ol>
          </div>

          <!-- Dropbox -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-6 space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-base font-semibold text-white">Dropbox</h3>
              <span class="rounded-lg bg-accent-500/10 border border-accent-400/20 px-2 py-0.5 text-[11px] font-semibold text-accent-300">OAuth 2.0</span>
            </div>
            <p class="text-xs text-zinc-400 leading-relaxed">
              Connect standard Dropbox accounts with full file read/write permissions.
            </p>
            <ol class="space-y-2.5 text-xs text-zinc-300 list-decimal list-inside leading-relaxed">
              <li>Create an app in the <a href="https://www.dropbox.com/developers/apps" target="_blank" rel="noopener" class="text-accent-400 hover:underline">Dropbox App Console</a> (Scoped access, Full Dropbox).</li>
              <li>Add Redirect URI:
                <div class="mt-1 font-mono text-[11px] bg-ink-950 p-2 rounded-lg text-accent-300 break-all select-all">
                  http://localhost:8080/api/v1/auth/callback/dropbox
                </div>
              </li>
              <li>Add permissions: <code class="font-mono text-zinc-300">files.content.write</code>, <code class="font-mono text-zinc-300">files.content.read</code>.</li>
              <li>Set <code class="font-mono text-zinc-300">DROPBOX_CLIENT_ID</code> and <code class="font-mono text-zinc-300">DROPBOX_CLIENT_SECRET</code> in <code class="font-mono text-zinc-300">.env</code>.</li>
            </ol>
          </div>
        </div>
      </div>

      <!-- Tab 3: Production & Reverse Proxy -->
      <div v-if="activeTab === 'production'" class="mt-8 space-y-6">
        <div class="grid gap-6 lg:grid-cols-[1.1fr_1fr]">
          <div class="space-y-6">
            <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-6 space-y-3">
              <h3 class="text-base font-semibold text-white">Exposing Poly Cloud to Your Domain</h3>
              <p class="text-xs leading-relaxed text-zinc-400">
                To access your self-hosted Poly Cloud instance securely outside your local LAN, place it behind a modern reverse proxy like <strong class="text-zinc-200">Caddy</strong> or <strong class="text-zinc-200">Nginx</strong> with automatic Let's Encrypt SSL.
              </p>
              <div class="flex items-center justify-between pt-2">
                <span class="font-mono text-xs text-zinc-500">Caddyfile example</span>
                <button
                  type="button"
                  class="text-[11px] font-medium text-accent-400 hover:underline"
                  @click="copyCode('caddy', caddySnippet)"
                >
                  {{ copiedKey === 'caddy' ? '✓ Copied' : 'Copy Caddyfile' }}
                </button>
              </div>
              <pre class="overflow-x-auto rounded-xl bg-ink-950 p-4 font-mono text-[12px] leading-relaxed text-zinc-300"><code>{{ caddySnippet }}</code></pre>
            </div>

            <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-6 space-y-3">
              <h3 class="text-base font-semibold text-white">Updating URLs in .env</h3>
              <p class="text-xs leading-relaxed text-zinc-400">
                When using custom domains, update the origins in <code class="font-mono text-zinc-300">.env</code> so OAuth redirects and CORS headers match your public domain:
              </p>
              <pre class="overflow-x-auto rounded-xl bg-ink-950 p-4 font-mono text-[12px] leading-relaxed text-zinc-300"><code>{{ envDomainSnippet }}</code></pre>
            </div>
          </div>

          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-6 space-y-4">
            <h3 class="text-base font-semibold text-white">Backup &amp; Data Persistence</h3>
            <p class="text-xs leading-relaxed text-zinc-400">
              Poly Cloud stores metadata and virtual folder trees in PostgreSQL. Actual files live safely on your cloud providers.
              A single Postgres dump captures your entire catalog and connected account configurations:
            </p>
            <div class="flex items-center justify-between pt-1">
              <span class="font-mono text-xs text-zinc-500">PostgreSQL Dump</span>
              <button
                type="button"
                class="text-[11px] font-medium text-accent-400 hover:underline"
                @click="copyCode('backup', pgBackupSnippet)"
              >
                {{ copiedKey === 'backup' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <pre class="overflow-x-auto rounded-xl bg-ink-950 p-4 font-mono text-[12px] leading-relaxed text-zinc-300"><code>{{ pgBackupSnippet }}</code></pre>

            <div class="rounded-xl border border-white/[0.06] bg-ink-900 p-4 text-xs text-zinc-400 space-y-2">
              <div class="font-semibold text-white">💡 Homelab Tip: Zero File Loss Guarantee</div>
              <p>
                Because files are uploaded whole directly into provider remotes, even if your Poly Cloud server burns down, all original files remain untouched and accessible directly in their respective Google Drive, S3, or Dropbox apps!
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab 4: Bare-metal Setup -->
      <div v-if="activeTab === 'baremetal'" class="mt-8 space-y-6">
        <p class="text-sm text-zinc-400 leading-relaxed max-w-3xl">
          Prefer running natively without container overhead? You can run Poly Cloud on any Linux, BSD, or macOS machine with native systemd services.
          Requires <strong class="text-zinc-200">Go 1.25+</strong>, <strong class="text-zinc-200">Node.js 20+</strong>, <strong class="text-zinc-200">PostgreSQL 16</strong>, and <strong class="text-zinc-200">rclone</strong>.
        </p>

        <div class="space-y-5">
          <!-- DB Init -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 overflow-hidden">
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3 bg-ink-800/40">
              <span class="text-sm font-semibold text-white">1. PostgreSQL Database &amp; Migrations</span>
              <button
                type="button"
                class="text-[11px] font-medium text-zinc-400 hover:text-white"
                @click="copyCode('bm-db', baremetalDb)"
              >
                {{ copiedKey === 'bm-db' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <pre class="overflow-x-auto p-5 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ baremetalDb }}</code></pre>
          </div>

          <!-- Rclone Daemon -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 overflow-hidden">
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3 bg-ink-800/40">
              <span class="text-sm font-semibold text-white">2. Launch rclone Remote Control Daemon</span>
              <button
                type="button"
                class="text-[11px] font-medium text-zinc-400 hover:text-white"
                @click="copyCode('bm-rc', baremetalRclone)"
              >
                {{ copiedKey === 'bm-rc' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <pre class="overflow-x-auto p-5 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ baremetalRclone }}</code></pre>
          </div>

          <!-- Backend & Frontend -->
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 overflow-hidden">
            <div class="flex items-center justify-between border-b border-white/[0.06] px-5 py-3 bg-ink-800/40">
              <span class="text-sm font-semibold text-white">3. Run Backend API &amp; Frontend Interface</span>
              <button
                type="button"
                class="text-[11px] font-medium text-zinc-400 hover:text-white"
                @click="copyCode('bm-app', baremetalApp)"
              >
                {{ copiedKey === 'bm-app' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <pre class="overflow-x-auto p-5 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ baremetalApp }}</code></pre>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
