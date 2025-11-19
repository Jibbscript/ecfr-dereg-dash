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
    vueUswds: {
      routerComponentName: 'NuxtLink'
    }
  },
  devtools: { enabled: true },
  nitro: {
    prerender: {
      crawlLinks: true
    },
    routeRules: {
      '/api/**': { proxy: 'http://localhost:8080/api/**' }
    }
  }
})
