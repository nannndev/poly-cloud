<script setup lang="ts">
// Daftar ini mencerminkan auth.OAuthProviders dan auth.KeyProviders di backend.
const providers = [
  { name: 'Google Drive', slug: 'googledrive', auth: 'OAuth' },
  { name: 'OneDrive', slug: 'microsoftonedrive', auth: 'OAuth' },
  { name: 'Dropbox', slug: 'dropbox', auth: 'OAuth' },
  { name: 'Amazon S3', slug: 'amazonwebservices', auth: 'Access key' },
  { name: 'Cloudflare R2', slug: 'cloudflare', auth: 'Access key' },
  { name: 'Backblaze B2', slug: 'backblaze', auth: 'Access key' }
]
</script>

<template>
  <section id="providers" class="border-b border-white/[0.06] bg-ink-900 scroll-mt-16">
    <div class="mx-auto max-w-6xl px-4 py-20 sm:px-6 lg:py-24">
      <div class="grid gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:items-center lg:gap-16">
        <div>
          <h2 class="text-2xl font-semibold tracking-tight text-white md:text-3xl">
            Six providers, one way to use them
          </h2>
          <p class="mt-4 max-w-[52ch] text-[15px] leading-relaxed text-zinc-400">
            OAuth providers connect through a consent page in the browser. Key-based
            providers need only a credentials form. Once connected, all of them are
            treated the same.
          </p>
          <p class="mt-4 max-w-[52ch] text-sm leading-relaxed text-zinc-500">
            File operations run through rclone, so adding a new provider touches neither
            the interface nor the database schema.
          </p>
        </div>

        <ul class="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <li
            v-for="p in providers"
            :key="p.slug"
            class="rounded-2xl border border-white/[0.08] bg-ink-850 px-4 py-5 text-center"
          >
            <img
              :src="`https://cdn.simpleicons.org/${p.slug}/a1a1aa`"
              :alt="p.name"
              width="28"
              height="28"
              loading="lazy"
              decoding="async"
              class="mx-auto size-7"
            >
            <p class="mt-3 text-[13px] font-medium text-zinc-200">{{ p.name }}</p>
            <p class="mt-0.5 text-[11px] text-zinc-500">{{ p.auth }}</p>
          </li>
        </ul>
      </div>
    </div>
  </section>
</template>
