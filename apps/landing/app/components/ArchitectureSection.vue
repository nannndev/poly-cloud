<script setup lang="ts">
const layers = [
  {
    name: 'Browser',
    detail: 'The Nuxt interface. Never touches the database or a provider directly.'
  },
  {
    name: 'Go API',
    detail: 'Picks the upload destination, records metadata, and streams bytes without storing them.'
  },
  {
    name: 'rclone',
    detail: 'Carries out file operations at each provider. Runs on a private network, never exposed.'
  },
  {
    name: 'Providers',
    detail: 'Where your files actually live. Poly Cloud never copies their contents onto your server.'
  }
]
</script>

<template>
  <section id="architecture" class="border-b border-white/[0.06] scroll-mt-16">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <h2 class="max-w-xl text-2xl font-semibold tracking-tight text-white md:text-3xl">
        Your files stay yours, right where they are
      </h2>
      <p class="mt-4 max-w-[58ch] text-[15px] leading-relaxed text-zinc-400">
        The database stores metadata only: name, size, and which account holds a given file.
        File contents never pass through the server.
      </p>

      <!-- Alur ditampilkan sebagai rantai bernomor, bukan diagram kotak-panah,
           supaya tetap terbaca di layar sempit. -->
      <ol class="mt-12 grid gap-4 md:grid-cols-4">
        <li
          v-for="(l, i) in layers"
          :key="l.name"
          class="reveal relative rounded-2xl border border-white/[0.08] bg-ink-850 p-6"
          :style="{ transitionDelay: `${i * 70}ms` }"
        >
          <span class="font-mono text-xs text-accent-400">{{ String(i + 1).padStart(2, '0') }}</span>
          <h3 class="mt-2 text-[15px] font-semibold text-white">{{ l.name }}</h3>
          <p class="mt-2 text-sm leading-relaxed text-zinc-400">{{ l.detail }}</p>

          <span
            v-if="i < layers.length - 1"
            class="absolute -right-2 top-1/2 hidden size-4 -translate-y-1/2 text-zinc-700 md:block"
            aria-hidden="true"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="m9 5 7 7-7 7" />
            </svg>
          </span>
        </li>
      </ol>

      <div class="mt-6 rounded-2xl border border-white/[0.08] bg-ink-900 p-6">
        <p class="text-sm leading-relaxed text-zinc-400">
          <span class="font-medium text-zinc-200">One file is stored whole on one account.</span>
          Splitting a file across accounts is not available yet, so a file larger than the
          roomiest account's free space is rejected rather than quietly split.
        </p>
      </div>
    </div>
  </section>
</template>
