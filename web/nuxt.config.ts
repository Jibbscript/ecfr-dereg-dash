export default defineNuxtConfig({
  app: {
    head: {
      title: 'eCFR Deregulation Dashboard',
      titleTemplate: '%s | eCFR Dashboard',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        {
          name: 'description',
          content: 'Explore and analyze the complexity of federal regulations. Track regulatory burden using the RSCS metric and identify opportunities for simplification.'
        },
        { name: 'theme-color', content: '#1a4480' },
        // Open Graph
        { property: 'og:type', content: 'website' },
        { property: 'og:title', content: 'eCFR Deregulation Dashboard' },
        {
          property: 'og:description',
          content: 'Analytics dashboard for exploring the Electronic Code of Federal Regulations and measuring regulatory complexity.'
        },
        { property: 'og:site_name', content: 'eCFR Dashboard' },
        // Twitter Card
        { name: 'twitter:card', content: 'summary_large_image' },
        { name: 'twitter:title', content: 'eCFR Deregulation Dashboard' },
        {
          name: 'twitter:description',
          content: 'Analytics dashboard for exploring federal regulations and measuring regulatory complexity.'
        },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' },
      ],
      htmlAttrs: {
        lang: 'en'
      }
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
      '/api/**': { proxy: `${process.env.API_BASE_URL || 'http://localhost:8080'}/api/**` }
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
