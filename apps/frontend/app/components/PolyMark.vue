<script setup lang="ts">
/**
 * Lambang "Arch" dari docs/logo. Digambar sekali di sini lalu dipakai ulang,
 * bukan disalin ke tiap tempat — supaya perubahan merek cukup satu berkas.
 *
 * Geometrinya persis mengikuti berkas sumber (viewBox 64, rak bawah 44x9 di
 * y=46, rak atas 36x9 di y=34, lengkung r=14). `currentColor` tak dipakai:
 * lambang ini bertumpu pada dua nada yang berbeda, dan meratakan keduanya akan
 * menghilangkan kedalaman rak yang jadi ciri khasnya.
 */
withDefaults(defineProps<{
  /** 'brand' memakai warna merek; 'mono' mengikuti warna teks induknya. */
  tone?: 'brand' | 'mono'
}>(), { tone: 'brand' })
</script>

<template>
  <svg viewBox="0 0 64 64" fill="none" aria-hidden="true">
    <rect
      x="14" y="34" width="36" height="9" rx="4.5"
      :fill="tone === 'mono' ? 'currentColor' : '#C4CBFA'"
      :opacity="tone === 'mono' ? 0.5 : 1"
    />
    <rect
      x="10" y="46" width="44" height="9" rx="4.5"
      :fill="tone === 'mono' ? 'currentColor' : '#8F9BF5'"
    />
    <path
      d="M18 26 A 14 14 0 0 1 46 26"
      fill="none"
      :stroke="tone === 'mono' ? 'currentColor' : '#8F9BF5'"
      stroke-width="9"
      stroke-linecap="round"
    />
  </svg>
</template>
