<script setup lang="ts">
withDefaults(defineProps<{
  compact?: boolean
}>(), {
  compact: false
})
</script>

<template>
  <NuxtLink to="/" class="group inline-flex select-none items-center gap-3" aria-label="Poly Cloud Home">
    <!-- Latar mengikuti app icon gelap resmi (#1E2430, sudut membulat 22.6%). -->
    <div
      class="relative flex size-10 shrink-0 items-center justify-center rounded-2xl border border-brand-400/25 bg-[#1E2430] shadow-xs transition-all duration-200 group-hover:scale-105 group-hover:border-brand-400/45"
    >
      <PolyMark class="size-6" />
      <span class="absolute -bottom-0.5 -right-0.5 size-2.5 rounded-full bg-brand-400 ring-2 ring-[#0d111a]" />
    </div>

    <div v-if="!compact" class="flex flex-col justify-center leading-none">
      <span class="font-sans text-[13.5px] font-extrabold uppercase tracking-wider text-white">
        Poly Cloud
      </span>
      <!-- Tagline resmi dari lockup merek, bukan deskripsi yang ditulis ulang. -->
      <span class="mt-1.5 block text-[9px] font-semibold uppercase tracking-[0.14em] text-zinc-500">
        One space, every drive
      </span>
    </div>
  </NuxtLink>
</template>
