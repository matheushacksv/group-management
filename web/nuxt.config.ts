// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxt/ui', 'nuxt-auth-utils'],
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    // server only, never sent to the browser
    apiBase: '', // NUXT_API_BASE=http://localhost:8088
    apiKey: '', // NUXT_API_KEY=same value as API_KEY in the Go API
    adminEmail: '', // NUXT_ADMIN_EMAIL
    adminPasswordHash: '', // NUXT_ADMIN_PASSWORD_HASH, generate with: bun scripts/hash-password.ts <password>
    public: {
      sessionId: '' // NUXT_PUBLIC_SESSION_ID, OpenWA session
    }
  }
})
