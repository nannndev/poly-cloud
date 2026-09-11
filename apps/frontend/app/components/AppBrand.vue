<script setup lang="ts">
withDefaults(defineProps<{
  compact?: boolean
}>(), {
  compact: false
})
</script>

<template>
  <NuxtLink to="/" class="inline-flex items-center gap-3 group select-none" aria-label="Poly Cloud Home">
    <!-- Custom Geometric Poly Cloud Mark -->
    <div class="relative flex size-10 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-b from-[#151c18] to-[#0d120f] border border-emerald-500/25 shadow-xs transition-all duration-200 group-hover:scale-105 group-hover:border-emerald-500/40">
      <svg viewBox="0 0 36 36" fill="none" xmlns="http://www.w3.org/2000/svg" class="size-6">
        <defs>
          <linearGradient id="poly-grad-top" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#6ee7b7" />
            <stop offset="100%" stop-color="#34d399" />
          </linearGradient>
          <linearGradient id="poly-grad-mid" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#10b981" />
            <stop offset="100%" stop-color="#059669" />
          </linearGradient>
          <linearGradient id="poly-grad-dark" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#059669" />
            <stop offset="100%" stop-color="#064e3b" />
          </linearGradient>
        </defs>

        <!-- Top Facets -->
        <polygon points="18,7 25,12 18,16 11,12" fill="url(#poly-grad-top)" />
        
        <!-- Left Wing Facets -->
        <polygon points="6,15 11,12 18,16 11,20 6,20" fill="url(#poly-grad-mid)" />
        <polygon points="6,20 11,20 14,26 8,26" fill="url(#poly-grad-dark)" />

        <!-- Right Wing Facets -->
        <polygon points="18,16 25,12 30,15 30,20 25,20" fill="url(#poly-grad-mid)" opacity="0.9" />
        <polygon points="25,20 30,20 28,26 22,26" fill="url(#poly-grad-dark)" />

        <!-- Central Core Facets -->
        <polygon points="11,20 18,16 25,20 18,25" fill="url(#poly-grad-top)" opacity="0.85" />
        <polygon points="11,20 18,25 14,28 14,26" fill="url(#poly-grad-dark)" />
        <polygon points="18,25 25,20 22,26 18,28" fill="url(#poly-grad-mid)" />

        <!-- Crisp Geometric Facet Edges -->
        <path
          d="M18 7L25 12L30 15V20L28 26H8L6 20V15L11 12L18 7Z"
          stroke="#a7f3d0"
          stroke-width="0.75"
          stroke-linejoin="round"
          opacity="0.4"
        />
        <path
          d="M18 7V16M18 16L11 12M18 16L25 12M18 16V25M11 20L18 16L25 20M14 26L18 25L22 26"
          stroke="#ffffff"
          stroke-width="0.6"
          stroke-linejoin="round"
          opacity="0.55"
        />
      </svg>

      <!-- Soft Active Status Dot -->
      <span class="absolute -bottom-0.5 -right-0.5 size-2.5 rounded-full bg-emerald-400 ring-2 ring-[#0c0c0e]" />
    </div>

    <!-- Symmetrical Clean Brand Title -->
    <div v-if="!compact" class="leading-none flex flex-col justify-center">
      <span class="font-extrabold tracking-wider text-[13.5px] text-white uppercase font-sans">
        Poly Cloud
      </span>
      <span class="block text-[9px] font-semibold tracking-[0.14em] text-zinc-500 uppercase mt-1.5">
        Cloud Storage Aggregator
      </span>
    </div>
  </NuxtLink>
</template>
