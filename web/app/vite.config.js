import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  build: {
    outDir: path.resolve(__dirname, '/web/dist'), // Output to web/dist
    emptyOutDir: true,                            // Clean old builds
  },
  server: {
    port: 5173,
    proxy: {
      // Forward requests to backend during dev
      '/ui': 'http://localhost:8000',
    },
  },
})
