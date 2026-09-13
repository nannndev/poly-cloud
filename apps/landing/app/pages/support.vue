<script setup lang="ts">
import { CRYPTO, DONATIONS, GITHUB } from '~/config/site'

const title = 'Support Poly Cloud'
const description = 'Ways to support Poly Cloud: sponsor the work, send a one-off tip, or contribute code and documentation.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description
})

// Hanya kanal yang akunnya sudah ada. Menampilkan tautan donasi yang mati lebih
// buruk daripada tak menampilkannya sama sekali.
const active = DONATIONS.filter(d => d.enabled)
const global = active.filter(d => d.region === 'global')
const indonesia = active.filter(d => d.region === 'id')

/**
 * Menyumbang bukan satu-satunya cara membantu, dan bagi sebagian orang bukan
 * cara yang mungkin. Kanal non-uang ditaruh setara, bukan sebagai catatan kaki.
 */
const otherWays = [
  {
    title: 'Report a bug',
    body: 'A reproducible bug report saves more time than it takes to write.',
    href: GITHUB.issues,
    cta: 'Open an issue'
  },
  {
    title: 'Send a pull request',
    body: 'Code, tests, or a typo fix in the docs — all of it lands the same way.',
    href: GITHUB.contributing,
    cta: 'Read the guide'
  },
  {
    title: 'Tell someone',
    body: 'Star the repo or mention it to someone whose storage is as scattered as yours was.',
    href: GITHUB.repo,
    cta: 'Star on GitHub'
  }
]

// Alamat kripto disalin, bukan ditautkan: menempelkannya ke href memancing
// dompet membuka transaksi dengan jumlah yang belum diperiksa pengirim.
const copiedAddress = ref('')
let timer: ReturnType<typeof setTimeout> | undefined

async function copyAddress(address: string) {
  try {
    await navigator.clipboard.writeText(address)
    copiedAddress.value = address
    clearTimeout(timer)
    timer = setTimeout(() => { copiedAddress.value = '' }, 2000)
  } catch {
    // Peramban tanpa izin papan klip: pengguna menyalin manual dari teksnya.
  }
}

onBeforeUnmount(() => clearTimeout(timer))
</script>

