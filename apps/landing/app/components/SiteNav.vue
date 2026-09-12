<script setup lang="ts">
const links = [
  { label: 'Cara kerja', href: '#cara-kerja' },
  { label: 'Provider', href: '#provider' },
  { label: 'Arsitektur', href: '#arsitektur' }
]
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-white/[0.06] bg-ink-950/85 backdrop-blur-xl">
    <nav class="mx-auto flex h-16 max-w-6xl items-center justify-between gap-6 px-4 sm:px-6">
      <a href="#" class="flex items-center gap-2.5 shrink-0">
        <span class="grid size-8 place-items-center rounded-xl border border-accent-500/25 bg-accent-500/10">
          <svg viewBox="0 0 24 24" class="size-4 text-accent-400" fill="none" stroke="currentColor" stroke-width="1.75" aria-hidden="true">
            <path d="M12 3 20 7.5v9L12 21 4 16.5v-9L12 3Z" stroke-linejoin="round" />
            <path d="m4 7.5 8 4.5 8-4.5M12 12v9" stroke-linejoin="round" />
          </svg>
        </span>
        <span class="text-sm font-semibold tracking-tight text-white">Poly Cloud</span>
      </a>

      <div class="hidden items-center gap-7 md:flex">
        <a
          v-for="link in links"
          :key="link.href"
          :href="link.href"
          class="text-[13px] font-medium text-zinc-400 transition-colors hover:text-white"
        >{{ link.label }}</a>
      </div>

      <a
        href="#pasang"
        class="rounded-xl bg-accent-600 px-3.5 py-2 text-[13px] font-semibold text-white shadow-xs transition-all hover:bg-accent-500 active:translate-y-px"
      >Pasang sendiri</a>
    </nav>
  </header>
</template>
