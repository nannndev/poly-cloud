<script setup lang="ts">
const title = 'Poly Cloud — Many Cloud Accounts, One Drive'
const description =
  'Open-source, self-hosted storage aggregator that merges Google Drive, Dropbox, OneDrive, '
  + 'S3, and B2 into a single drive. Uploads route automatically to the account with the most free space.'

useHead({
  title,
  meta: [
    { charset: 'utf-8' },
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    { name: 'theme-color', content: '#0b0e14' }
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
