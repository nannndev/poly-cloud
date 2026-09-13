// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@pinia/nuxt',
    '@nuxt/ui',
    '@vueuse/nuxt'
  ],

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  ui: {
    // 'brand' bukan warna bawaan Tailwind, jadi harus didaftarkan di sini —
    // app.config.ts saja tak cukup: tanpa baris ini Nuxt UI tak pernah
    // mendefinisikan --ui-primary, dan seluruh komponen primary kehilangan
    // warnanya tanpa pesan galat.
    theme: {
      colors: ['brand', 'secondary', 'success', 'info', 'warning', 'error']
    }
  },

  app: {
    pageTransition: { name: 'page', mode: 'out-in' },
    layoutTransition: false
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1'
    }
  },

  compatibilityDate: '2025-01-01',

  icon: {
    clientBundle: {
      scan: true
    }
  }
})
