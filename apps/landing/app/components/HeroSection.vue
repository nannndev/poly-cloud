<script setup lang="ts">
import { GITHUB } from '~/config/site'

// Tak ada angka yang bisa dikira metrik produksi di sini. Hero membawa
// proposisi nilai dan dua jalur tindakan: pasang, atau baca kodenya.
</script>

<template>
  <section class="relative overflow-hidden border-b border-white/[0.06]">
    <!-- Cahaya latar yang sangat samar; tidak ada gradien penuh warna. -->
    <div
      class="pointer-events-none absolute -top-40 left-1/2 size-[36rem] -translate-x-1/2 rounded-full bg-accent-500/[0.07] blur-3xl"
      aria-hidden="true"
    />

    <div class="relative mx-auto grid max-w-6xl gap-12 px-4 pb-20 pt-16 sm:px-6 lg:grid-cols-[1.05fr_1fr] lg:items-center lg:gap-16 lg:pb-28 lg:pt-24">
      <div>
        <p class="mb-5 inline-flex items-center gap-2 rounded-full border border-white/[0.08] bg-ink-850 px-3 py-1 text-[11px] font-medium uppercase tracking-[0.14em] text-zinc-400">
          Self-hosted
        </p>

        <h1 class="text-pretty text-4xl font-semibold leading-[1.08] tracking-tight text-white md:text-5xl lg:text-[3.4rem]">
          Many cloud accounts.
          <span class="text-accent-400">One drive.</span>
        </h1>

        <p class="mt-5 max-w-[52ch] text-[15px] leading-relaxed text-zinc-400">
          Poly Cloud merges Google Drive, Dropbox, OneDrive, S3, and B2 into a single
          storage space. Uploads are routed automatically to whichever account has the
          most room left.
        </p>

        <div class="mt-8 flex flex-wrap items-center gap-3">
          <NuxtLink
            to="/#install"
            class="rounded-xl bg-accent-600 px-5 py-2.5 text-sm font-semibold text-white shadow-xs transition-all hover:bg-accent-500 active:translate-y-px"
          >Self-host it</NuxtLink>
          <a
            :href="GITHUB.repo"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center gap-2 rounded-xl border border-white/[0.1] bg-ink-850 px-5 py-2.5 text-sm font-semibold text-zinc-200 transition-colors hover:border-white/20 hover:text-white"
          >
            <svg viewBox="0 0 24 24" class="size-4" fill="currentColor" aria-hidden="true">
              <path d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.09.68-.22.68-.49l-.01-1.72c-2.78.62-3.37-1.37-3.37-1.37-.46-1.18-1.11-1.5-1.11-1.5-.91-.64.07-.62.07-.62 1 .07 1.53 1.06 1.53 1.06.89 1.57 2.34 1.12 2.91.85.09-.66.35-1.12.63-1.38-2.22-.26-4.56-1.14-4.56-5.07 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05A9.3 9.3 0 0 1 12 6.84c.85 0 1.71.12 2.51.34 1.91-1.33 2.75-1.05 2.75-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.94-2.34 4.81-4.57 5.06.36.32.68.94.68 1.9l-.01 2.818c0 .27.18.59.69.49A10.26 10.26 0 0 0 22 12.25C22 6.58 17.52 2 12 2Z" />
            </svg>
            Read the source
          </a>
        </div>

        <p class="mt-5 text-[13px] text-zinc-500">
          Free and open source. Runs entirely on your own machine.
        </p>
      </div>

      <div class="reveal relative">
        <div class="overflow-hidden rounded-2xl border border-white/[0.08] bg-ink-850 shadow-2xl shadow-black/40">
          <img
            src="/shots/explorer.png"
            alt="The Poly Cloud interface showing files from several cloud accounts in one list"
            width="2940"
            height="1990"
            fetchpriority="high"
            decoding="async"
            class="w-full"
          >
        </div>
        <p class="mt-3 text-center text-[11px] text-zinc-500">
          Interface preview. The figures shown come from sample data.
        </p>
      </div>
    </div>
  </section>
</template>
