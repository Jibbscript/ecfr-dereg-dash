export default defineNuxtConfig({
  app: {
    head: {
      title: 'eCFR Dashboard'
    }
  },
  modules: ['nuxt-uswds'],
  nuxtUswds: {
    autoImportComponents: true,
    componentPrefix: 'Usa',
  },
  devtools: { enabled: true },
  nitro: {
    prerender: {
      crawlLinks: true
    },
    routeRules: {
      '/api/**': { proxy: 'http://localhost:8080/api/**' }
    }
  },
  vite: {
    build: {
      rollupOptions: {
        output: {
          manualChunks(id: string) {
            if (id.includes('node_modules')) {
              return 'vendor';
            }
          }
        }
      }
    }
  }
})