<template>
  <div class="min-h-[100dvh]">
    <SiteNav />

    <main>
      <section class="relative overflow-hidden border-b border-white/[0.06]">
        <div
          class="pointer-events-none absolute -top-40 left-1/2 size-[32rem] -translate-x-1/2 rounded-full bg-accent-500/[0.07] blur-3xl"
          aria-hidden="true"
        />

        <div class="relative mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:py-20">
          <NuxtLink
            to="/"
            class="inline-flex items-center gap-1.5 text-[13px] font-medium text-zinc-500 transition-colors hover:text-white"
          >
            <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="m15 5-7 7 7 7" />
            </svg>
            Back to home
          </NuxtLink>

          <h1 class="mt-6 max-w-[20ch] text-3xl font-semibold tracking-tight text-white md:text-4xl">
            Poly Cloud is free. Keeping it going is not.
          </h1>
          <p class="mt-4 max-w-[58ch] text-[15px] leading-relaxed text-zinc-400">
            There is no paid tier and no hosted plan to upsell — you run it on your own
            machine and pay nothing. If it saved you the cost of another storage
            subscription, one of the options below sends some of that back.
          </p>

          <p class="mt-6 max-w-[58ch] rounded-2xl border border-white/[0.08] bg-ink-850 p-5 text-sm leading-relaxed text-zinc-400">
            <span class="font-medium text-zinc-200">Nothing is gated behind this.</span>
            Every feature stays available whether or not you give anything, and none of
            these pages track who does.
          </p>
        </div>
      </section>

      <section class="border-b border-white/[0.06] bg-ink-900">
        <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:py-20">
          <h2 class="text-2xl font-semibold tracking-tight text-white md:text-3xl">
            One-off or recurring
          </h2>
          <p class="mt-3 max-w-[52ch] text-[15px] leading-relaxed text-zinc-400">
            Pick whichever you already have an account with. They all reach the same place.
          </p>

          <ul class="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <li
              v-for="(d, i) in global"
              :key="d.id"
              class="reveal"
              :style="{ transitionDelay: `${i * 60}ms` }"
            >
              <a
                :href="d.url"
                target="_blank"
                rel="noopener"
                class="flex h-full flex-col rounded-2xl border border-white/[0.08] bg-ink-850 p-6 transition-colors hover:border-accent-500/30"
              >
                <span
                  class="grid size-10 place-items-center rounded-xl border border-white/[0.07] bg-ink-800"
                  :style="{ color: d.tint }"
                >
                  <svg v-if="d.icon === 'github'" viewBox="0 0 24 24" class="size-5" fill="currentColor" aria-hidden="true">
                    <path d="M12 2C6.48 2 2 6.58 2 12.25c0 4.53 2.87 8.37 6.84 9.73.5.09.68-.22.68-.49l-.01-1.72c-2.78.62-3.37-1.37-3.37-1.37-.46-1.18-1.11-1.5-1.11-1.5-.91-.64.07-.62.07-.62 1 .07 1.53 1.06 1.53 1.06.89 1.57 2.34 1.12 2.91.85.09-.66.35-1.12.63-1.38-2.22-.26-4.56-1.14-4.56-5.07 0-1.12.39-2.03 1.03-2.75-.1-.26-.45-1.3.1-2.71 0 0 .84-.28 2.75 1.05A9.3 9.3 0 0 1 12 6.84c.85 0 1.71.12 2.51.34 1.91-1.33 2.75-1.05 2.75-1.05.55 1.41.2 2.45.1 2.71.64.72 1.03 1.63 1.03 2.75 0 3.94-2.34 4.81-4.57 5.06.36.32.68.94.68 1.9l-.01 2.818c0 .27.18.59.69.49A10.26 10.26 0 0 0 22 12.25C22 6.58 17.52 2 12 2Z" />
                  </svg>
                  <svg v-else-if="d.icon === 'kofi'" viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M4 8h12v6a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4Z" />
                    <path d="M16 10h2a2.5 2.5 0 0 1 0 5h-2" />
                    <path d="M7 3.5v1.8M11 3.5v1.8" />
                  </svg>
                  <svg v-else viewBox="0 0 24 24" class="size-5" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M6.5 19 9 5h5.5a3.5 3.5 0 0 1 0 7H10" />
                    <path d="M9.5 15h4a3.5 3.5 0 0 0 0-7" />
                  </svg>
                </span>

                <h3 class="mt-4 text-[15px] font-semibold text-white">{{ d.name }}</h3>
                <p class="mt-2 flex-1 text-sm leading-relaxed text-zinc-400">{{ d.note }}</p>

                <span class="mt-4 inline-flex items-center gap-1.5 text-[13px] font-medium text-accent-400">
                  Open
                  <svg viewBox="0 0 24 24" class="size-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M7 17 17 7M9 7h8v8" />
                  </svg>
                </span>
              </a>
            </li>
          </ul>

          <!-- Kanal lokal dipisah: keduanya menerima QRIS/e-wallet, yang tak
               berguna bagi penyumbang di luar Indonesia dan sebaliknya. -->
          <div v-if="indonesia.length > 0" class="mt-12">
            <h3 class="flex items-center gap-2.5 text-[15px] font-semibold text-white">
              From Indonesia
              <span class="rounded-full border border-white/[0.08] bg-ink-850 px-2.5 py-0.5 text-[11px] font-medium text-zinc-400">
                QRIS &amp; e-wallet
              </span>
            </h3>
            <p class="mt-2 max-w-[52ch] text-sm leading-relaxed text-zinc-500">
              No card needed — these accept GoPay, OVO, Dana, ShopeePay, and bank transfer.
            </p>

            <ul class="mt-5 grid gap-4 sm:grid-cols-2">
              <li v-for="d in indonesia" :key="d.id">
                <a
                  :href="d.url"
                  target="_blank"
                  rel="noopener"
                  class="flex h-full items-start gap-4 rounded-2xl border border-white/[0.08] bg-ink-850 p-5 transition-colors hover:border-accent-500/30"
                >
                  <span
                    class="grid size-10 shrink-0 place-items-center rounded-xl border border-white/[0.07] bg-ink-800 text-sm font-bold"
                    :style="{ color: d.tint }"
                  >{{ d.name.charAt(0) }}</span>
                  <span class="min-w-0">
                    <span class="block text-[15px] font-semibold text-white">{{ d.name }}</span>
                    <span class="mt-1 block text-sm leading-relaxed text-zinc-400">{{ d.note }}</span>
                  </span>
                </a>
              </li>
            </ul>
          </div>

          <div v-if="CRYPTO.length > 0" class="mt-12">
            <h3 class="text-[15px] font-semibold text-white">Crypto</h3>
            <ul class="mt-5 grid gap-3 sm:grid-cols-2">
              <li
                v-for="c in CRYPTO"
                :key="c.symbol"
                class="rounded-2xl border border-white/[0.08] bg-ink-850 p-5"
              >
                <div class="flex items-center justify-between gap-3">
                  <span class="text-[13px] font-semibold text-white">
                    {{ c.symbol }}
                    <span class="ml-1 font-normal text-zinc-500">{{ c.network }}</span>
                  </span>
                  <button
                    type="button"
                    class="rounded-lg px-2.5 py-1 text-[11px] font-medium text-zinc-400 transition-colors hover:bg-white/[0.06] hover:text-white"
                    @click="copyAddress(c.address)"
                  >{{ copiedAddress === c.address ? 'Copied' : 'Copy' }}</button>
                </div>
                <p class="mt-2 break-all font-mono text-[11px] leading-relaxed text-zinc-500">
                  {{ c.address }}
                </p>
              </li>
            </ul>
          </div>
        </div>
      </section>

      <section class="border-b border-white/[0.06]">
        <div class="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:py-20">
          <h2 class="text-2xl font-semibold tracking-tight text-white md:text-3xl">
            Not everything helpful costs money
          </h2>
          <p class="mt-3 max-w-[52ch] text-[15px] leading-relaxed text-zinc-400">
            These are worth as much as a donation, and some weeks more.
          </p>

          <div class="mt-10 grid gap-4 lg:grid-cols-3">
            <article
              v-for="(w, i) in otherWays"
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

          <div class="mt-10 rounded-2xl border border-white/[0.08] bg-ink-900 p-6">
            <p class="text-sm leading-relaxed text-zinc-400">
              <span class="font-medium text-zinc-200">Where the money goes.</span>
              Domain renewal, the machine that builds and tests releases, and the hours
              spent on issues and reviews. There is no company behind this and no
              investor to pay back.
            </p>
          </div>
        </div>
      </section>
    </main>

    <SiteFooter />
  </div>
</template>
