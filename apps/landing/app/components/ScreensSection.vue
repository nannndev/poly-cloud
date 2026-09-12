<script setup lang="ts">
const shots = [
  {
    src: '/shots/accounts.png',
    alt: 'The connected accounts page showing remaining quota for each provider',
    caption: 'Per-account quota and the total, read from each provider on sync.'
  },
  {
    src: '/shots/connect.png',
    alt: 'The dialog for connecting a new storage account',
    caption: 'Add an account from the interface, without editing a config file.'
  }
]
</script>

<template>
  <section class="border-b border-white/[0.06] bg-ink-900">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <h2 class="max-w-xl text-2xl font-semibold tracking-tight text-white md:text-3xl">
        Inside the app
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
        Screenshots use sample data for preview purposes.
      </p>
    </div>
  </section>
</template>
