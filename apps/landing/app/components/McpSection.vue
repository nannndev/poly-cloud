<script setup lang="ts">
import { ref } from 'vue'

const activeTab = ref<'claude' | 'cursor' | 'antigravity' | 'curl'>('claude')
const copied = ref(false)

const claudeConfig = `{
  "mcpServers": {
    "poly-cloud": {
      "command": "go",
      "args": ["run", "./apps/backend/cmd/mcp"],
      "env": {
        "POLYCLOUD_API_URL": "http://localhost:8080"
      }
    }
  }
}`

const cursorConfig = `# Add in Cursor Settings → Features → MCP → Add New MCP Server
Name: poly-cloud
Type: command
Command: go run ./apps/backend/cmd/mcp
Environment: POLYCLOUD_API_URL=http://localhost:8080`

const antigravityConfig = `# Built-in Antigravity Skill (.agents/skills/poly-cloud-mcp)
# Simply ask in chat:
"Antigravity, search my connected cloud storage for 'contract-2025' and summarize it"`

const curlSnippet = `curl -X POST http://localhost:8080/api/v1/mcp \\
  -H "Content-Type: application/json" \\
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "search_files",
      "arguments": { "query": "quarterly-report" }
    }
  }'`

async function copyCode(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // fallback
  }
}

const skills = [
  {
    name: 'search_files',
    category: 'Search',
    badgeClass: 'text-sky-400 bg-sky-500/10 border-sky-500/20',
    desc: 'Instant metadata & full-text search across all combined cloud providers via PostgreSQL index.',
    example: 'search_files({ query: "invoice-2026.pdf" })'
  },
  {
    name: 'list_files',
    category: 'Explore',
    badgeClass: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20',
    desc: 'Browse unified virtual directories and inspect file hierarchies regardless of which drive holds them.',
    example: 'list_files({ folder_id: "root", limit: 20 })'
  },
  {
    name: 'read_file',
    category: 'Context Stream',
    badgeClass: 'text-purple-400 bg-purple-500/10 border-purple-500/20',
    desc: 'Stream file content directly into model context window on-demand without local caching.',
    example: 'read_file({ file_id: "fl_98a72b" })'
  },
  {
    name: 'upload_file',
    category: 'Autonomous Write',
    badgeClass: 'text-amber-400 bg-amber-500/10 border-amber-500/20',
    desc: 'AI uploads files with automatic smart routing to whichever connected drive has the most free space.',
    example: 'upload_file({ filename: "notes.md", content: "..." })'
  },
  {
    name: 'get_storage_quota',
    category: 'Analytics',
    badgeClass: 'text-indigo-400 bg-indigo-500/10 border-indigo-500/20',
    desc: 'Inspect real-time aggregate capacity, used/free bytes, and per-provider quotas across clouds.',
    example: 'get_storage_quota()'
  },
  {
    name: 'list_accounts',
    category: 'Telemetry',
    badgeClass: 'text-zinc-400 bg-zinc-500/10 border-zinc-500/20',
    desc: 'List active cloud remotes (Google Drive, OneDrive, S3, Dropbox) and their connectivity status.',
    example: 'list_accounts()'
  },
  {
    name: 'create_folder',
    category: 'Structure',
    badgeClass: 'text-teal-400 bg-teal-500/10 border-teal-500/20',
    desc: 'Create virtual folders in the database namespace without modifying remote provider roots.',
    example: 'create_folder({ name: "Q3 Backups" })'
  },
  {
    name: 'delete_file',
    category: 'Prune',
    badgeClass: 'text-rose-400 bg-rose-500/10 border-rose-500/20',
    desc: 'Safely purge obsolete files from cloud storage with automatic index cleanup.',
    example: 'delete_file({ file_id: "fl_123abc" })'
  }
]
</script>

