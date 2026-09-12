import tailwindcss from '@tailwindcss/vite'

// Landing page dibangun sepenuhnya statis: tak ada panggilan ke backend
// Poly Cloud, karena backend itu self-hosted dan tak dapat dijangkau dari
// halaman publik di Vercel.
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: false },

  css: ['~/assets/css/main.css'],

  vite: {
    plugins: [tailwindcss()]
  },

  nitro: {
    // Seluruh rute dirender saat build sehingga hasilnya berupa berkas statis.
    prerender: {
      routes: ['/'],
      crawlLinks: true
    }
  },

  app: {
    head: {
      htmlAttrs: { lang: 'en', class: 'dark' },
      link: [{ rel: 'icon', href: '/favicon.svg', type: 'image/svg+xml' }]
    }
  }
})
