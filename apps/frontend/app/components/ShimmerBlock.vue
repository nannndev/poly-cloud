<script setup lang="ts">
/**
 * Pelat penanda muat dengan sapuan shimmer. Dipakai saat byte-nya sedang
 * dialirkan dari provider — pratinjau menunggu jaringan pihak ketiga, jadi
 * jedanya bisa terasa dan area kosong terbaca seperti gagal.
 *
 * Sapuannya dimatikan lewat @media prefers-reduced-motion di main.css, dan
 * pelatnya sendiri tetap tampil sebagai blok diam.
 */
withDefaults(defineProps<{
  /** Kelas tinggi/lebar dari pemanggil; pelat mengisi ruang yang diberikan. */
  rounded?: string
}>(), { rounded: 'rounded-2xl' })
</script>

<template>
  <div
    class="relative overflow-hidden border border-white/[0.06] bg-[#151a27]"
    :class="rounded"
    role="status"
    aria-label="Loading preview"
  >
    <div
      class="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.07] to-transparent animate-shimmer"
    />
    <slot />
  </div>
</template>