<template>
  <section id="mcp" class="border-b border-white/[0.06] scroll-mt-16 bg-gradient-to-b from-ink-950 via-ink-900/40 to-ink-950">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-28">
      
      <!-- Section Header -->
      <div class="max-w-3xl">
        <div class="inline-flex items-center gap-2 rounded-full border border-accent-500/30 bg-accent-500/10 px-3 py-1 text-xs font-semibold text-accent-400">
          <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="3" y="11" width="18" height="10" rx="2" />
            <circle cx="12" cy="5" r="2" />
            <path d="M12 7v4" />
            <line x1="8" y1="16" x2="8" y2="16.01" stroke-width="2.5" />
            <line x1="16" y1="16" x2="16" y2="16.01" stroke-width="2.5" />
          </svg>
          Model Context Protocol (MCP) & AI Skills
        </div>

        <h2 class="mt-4 text-2xl font-semibold tracking-tight text-white sm:text-3xl md:text-4xl">
          Give AI Assistants Direct Eyes & Hands on All Your Cloud Drives
        </h2>
        <p class="mt-4 text-base leading-relaxed text-zinc-400">
          Poly Cloud includes an official implementation of Anthropic's <strong class="text-white">Model Context Protocol (MCP)</strong>. 
          Connect Claude Desktop, Cursor, Antigravity, or custom LLM agents directly to search, read, upload, and organize files 
          across Google Drive, OneDrive, S3, and Dropbox simultaneously.
        </p>
      </div>

      <!-- Live Agent Simulation Preview -->
      <div class="mt-12 rounded-2xl border border-white/[0.09] bg-ink-850/80 p-5 sm:p-7 backdrop-blur-md">
        <div class="flex flex-wrap items-center justify-between gap-4 border-b border-white/[0.06] pb-4">
          <div class="flex items-center gap-2.5">
            <div class="flex size-7 items-center justify-center rounded-lg bg-accent-500/15 text-accent-400 border border-accent-500/30">
              <svg viewBox="0 0 24 24" class="size-4" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
              </svg>
            </div>
            <span class="text-sm font-semibold text-white">How Claude or Cursor Interacts With Poly Cloud</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="inline-flex items-center gap-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/25 px-2.5 py-0.5 text-[11px] font-medium text-emerald-400">
              <span class="size-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
              MCP JSON-RPC 2.0 Live
            </span>
          </div>
        </div>

        <div class="mt-5 space-y-3.5 text-xs font-mono">
          <!-- User prompt -->
          <div class="flex items-start gap-3 rounded-xl bg-ink-800/70 p-3.5 border border-white/[0.05]">
            <span class="shrink-0 rounded bg-white/10 px-1.5 py-0.5 text-[10px] font-semibold text-zinc-300">USER</span>
            <span class="text-zinc-200">"Cari file laporan keuangan 'Q4-financials.pdf' dan berapa sisa kuota storage saya?"</span>
          </div>

          <!-- MCP Autonomous Calls -->
          <div class="space-y-2 pl-4 border-l-2 border-accent-500/40">
            <div class="flex items-center gap-2 text-accent-400">
              <svg viewBox="0 0 24 24" class="size-3.5 animate-spin" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" stroke-opacity="0.25" />
                <path d="M12 2a10 10 0 0 1 10 10" />
              </svg>
              <span>Calling Tool: <code class="font-bold text-white">poly-cloud:search_files({ "query": "Q4-financials" })</code></span>
            </div>
            <div class="rounded-lg bg-black/40 p-2.5 text-zinc-400 border border-white/[0.04]">
              → Found: <span class="text-emerald-400">"Q4-financials.pdf"</span> (4.2 MB) on <span class="text-sky-400">onedrive-office</span> [ID: fl_8912ba]
            </div>

            <div class="flex items-center gap-2 text-accent-400 pt-1">
              <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span>Calling Tool: <code class="font-bold text-white">poly-cloud:get_storage_quota()</code></span>
            </div>
            <div class="rounded-lg bg-black/40 p-2.5 text-zinc-400 border border-white/[0.04]">
              → Aggregate: <span class="text-zinc-200">46.5 GB free</span> across 4 cloud remotes (Google Drive, OneDrive, S3, Dropbox).
            </div>
          </div>

          <!-- AI Model Answer -->
          <div class="flex items-start gap-3 rounded-xl bg-accent-500/[0.08] p-3.5 border border-accent-500/20">
            <span class="shrink-0 rounded bg-accent-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-accent-300">CLAUDE</span>
            <span class="text-zinc-200 leading-relaxed font-sans text-xs">
              File <strong class="text-white">Q4-financials.pdf</strong> ditemukan tersimpan di OneDrive. Total storage gabungan Anda masih memiliki <strong>46.5 GB free</strong> dari total 75 GB di 4 akun cloud.
            </span>
          </div>
        </div>
      </div>

      <!-- 8 MCP Skills Grid -->
      <div class="mt-16">
        <div class="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-white sm:text-xl">
              Complete 8-Tool AI Skill Set
            </h3>
            <p class="mt-1 text-sm text-zinc-400">
              Autonomous functions exposed to LLMs with strict parameter validation and zero vendor lock-in.
            </p>
          </div>
          <a
            href="https://github.com/nannndev/poly-cloud/blob/main/.agents/skills/poly-cloud-mcp/SKILL.md"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center gap-1.5 text-xs font-semibold text-accent-400 hover:text-accent-300 transition-colors"
          >
            <span>View SKILL.md specification</span>
            <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M5 12h14M12 5l7 7-7 7" />
            </svg>
          </a>
        </div>

        <div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <article
            v-for="skill in skills"
            :key="skill.name"
            class="reveal rounded-xl border border-white/[0.08] bg-ink-850 p-4 transition-all hover:border-accent-500/40 hover:bg-ink-800"
          >
            <div class="flex items-center justify-between gap-2">
              <code class="text-xs font-bold font-mono text-white">{{ skill.name }}</code>
              <span class="rounded-md border px-1.5 py-0.5 text-[10px] font-semibold font-mono" :class="skill.badgeClass">
                {{ skill.category }}
              </span>
            </div>
            <p class="mt-2.5 text-xs leading-relaxed text-zinc-400">
              {{ skill.desc }}
            </p>
            <div class="mt-3 rounded bg-black/40 px-2 py-1 font-mono text-[11px] text-zinc-500 truncate border border-white/[0.03]">
              {{ skill.example }}
            </div>
          </article>
        </div>
      </div>

      <!-- Quick Integration Snippets -->
      <div class="mt-16 rounded-2xl border border-white/[0.08] bg-ink-850 p-6 sm:p-8">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-white/[0.06] pb-5">
          <div>
            <h3 class="text-base font-semibold text-white">1-Click Integration With Your AI Assistant</h3>
            <p class="mt-1 text-xs text-zinc-400">Copy configuration into your client and start querying your multi-cloud immediately.</p>
          </div>

          <!-- Tabs -->
          <div class="flex flex-wrap items-center gap-1 rounded-xl bg-ink-950 p-1 border border-white/[0.06]">
            <button
              type="button"
              class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
              :class="activeTab === 'claude' ? 'bg-accent-500 text-white font-semibold shadow-xs' : 'text-zinc-400 hover:text-white'"
              @click="activeTab = 'claude'"
            >
              Claude Desktop
            </button>
            <button
              type="button"
              class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
              :class="activeTab === 'cursor' ? 'bg-accent-500 text-white font-semibold shadow-xs' : 'text-zinc-400 hover:text-white'"
              @click="activeTab = 'cursor'"
            >
              Cursor AI
            </button>
            <button
              type="button"
              class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
              :class="activeTab === 'antigravity' ? 'bg-accent-500 text-white font-semibold shadow-xs' : 'text-zinc-400 hover:text-white'"
              @click="activeTab = 'antigravity'"
            >
              Antigravity IDE
            </button>
            <button
              type="button"
              class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
              :class="activeTab === 'curl' ? 'bg-accent-500 text-white font-semibold shadow-xs' : 'text-zinc-400 hover:text-white'"
              @click="activeTab = 'curl'"
            >
              HTTP / cURL
            </button>
          </div>
        </div>

        <div class="mt-6 relative">
          <div class="flex items-center justify-between pb-2">
            <span class="font-mono text-[11px] text-zinc-500">
              <template v-if="activeTab === 'claude'">~/Library/Application Support/Claude/claude_desktop_config.json</template>
              <template v-else-if="activeTab === 'cursor'">Cursor Settings → Features → MCP</template>
              <template v-else-if="activeTab === 'antigravity'">.agents/skills/poly-cloud-mcp/SKILL.md</template>
              <template v-else>POST /api/v1/mcp (JSON-RPC 2.0)</template>
            </span>

            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-white/[0.1] bg-ink-800 px-3 py-1 text-xs font-medium text-zinc-300 transition-colors hover:border-white/20 hover:text-white"
              @click="copyCode(
                activeTab === 'claude' ? claudeConfig :
                activeTab === 'cursor' ? cursorConfig :
                activeTab === 'antigravity' ? antigravityConfig : curlSnippet
              )"
            >
              <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2">
                <template v-if="copied">
                  <polyline points="20 6 9 17 4 12" />
                </template>
                <template v-else>
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </template>
              </svg>
              <span>{{ copied ? 'Copied!' : 'Copy' }}</span>
            </button>
          </div>

          <pre class="overflow-x-auto rounded-xl border border-white/[0.06] bg-ink-950 p-4 font-mono text-xs text-zinc-300"><code><template v-if="activeTab === 'claude'">{{ claudeConfig }}</template><template v-else-if="activeTab === 'cursor'">{{ cursorConfig }}</template><template v-else-if="activeTab === 'antigravity'">{{ antigravityConfig }}</template><template v-else>{{ curlSnippet }}</template></code></pre>
        </div>
      </div>

    </div>
  </section>
</template>
