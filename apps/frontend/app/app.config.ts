export default defineAppConfig({
  ui: {
    colors: {
      // 'brand' diturunkan dari lambang Arch; skalanya didefinisikan di
      // app/assets/css/main.css agar Tailwind dan Nuxt UI memakai nilai sama.
      primary: 'brand',
      neutral: 'slate'
    }
  }
})
