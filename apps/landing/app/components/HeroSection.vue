<script setup lang="ts">
// Empat angka yang mungkin dikira metrik produksi sengaja tidak dipakai di sini.
// Hero hanya membawa proposisi nilai dan satu jalur tindakan.
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
          Banyak akun cloud.
          <span class="text-accent-400">Satu drive.</span>
        </h1>

        <p class="mt-5 max-w-[52ch] text-[15px] leading-relaxed text-zinc-400">
          Poly Cloud menyatukan Google Drive, Dropbox, OneDrive, S3, dan B2 jadi satu ruang
          penyimpanan. Unggahan otomatis diarahkan ke akun yang ruangnya paling lega.
        </p>

        <div class="mt-8 flex flex-wrap items-center gap-3">
          <a
            href="#pasang"
            class="rounded-xl bg-accent-600 px-5 py-2.5 text-sm font-semibold text-white shadow-xs transition-all hover:bg-accent-500 active:translate-y-px"
          >Pasang sendiri</a>
          <a
            href="#cara-kerja"
            class="rounded-xl border border-white/[0.1] bg-ink-850 px-5 py-2.5 text-sm font-semibold text-zinc-200 transition-colors hover:border-white/20 hover:text-white"
          >Lihat cara kerjanya</a>
        </div>
      </div>

      <div class="reveal relative">
        <div class="overflow-hidden rounded-2xl border border-white/[0.08] bg-ink-850 shadow-2xl shadow-black/40">
          <img
            src="/shots/explorer.png"
            alt="Antarmuka Poly Cloud menampilkan berkas dari beberapa akun cloud dalam satu daftar"
            width="2940"
            height="1990"
            fetchpriority="high"
            decoding="async"
            class="w-full"
          >
        </div>
        <p class="mt-3 text-center text-[11px] text-zinc-500">
          Pratinjau antarmuka. Angka pada gambar berasal dari data contoh.
        </p>
      </div>
    </div>
  </section>
</template>
