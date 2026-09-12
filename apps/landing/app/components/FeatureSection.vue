<script setup lang="ts">
// Setiap butir di sini adalah kemampuan yang sudah berjalan, bukan rencana.
const features = [
  {
    title: 'Penempatan otomatis',
    body: 'Saat mengunggah, sistem memilih akun dengan sisa ruang terbanyak. Anda tidak perlu memikirkan tujuannya.',
    icon: 'route'
  },
  {
    title: 'Folder virtual',
    body: 'Struktur folder hidup di basis data platform. Memindahkan berkas antar folder hanya mengubah catatan, tidak memindahkan data.',
    icon: 'folder'
  },
  {
    title: 'Tanpa singgah di server',
    body: 'Unggahan dan unduhan dialirkan langsung antara peramban dan penyedia. Server tidak menyimpan salinan berkas.',
    icon: 'stream'
  },
  {
    title: 'Token terenkripsi',
    body: 'Kredensial penyedia disimpan terenkripsi AES-256-GCM di PostgreSQL, terpisah dari data akun lainnya.',
    icon: 'lock'
  },
  {
    title: 'Pencarian lintas akun',
    body: 'Metadata berkas diindeks di basis data, jadi pencarian dijawab tanpa menunggu respons penyedia.',
    icon: 'search'
  }
]
</script>

<template>
  <section id="cara-kerja" class="border-b border-white/[0.06] scroll-mt-16">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <h2 class="max-w-xl text-2xl font-semibold tracking-tight text-white md:text-3xl">
        Yang dikerjakan Poly Cloud untuk Anda
      </h2>

      <!-- Lima butir dalam susunan asimetris: dua kartu lebar di baris pertama,
           tiga kartu sempit di baris kedua. Jumlah sel persis sama dengan isi. -->
      <div class="mt-12 grid gap-4 lg:grid-cols-6">
        <article
          v-for="(f, i) in features"
          :key="f.title"
          class="reveal rounded-2xl border border-white/[0.08] bg-ink-850 p-6 transition-colors hover:border-accent-500/30"
          :class="i < 2 ? 'lg:col-span-3' : 'lg:col-span-2'"
          :style="{ transitionDelay: `${i * 60}ms` }"
        >
          <span class="grid size-9 place-items-center rounded-xl border border-white/[0.07] bg-ink-800 text-accent-400">
            <svg viewBox="0 0 24 24" class="size-[18px]" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <template v-if="f.icon === 'route'">
                <circle cx="6" cy="19" r="2.5" /><circle cx="18" cy="5" r="2.5" />
                <path d="M8.5 19h5a4 4 0 0 0 0-8h-3a4 4 0 0 1 0-8h5" />
              </template>
              <template v-else-if="f.icon === 'folder'">
                <path d="M3 7.5A1.5 1.5 0 0 1 4.5 6h4l2 2.5h7A1.5 1.5 0 0 1 19 10v7.5a1.5 1.5 0 0 1-1.5 1.5h-13A1.5 1.5 0 0 1 3 17.5Z" />
              </template>
              <template v-else-if="f.icon === 'stream'">
                <path d="M4 8h10M4 12h16M4 16h10" /><path d="m17 5 3 3-3 3" />
              </template>
              <template v-else-if="f.icon === 'lock'">
                <rect x="4.5" y="10.5" width="15" height="9.5" rx="2" />
                <path d="M8 10.5V7.5a4 4 0 0 1 8 0v3" />
              </template>
              <template v-else>
                <circle cx="11" cy="11" r="6" /><path d="m20 20-4.5-4.5" />
              </template>
            </svg>
          </span>

          <h3 class="mt-4 text-[15px] font-semibold text-white">{{ f.title }}</h3>
          <p class="mt-2 text-sm leading-relaxed text-zinc-400">{{ f.body }}</p>
        </article>
      </div>
    </div>
  </section>
</template>
