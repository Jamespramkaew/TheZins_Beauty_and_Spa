import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '')

  return {
    plugins: [react(), tailwindcss()],
    server: {
      // Same-origin /api in dev so the backend session cookie is sent (see API_PROXY_TARGET in .env)
      proxy: {
        '/api': { target: env.API_PROXY_TARGET || 'http://localhost:8000', changeOrigin: true },
      },
    },
  }
})
