<script setup lang="ts">
import { REPO_URL } from '~/config/site'

const command = `git clone ${REPO_URL}
cd poly-cloud
cp .env.example .env
make dev`

const copied = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

async function copy() {
  try {
    await navigator.clipboard.writeText(command)
    copied.value = true
    clearTimeout(timer)
    timer = setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // Peramban tanpa izin papan klip: biarkan pengguna menyalin manual.
  }
}

onBeforeUnmount(() => clearTimeout(timer))
</script>

<template>
  <section id="install" class="border-b border-white/[0.06] scroll-mt-16">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <div class="grid gap-10 lg:grid-cols-[1fr_1.15fr] lg:items-start lg:gap-14">
        <div>
          <h2 class="text-2xl font-semibold tracking-tight text-white md:text-3xl">
            Run it on your own machine
          </h2>
          <p class="mt-4 max-w-[48ch] text-[15px] leading-relaxed text-zinc-400">
            The whole system runs through Docker Compose: the API, the interface,
            PostgreSQL, and rclone. There is no third-party service to sign up for.
          </p>

          <dl class="mt-8 space-y-4 text-sm">
            <div class="flex gap-3">
              <dt class="w-28 shrink-0 text-zinc-500">Requirements</dt>
              <dd class="text-zinc-300">Docker and Docker Compose</dd>
            </div>
            <div class="flex gap-3">
              <dt class="w-28 shrink-0 text-zinc-500">Interface</dt>
              <dd class="font-mono text-zinc-300">localhost:3000</dd>
            </div>
            <div class="flex gap-3">
              <dt class="w-28 shrink-0 text-zinc-500">API</dt>
              <dd class="font-mono text-zinc-300">localhost:8080</dd>
            </div>
          </dl>

          <p class="mt-8 max-w-[48ch] text-sm leading-relaxed text-zinc-500">
            Key-based providers like S3 work straight away. For Google Drive, Dropbox, and
            OneDrive, register your own OAuth app and put its credentials in the
            <code class="font-mono text-zinc-400">.env</code> file.
          </p>
        </div>

        <div class="overflow-hidden rounded-2xl border border-white/[0.08] bg-ink-850">
          <div class="flex items-center justify-between border-b border-white/[0.06] px-4 py-2.5">
            <span class="font-mono text-[11px] text-zinc-500">bash</span>
            <button
              type="button"
              class="rounded-lg px-2.5 py-1 text-[11px] font-medium text-zinc-400 transition-colors hover:bg-white/[0.06] hover:text-white"
              @click="copy"
            >{{ copied ? 'Copied' : 'Copy' }}</button>
          </div>
          <pre class="overflow-x-auto px-5 py-5 font-mono text-[13px] leading-relaxed text-zinc-300"><code>{{ command }}</code></pre>
        </div>
      </div>
    </div>
  </section>
</template>
