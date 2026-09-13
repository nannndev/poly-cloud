<script setup lang="ts">
import { GITHUB } from '~/config/site'
import { formatCount, useRepoStats } from '~/composables/useGithub'

const links = [
  { label: 'How it works', to: '/#how-it-works' },
  { label: 'Providers', to: '/#providers' },
  { label: 'Architecture', to: '/#architecture' },
  { label: 'Contributors', to: '/contributors' },
  { label: 'Support', to: '/support' }
]

const { data: stats } = await useRepoStats()

// Menu mobil: tanpa ini, lima tautan + dua tombol tak muat di layar sempit.
const isOpen = ref(false)
const route = useRoute()
watch(() => route.fullPath, () => { isOpen.value = false })
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-white/[0.06] bg-ink-950/85 backdrop-blur-xl">
    <nav class="mx-auto flex h-16 max-w-6xl items-center justify-between gap-6 px-4 sm:px-6">
      <NuxtLink to="/" class="flex shrink-0 items-center gap-2.5">
        <span class="grid size-8 place-items-center rounded-xl border border-accent-400/25 bg-[#1E2430]">
          <PolyMark class="size-[19px]" />
        </span>
        <span class="text-sm font-semibold tracking-tight text-white">Poly Cloud</span>
      </NuxtLink>

      <div class="hidden items-center gap-7 md:flex">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="text-[13px] font-medium text-zinc-400 transition-colors hover:text-white"
        >{{ link.label }}</NuxtLink>
      </div>

      <div class="flex items-center gap-2">
        <!-- Jumlah bintang hanya tampil bila GitHub sempat menjawab saat build;
             "0 stars" pada repo yang belum terbaca akan menyesatkan. -->
        <a
          :href="GITHUB.repo"
          target="_blank"
          rel="noopener"
          class="hidden items-center gap-2 rounded-xl border border-white/[0.1] bg-ink-850 px-3 py-2 text-[13px] font-medium text-zinc-300 transition-colors hover:border-white/20 hover:text-white sm:inline-flex"
        >
          <svg viewBox="0 0 24 24" class="size-4" fill="currentColor" aria-hidden="true">
            <path d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.09.68-.22.68-.49l-.01-1.72c-2.78.62-3.37-1.37-3.37-1.37-.46-1.18-1.11-1.5-1.11-1.5-.91-.64.07-.62.07-.62 1 .07 1.53 1.06 1.53 1.06.89 1.57 2.34 1.12 2.91.85.09-.66.35-1.12.63-1.38-2.22-.26-4.56-1.14-4.56-5.07 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05A9.3 9.3 0 0 1 12 6.84c.85 0 1.71.12 2.51.34 1.91-1.33 2.75-1.05 2.75-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.94-2.34 4.81-4.57 5.06.36.32.68.94.68 1.9l-.01 2.818c0 .27.18.59.69.49A10.26 10.26 0 0 0 22 12.25C22 6.58 17.52 2 12 2Z" />
          </svg>
          <span>GitHub</span>
          <span v-if="stats && stats.stars > 0" class="font-mono text-[11px] text-zinc-500">
            {{ formatCount(stats.stars) }}
          </span>
        </a>

        <NuxtLink
          to="/#install"
          class="hidden rounded-xl bg-accent-600 px-3.5 py-2 text-[13px] font-semibold text-white shadow-xs transition-all hover:bg-accent-500 active:translate-y-px sm:block"
        >Self-host it</NuxtLink>

        <button
          type="button"
          class="grid size-9 place-items-center rounded-xl border border-white/[0.1] bg-ink-850 text-zinc-300 transition-colors hover:text-white md:hidden"
          :aria-expanded="isOpen"
          aria-label="Toggle navigation menu"
          @click="isOpen = !isOpen"
        >
          <svg viewBox="0 0 24 24" class="size-4" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
            <template v-if="isOpen"><path d="m6 6 12 12M18 6 6 18" /></template>
            <template v-else><path d="M4 7h16M4 12h16M4 17h16" /></template>
          </svg>
        </button>
      </div>
    </nav>

    <div v-if="isOpen" class="border-t border-white/[0.06] bg-ink-950 md:hidden">
      <div class="mx-auto flex max-w-6xl flex-col gap-1 px-4 py-3 sm:px-6">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="rounded-lg px-2 py-2.5 text-sm font-medium text-zinc-300 transition-colors hover:bg-white/[0.05] hover:text-white"
        >{{ link.label }}</NuxtLink>

        <div class="mt-2 flex gap-2 border-t border-white/[0.06] pt-3">
          <a
            :href="GITHUB.repo"
            target="_blank"
            rel="noopener"
            class="flex-1 rounded-xl border border-white/[0.1] bg-ink-850 px-3 py-2 text-center text-[13px] font-medium text-zinc-200"
          >GitHub</a>
          <NuxtLink
            to="/#install"
            class="flex-1 rounded-xl bg-accent-600 px-3 py-2 text-center text-[13px] font-semibold text-white"
          >Self-host it</NuxtLink>
        </div>
      </div>
    </div>
  </header>
</template>
