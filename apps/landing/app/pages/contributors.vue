<script setup lang="ts">
import { GITHUB } from '~/config/site'
import { useContributors } from '~/composables/useGithub'

const { data: contributors } = await useContributors()

const title = 'Contributors — Poly Cloud'
const description = 'The people who build and maintain Poly Cloud, pulled straight from the GitHub repository.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description
})

const total = computed(() => contributors.value?.length ?? 0)
const commits = computed(() =>
  (contributors.value ?? []).reduce((sum, c) => sum + c.contributions, 0))

/**
 * Bar proporsi memakai kontributor teratas sebagai acuan, bukan total commit.
 * Dengan total, seorang maintainer yang memegang 90% commit akan membuat semua
 * bar lain tak terlihat sama sekali.
 */
const topCount = computed(() => contributors.value?.[0]?.contributions ?? 1)
</script>

<template>
  <div class="min-h-[100dvh]">
    <SiteNav />

    <main>
      <section class="border-b border-white/[0.06]">
        <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:py-20">
          <NuxtLink
            to="/"
            class="inline-flex items-center gap-1.5 text-[13px] font-medium text-zinc-500 transition-colors hover:text-white"
          >
            <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="m15 5-7 7 7 7" />
            </svg>
            Back to home
          </NuxtLink>

          <h1 class="mt-6 text-3xl font-semibold tracking-tight text-white md:text-4xl">
            Contributors
          </h1>
          <p class="mt-4 max-w-[58ch] text-[15px] leading-relaxed text-zinc-400">
            Poly Cloud is built by people who wanted their own storage to stop being
            scattered. This list comes straight from the repository and is rebuilt on
            every deploy.
          </p>

          <div v-if="total > 0" class="mt-8 flex flex-wrap gap-px overflow-hidden rounded-2xl border border-white/[0.08] bg-white/[0.06]">
            <div class="flex-1 bg-ink-850 px-6 py-5">
              <p class="font-mono text-2xl font-semibold text-white">{{ total }}</p>
              <p class="mt-1 text-[11px] uppercase tracking-[0.12em] text-zinc-500">
                Contributor{{ total === 1 ? '' : 's' }}
              </p>
            </div>
            <div class="flex-1 bg-ink-850 px-6 py-5">
              <p class="font-mono text-2xl font-semibold text-white">{{ commits.toLocaleString('en-US') }}</p>
              <p class="mt-1 text-[11px] uppercase tracking-[0.12em] text-zinc-500">Commits</p>
            </div>
          </div>
        </div>
      </section>

      <section class="border-b border-white/[0.06] bg-ink-900">
        <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:py-20">
          <ul v-if="total > 0" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <li
              v-for="(c, i) in contributors"
              :key="c.login"
              class="reveal rounded-2xl border border-white/[0.08] bg-ink-850 p-5 transition-colors hover:border-accent-500/30"
              :style="{ transitionDelay: `${Math.min(i, 8) * 50}ms` }"
            >
              <a :href="c.url" target="_blank" rel="noopener" class="flex items-center gap-4">
                <img
                  :src="c.avatar"
                  :alt="c.login"
                  width="48"
                  height="48"
                  loading="lazy"
                  decoding="async"
                  class="size-12 shrink-0 rounded-full border border-white/[0.08] bg-ink-800"
                >
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-semibold text-white">{{ c.login }}</p>
                  <p class="mt-0.5 font-mono text-[11px] text-zinc-500">
                    {{ c.contributions.toLocaleString('en-US') }}
                    commit{{ c.contributions === 1 ? '' : 's' }}
                  </p>
                  <div class="mt-2 h-1 w-full overflow-hidden rounded-full bg-white/[0.06]">
                    <div
                      class="h-full rounded-full bg-accent-500"
                      :style="{ width: `${Math.max(4, Math.round((c.contributions / topCount) * 100))}%` }"
                    />
                  </div>
                </div>
              </a>
            </li>
          </ul>

          <!-- GitHub bisa tak terjangkau saat build (batas kuota, gangguan).
               Halaman tetap terbit dengan jalan keluar yang jujur. -->
          <div v-else class="rounded-2xl border border-white/[0.08] bg-ink-850 p-10 text-center">
            <p class="text-sm text-zinc-400">
              The contributor list could not be loaded at build time.
            </p>
            <a
              :href="`${GITHUB.repo}/graphs/contributors`"
              target="_blank"
              rel="noopener"
              class="mt-4 inline-block rounded-xl border border-white/[0.1] bg-ink-800 px-4 py-2 text-[13px] font-semibold text-zinc-200 transition-colors hover:border-white/20 hover:text-white"
            >View it on GitHub</a>
          </div>
        </div>
      </section>

      <section class="border-b border-white/[0.06]">
        <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:py-20">
          <div class="rounded-2xl border border-white/[0.08] bg-ink-850 p-8 sm:p-10">
            <h2 class="text-xl font-semibold tracking-tight text-white md:text-2xl">
              Want your face on this page?
            </h2>
            <p class="mt-3 max-w-[56ch] text-[15px] leading-relaxed text-zinc-400">
              Bug reports count. So do documentation fixes, a provider you tested, or a
              rough edge you smoothed out. Start with an issue labelled
              <span class="font-medium text-zinc-200">good first issue</span> — they are
              scoped small deliberately.
            </p>
            <div class="mt-6 flex flex-wrap gap-3">
              <a
                :href="GITHUB.goodFirstIssues"
                target="_blank"
                rel="noopener"
                class="rounded-xl bg-accent-600 px-5 py-2.5 text-sm font-semibold text-white shadow-xs transition-all hover:bg-accent-500 active:translate-y-px"
              >Find a first issue</a>
              <a
                :href="GITHUB.contributing"
                target="_blank"
                rel="noopener"
                class="rounded-xl border border-white/[0.1] bg-ink-800 px-5 py-2.5 text-sm font-semibold text-zinc-200 transition-colors hover:border-white/20 hover:text-white"
              >Contributing guide</a>
            </div>
          </div>
        </div>
      </section>
    </main>

    <SiteFooter />
  </div>
</template>
