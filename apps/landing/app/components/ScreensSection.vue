<script setup lang="ts">
const shots = [
  {
    src: '/shots/accounts.png',
    alt: 'Halaman akun terhubung menampilkan sisa kuota tiap penyedia',
    caption: 'Kuota tiap akun dan totalnya, dibaca dari penyedia saat sinkronisasi.'
  },
  {
    src: '/shots/connect.png',
    alt: 'Dialog menghubungkan akun penyimpanan baru',
    caption: 'Menambah akun dari antarmuka, tanpa menyentuh berkas konfigurasi.'
  }
]
</script>

<template>
  <section class="border-b border-white/[0.06] bg-ink-900">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <h2 class="max-w-xl text-2xl font-semibold tracking-tight text-white md:text-3xl">
        Di dalam aplikasi
      </h2>

      <div class="mt-10 grid gap-6 lg:grid-cols-2">
        <figure v-for="(s, i) in shots" :key="s.src" class="reveal" :style="{ transitionDelay: `${i * 80}ms` }">
          <div class="overflow-hidden rounded-2xl border border-white/[0.08] bg-ink-850">
            <img
              :src="s.src"
              :alt="s.alt"
              width="2940"
              height="1990"
              loading="lazy"
              decoding="async"
              class="w-full"
            >
          </div>
          <figcaption class="mt-3 text-[13px] text-zinc-500">{{ s.caption }}</figcaption>
        </figure>
      </div>

      <p class="mt-8 text-[11px] text-zinc-600">
        Tangkapan layar memakai data contoh untuk keperluan pratinjau.
      </p>
    </div>
  </section>
</template>
