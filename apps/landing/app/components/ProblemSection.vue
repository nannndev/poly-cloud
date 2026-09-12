<template>
  <section class="border-b border-white/[0.06] bg-ink-900">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <div class="max-w-2xl">
        <h2 class="text-2xl font-semibold tracking-tight text-white md:text-3xl">
          You already have plenty of storage. It's just scattered.
        </h2>
        <p class="mt-4 max-w-[60ch] text-[15px] leading-relaxed text-zinc-400">
          One account fills up while another sits empty. Finding a single file means
          opening four different sites. Poly Cloud removes those walls without moving
          anything out of the provider it already lives on.
        </p>
      </div>

      <div class="mt-12 grid gap-px overflow-hidden rounded-2xl border border-white/[0.08] bg-white/[0.06] sm:grid-cols-2">
        <div class="bg-ink-850 p-7">
          <p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-500">Without Poly Cloud</p>
          <ul class="mt-5 space-y-3.5 text-sm text-zinc-400">
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-zinc-600" />
              Opening four sites to track down one file
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-zinc-600" />
              Guessing which account still has room
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-zinc-600" />
              A different folder structure on every provider
            </li>
          </ul>
        </div>

        <div class="bg-ink-850 p-7">
          <p class="text-[11px] font-semibold uppercase tracking-[0.16em] text-accent-400">With Poly Cloud</p>
          <ul class="mt-5 space-y-3.5 text-sm text-zinc-300">
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-accent-500" />
              One file list across every account at once
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-accent-500" />
              Upload destinations chosen for you automatically
            </li>
            <li class="flex gap-3">
              <span class="mt-2 size-1 shrink-0 rounded-full bg-accent-500" />
              Your own folders, independent of any provider
            </li>
          </ul>
        </div>
      </div>
    </div>
  </section>
</template>
