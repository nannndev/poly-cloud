<script setup lang="ts">
import { GITHUB } from '~/config/site'

const columns = [
  {
    title: 'Product',
    links: [
      { label: 'How it works', to: '/#how-it-works', external: false },
      { label: 'AI & MCP Protocol', to: '/#mcp', external: false },
      { label: 'Self-Host Tutorial', to: '/#install', external: false },
      { label: 'Supported Providers', to: '/#providers', external: false },
      { label: 'Zero-Knowledge Architecture', to: '/#architecture', external: false }
    ]
  },
  {
    title: 'Open Source',
    links: [
      { label: 'FOSS Manifesto', to: '/#open-source', external: false },
      { label: 'Contributors', to: '/contributors', external: false },
      { label: 'Support & Sponsor', to: '/support', external: false },
      { label: 'Releases', to: GITHUB.releases, external: true },
      { label: 'MIT License', to: GITHUB.license, external: true }
    ]
  },
  {
    title: 'Community',
    links: [
      { label: 'GitHub Repository', to: GITHUB.repo, external: true },
      { label: 'Issues & Bug Reports', to: GITHUB.issues, external: true },
      { label: 'Discussions & Help', to: GITHUB.discussions, external: true },
      { label: 'Contributing Guide', to: GITHUB.contributing, external: true }
    ]
  }
]

const year = new Date().getFullYear()
</script>

<template>
  <footer class="border-t border-white/[0.06] bg-ink-950">
    <div class="mx-auto max-w-6xl px-4 py-14 sm:px-6">
      <div class="grid gap-10 lg:grid-cols-[1.4fr_2fr]">
        <div>
          <NuxtLink to="/" class="flex items-center gap-2.5">
            <span class="grid size-7 place-items-center rounded-lg border border-accent-400/25 bg-[#1E2430]">
              <PolyMark class="size-4" />
            </span>
            <span class="text-[13px] font-medium text-zinc-300">Poly Cloud</span>
          </NuxtLink>

          <p class="mt-4 max-w-[42ch] text-[13px] leading-relaxed text-zinc-500">
            Self-hosted storage aggregator. Your cloud accounts, one drive, nothing
            copied onto anyone else's server.
          </p>

          <div class="mt-5 flex items-center gap-2">
            <a
              :href="GITHUB.repo"
              target="_blank"
              rel="noopener"
              aria-label="Poly Cloud on GitHub"
              class="grid size-9 place-items-center rounded-xl border border-white/[0.08] bg-ink-850 text-zinc-400 transition-colors hover:border-white/20 hover:text-white"
            >
              <svg viewBox="0 0 24 24" class="size-4" fill="currentColor" aria-hidden="true">
                <path d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.09.68-.22.68-.49l-.01-1.72c-2.78.62-3.37-1.37-3.37-1.37-.46-1.18-1.11-1.5-1.11-1.5-.91-.64.07-.62.07-.62 1 .07 1.53 1.06 1.53 1.06.89 1.57 2.34 1.12 2.91.85.09-.66.35-1.12.63-1.38-2.22-.26-4.56-1.14-4.56-5.07 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05A9.3 9.3 0 0 1 12 6.84c.85 0 1.71.12 2.51.34 1.91-1.33 2.75-1.05 2.75-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.94-2.34 4.81-4.57 5.06.36.32.68.94.68 1.9l-.01 2.818c0 .27.18.59.69.49A10.26 10.26 0 0 0 22 12.25C22 6.58 17.52 2 12 2Z" />
              </svg>
            </a>
            <NuxtLink
              to="/support"
              class="inline-flex items-center gap-2 rounded-xl border border-white/[0.08] bg-ink-850 px-3.5 py-2 text-[13px] font-medium text-zinc-300 transition-colors hover:border-accent-500/30 hover:text-white"
            >
              <svg viewBox="0 0 24 24" class="size-4 text-accent-400" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M12 20s-7-4.35-7-9a3.8 3.8 0 0 1 7-2.1A3.8 3.8 0 0 1 19 11c0 4.65-7 9-7 9Z" />
              </svg>
              Sponsor
            </NuxtLink>
          </div>
        </div>

        <div class="grid gap-8 sm:grid-cols-3">
          <div v-for="col in columns" :key="col.title">
            <p class="text-[11px] font-semibold uppercase tracking-[0.14em] text-zinc-500">
              {{ col.title }}
            </p>
            <ul class="mt-4 space-y-2.5">
              <li v-for="link in col.links" :key="link.label">
                <a
                  v-if="link.external"
                  :href="link.to"
                  target="_blank"
                  rel="noopener"
                  class="text-[13px] text-zinc-400 transition-colors hover:text-white"
                >{{ link.label }}</a>
                <NuxtLink
                  v-else
                  :to="link.to"
                  class="text-[13px] text-zinc-400 transition-colors hover:text-white"
                >{{ link.label }}</NuxtLink>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <div class="mt-12 flex flex-col gap-3 border-t border-white/[0.06] pt-6 sm:flex-row sm:items-center sm:justify-between">
        <p class="text-[12px] text-zinc-600">
          © {{ year }} Poly Cloud — open source, self-hosted.
        </p>
        <p class="text-[12px] text-zinc-600">
          File operations powered by
          <a href="https://rclone.org" target="_blank" rel="noopener" class="text-zinc-500 transition-colors hover:text-white">rclone</a>.
        </p>
      </div>
    </div>
  </footer>
</template>
