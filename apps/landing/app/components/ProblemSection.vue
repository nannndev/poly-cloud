<template>
  <section class="border-b border-white/[0.06] bg-ink-900">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <div class="max-w-2xl">
        <h2 class="text-2xl font-semibold tracking-tight text-white md:text-3xl">
          Ruang penyimpanan Anda sebenarnya besar. Masalahnya terpotong-potong.
        </h2>
        <p class="mt-4 max-w-[60ch] text-[15px] leading-relaxed text-zinc-400">
          Satu akun penuh sementara yang lain masih kosong. Mencari satu berkas berarti
          membuka empat situs. Poly Cloud menghilangkan sekat itu tanpa memindahkan
          apa pun keluar dari penyedia asalnya.
        </p>
      </div>

      <div class="mt-12 grid gap-px overflow-hidden rounded-2xl border border-white/[0.08] bg-white/[0.06] sm:grid-cols-2">
        <div class="bg-ink-850 p-7">
          <p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-500">Tanpa Poly Cloud</p>
          <ul class="mt-5 space-y-3.5 text-sm text-zinc-400">
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-zinc-600" />
              Membuka empat situs untuk mencari satu berkas
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-zinc-600" />
              Menebak sendiri akun mana yang masih muat
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-zinc-600" />
              Struktur folder berbeda-beda di tiap penyedia
            </li>
          </ul>
        </div>

        <div class="bg-ink-850 p-7">
          <p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-accent-400">Dengan Poly Cloud</p>
          <ul class="mt-5 space-y-3.5 text-sm text-zinc-300">
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-accent-500" />
              Satu daftar berkas untuk semua akun sekaligus
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-accent-500" />
              Tujuan unggahan dipilihkan otomatis oleh sistem
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-accent-500" />
              Folder Anda sendiri, lepas dari struktur penyedia
            </li>
          </ul>
        </div>
      </div>
    </div>
  </section>
</template>
