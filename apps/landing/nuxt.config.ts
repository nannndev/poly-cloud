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
      routes: ['/', '/contributors', '/support'],
      crawlLinks: true,
      // Data GitHub diambil saat build. Bila API sedang membatasi kuota, halaman
      // harus tetap terbit dengan nilai bawaannya alih-alih menggagalkan deploy.
      failOnError: false
    }
  },

  app: {
    head: {
      htmlAttrs: { lang: 'en', class: 'dark' },
      link: [
        { rel: 'icon', href: '/favicon.svg', type: 'image/svg+xml' },
        { rel: 'apple-touch-icon', href: '/og-icon.png' },
        // Outfit adalah tipografi merek. Tanpa dimuat, --font-sans akan jatuh
        // diam-diam ke system-ui dan judulnya tak lagi cocok dengan lockup.
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700&display=swap'
        }
      ]
    }
  }
})
