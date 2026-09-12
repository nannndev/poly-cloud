<script setup lang="ts">
const title = 'Poly Cloud - Satukan Semua Akun Cloud Jadi Satu Drive'
const description =
  'Platform self-hosted yang menggabungkan Google Drive, Dropbox, OneDrive, S3, dan B2 '
  + 'jadi satu ruang penyimpanan. Unggahan diarahkan otomatis ke akun dengan sisa ruang terbanyak.'

useHead({
  title,
  meta: [
    { charset: 'utf-8' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    { name: 'theme-color', content: '#09090b' }
  ]
})

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  ogType: 'website',
  twitterCard: 'summary_large_image'
})
</script>

<template>
  <NuxtPage />
</template>
