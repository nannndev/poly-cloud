<script setup lang="ts">
const layers = [
  {
    name: 'Peramban',
    detail: 'Antarmuka Nuxt. Tak pernah menyentuh basis data atau penyedia secara langsung.'
  },
  {
    name: 'API Go',
    detail: 'Menentukan tujuan unggahan, mencatat metadata, dan mengalirkan berkas tanpa menyimpannya.'
  },
  {
    name: 'rclone',
    detail: 'Menjalankan operasi berkas di tiap penyedia. Berjalan di jaringan privat, tak terekspos keluar.'
  },
  {
    name: 'Penyedia',
    detail: 'Tempat berkas benar-benar berada. Poly Cloud tidak pernah menyalin isinya ke server Anda.'
  }
]
</script>

<template>
  <section id="arsitektur" class="border-b border-white/[0.06] scroll-mt-16">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <h2 class="max-w-xl text-2xl font-semibold tracking-tight text-white md:text-3xl">
        Berkas tetap milik Anda, di tempatnya semula
      </h2>
      <p class="mt-4 max-w-[58ch] text-[15px] leading-relaxed text-zinc-400">
        Basis data hanya menyimpan metadata: nama, ukuran, dan di akun mana sebuah berkas berada.
        Isi berkas tidak pernah singgah di server.
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
          <span class="font-medium text-zinc-200">Satu berkas disimpan utuh di satu akun.</span>
          Pemecahan berkas lintas akun belum tersedia, jadi berkas yang lebih besar daripada sisa
          ruang akun terlega akan ditolak, bukan dipecah diam-diam.
        </p>
      </div>
    </div>
  </section>
</template>
