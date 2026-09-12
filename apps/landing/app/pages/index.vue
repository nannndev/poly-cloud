<script setup lang="ts">
// Bagian bertanda .reveal dimunculkan saat masuk viewport. IntersectionObserver
// dipakai alih-alih pendengar scroll agar tak berjalan tiap frame. Pengguna
// dengan preferensi gerak minimal melihat konten final seketika karena
// transisinya dimatikan lewat CSS.
let observer: IntersectionObserver | undefined

onMounted(() => {
  const targets = document.querySelectorAll('.reveal')

  if (!('IntersectionObserver' in window)) {
    return
  }

  // Menandai bahwa skrip hidup; sebelum baris ini konten sudah tampil penuh.
  document.documentElement.classList.add('js')

  observer = new IntersectionObserver((entries, obs) => {
    for (const entry of entries) {
      if (entry.isIntersecting) {
        entry.target.classList.add('is-visible')
        obs.unobserve(entry.target)
      }
    }
  }, { threshold: 0.12, rootMargin: '0px 0px -32px 0px' })

  targets.forEach(el => observer!.observe(el))
})

// Pelepasan didaftarkan di lingkup setup, bukan di dalam onMounted, supaya
// benar-benar terpasang pada siklus hidup komponen.
onBeforeUnmount(() => observer?.disconnect())
</script>

<template>
  <div class="min-h-[100dvh]">
    <SiteNav />
    <main>
      <HeroSection />
      <ProblemSection />
      <FeatureSection />
      <ProviderSection />
      <ArchitectureSection />
      <ScreensSection />
      <InstallSection />
    </main>
    <SiteFooter />
  </div>
</template>
