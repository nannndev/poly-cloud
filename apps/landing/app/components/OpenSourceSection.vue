<script setup lang="ts">
import { GITHUB } from '~/config/site'
import { formatCount, useContributors, useRepoStats } from '~/composables/useGithub'

const { data: stats } = await useRepoStats()
const { data: contributors } = await useContributors()

/**
 * Angka repo hanya ditampilkan bila GitHub sempat menjawab saat build. Sebuah
 * kartu "0 stars" pada repo yang sebenarnya punya bintang lebih buruk daripada
 * tak menampilkan apa pun.
 */
const hasStats = computed(() => Boolean(stats.value?.pushedAt))

const metrics = computed(() => [
  { label: 'Stars', value: formatCount(stats.value?.stars ?? 0) },
  { label: 'Forks', value: formatCount(stats.value?.forks ?? 0) },
  { label: 'Open issues', value: formatCount(stats.value?.openIssues ?? 0) },
  { label: 'License', value: stats.value?.license ?? 'MIT' }
])

const topContributors = computed(() => (contributors.value ?? []).slice(0, 8))

const ways = [
  {
    title: 'Report what breaks',
    body: 'Bug reports with steps to reproduce are the most useful thing you can send.',
    href: GITHUB.issues,
    cta: 'Open an issue'
  },
  {
    title: 'Pick up a first issue',
    body: 'Issues labelled good first issue are scoped small on purpose.',
    href: GITHUB.goodFirstIssues,
    cta: 'Browse issues'
  },
  {
    title: 'Improve the docs',
    body: 'Setup steps that tripped you up are worth fixing for the next person.',
    href: GITHUB.contributing,
    cta: 'Read the guide'
  }
]
</script>

<template>
  <section id="open-source" class="border-b border-white/[0.06] scroll-mt-16">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <div class="grid gap-10 lg:grid-cols-[1fr_1fr] lg:items-start lg:gap-16">
        <div>
          <h2 class="text-2xl font-semibold tracking-tight text-white md:text-3xl">
            Built in the open
          </h2>
          <p class="mt-4 max-w-[52ch] text-[15px] leading-relaxed text-zinc-400">
            Every line of Poly Cloud is public: the Go API, the Nuxt interface, the database
            schema, and the architecture decisions behind them. Read it, fork it, or change
            what does not suit you.
          </p>

          <div v-if="hasStats" class="mt-8 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-white/[0.08] bg-white/[0.06] sm:grid-cols-4">
            <div v-for="m in metrics" :key="m.label" class="bg-ink-850 px-4 py-5 text-center">
              <p class="font-mono text-xl font-semibold text-white">{{ m.value }}</p>
              <p class="mt-1 text-[11px] uppercase tracking-[0.12em] text-zinc-500">{{ m.label }}</p>
            </div>
          </div>

          <div class="mt-8 flex flex-wrap items-center gap-3">
            <a
              :href="GITHUB.repo"
              target="_blank"
              rel="noopener"
              class="inline-flex items-center gap-2 rounded-xl bg-accent-600 px-5 py-2.5 text-sm font-semibold text-white shadow-xs transition-all hover:bg-accent-500 active:translate-y-px"
            >
              <svg viewBox="0 0 24 24" class="size-4" fill="currentColor" aria-hidden="true">
                <path d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.09.68-.22.68-.49l-.01-1.72c-2.78.62-3.37-1.37-3.37-1.37-.46-1.18-1.11-1.5-1.11-1.5-.91-.64.07-.62.07-.62 1 .07 1.53 1.06 1.53 1.06.89 1.57 2.34 1.12 2.91.85.09-.66.35-1.12.63-1.38-2.22-.26-4.56-1.14-4.56-5.07 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05A9.3 9.3 0 0 1 12 6.84c.85 0 1.71.12 2.51.34 1.91-1.33 2.75-1.05 2.75-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.94-2.34 4.81-4.57 5.06.36.32.68.94.68 1.9l-.01 2.818c0 .27.18.59.69.49A10.26 10.26 0 0 0 22 12.25C22 6.58 17.52 2 12 2Z" />
              </svg>
              Star on GitHub
            </a>
            <NuxtLink
              to="/support"
              class="rounded-xl border border-white/[0.1] bg-ink-850 px-5 py-2.5 text-sm font-semibold text-zinc-200 transition-colors hover:border-white/20 hover:text-white"
            >Support the project</NuxtLink>
          </div>
        </div>

        <div class="grid gap-4">
          <article
            v-for="(w, i) in ways"
            :key="w.title"
            class="reveal rounded-2xl border border-white/[0.08] bg-ink-850 p-6 transition-colors hover:border-accent-500/30"
            :style="{ transitionDelay: `${i * 60}ms` }"
          >
            <h3 class="text-[15px] font-semibold text-white">{{ w.title }}</h3>
            <p class="mt-2 text-sm leading-relaxed text-zinc-400">{{ w.body }}</p>
            <a
              :href="w.href"
              target="_blank"
              rel="noopener"
              class="mt-3 inline-flex items-center gap-1.5 text-[13px] font-medium text-accent-400 transition-colors hover:text-accent-300"
            >
              {{ w.cta }}
              <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m9 5 7 7-7 7" />
              </svg>
            </a>
          </article>
        </div>
      </div>

      <!-- Wajah kontributor hanya muncul bila memang ada yang terbaca; saat
           GitHub tak terjangkau saat build, ajakannya tetap berdiri sendiri. -->
      <div v-if="topContributors.length > 0" class="mt-12 flex flex-wrap items-center gap-4 rounded-2xl border border-white/[0.08] bg-ink-900 p-6">
        <div class="flex -space-x-2">
          <a
            v-for="c in topContributors"
            :key="c.login"
            :href="c.url"
            target="_blank"
            rel="noopener"
            :title="`${c.login} — ${c.contributions} commits`"
          >
            <img
              :src="c.avatar"
              :alt="c.login"
              width="36"
              height="36"
              loading="lazy"
              decoding="async"
              class="size-9 rounded-full border-2 border-ink-900 bg-ink-850 transition-transform hover:z-10 hover:scale-110"
            >
          </a>
        </div>
        <p class="text-sm text-zinc-400">
          Built by
          <NuxtLink to="/contributors" class="font-medium text-accent-400 transition-colors hover:text-accent-300">
            {{ contributors?.length }} contributor{{ contributors?.length === 1 ? '' : 's' }}
          </NuxtLink>
          and counting.
        </p>
      </div>

      <div v-else class="mt-12 flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-white/[0.08] bg-ink-900 p-6">
        <p class="text-sm text-zinc-400">
          Poly Cloud is built by contributors, not a company.
        </p>
        <NuxtLink
          to="/contributors"
          class="text-[13px] font-medium text-accent-400 transition-colors hover:text-accent-300"
        >See who builds it</NuxtLink>
      </div>
    </div>
  </section>
</template>
