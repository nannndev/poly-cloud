<script setup lang="ts">
const colorMode = useColorMode()

const color = computed(() => colorMode.value === 'dark' ? '#0b0e14' : '#ffffff')

useHead({
  meta: [
    { charset: 'utf-8' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    { key: 'theme-color', name: 'theme-color', content: color }
  ],
  link: [
    { rel: 'icon', href: '/favicon.svg', type: 'image/svg+xml' },
    { rel: 'apple-touch-icon', href: '/app-icon-192.png' }
  ],
  htmlAttrs: {
    lang: 'en'
  }
})

const title = 'Poly Cloud — One space, every drive'
const description = 'Unified cloud storage aggregator across Google Drive, OneDrive, Dropbox, and AWS S3 with smart routing, instant search, and centralized quota control.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  twitterCard: 'summary_large_image'
})
</script>

<template>
  <!-- Toast di atas-tengah: sudut kanan-bawah bertabrakan dengan modal progres
       unggah, yang justru paling sering terbuka saat toast terbit. -->
  <UApp :toaster="{ position: 'top-center', duration: 4000 }">
    <NuxtLoadingIndicator color="var(--ui-primary)" />

    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
