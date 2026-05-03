import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return
          }

          if (id.includes('/node_modules/vue/') || id.includes('/node_modules/@vue/')) {
            return 'vue'
          }

          if (
            id.includes('/node_modules/element-plus/') ||
            id.includes('/node_modules/@element-plus/') ||
            id.includes('/node_modules/@popperjs/') ||
            id.includes('/node_modules/dayjs/')
          ) {
            return 'element-plus'
          }

          if (id.includes('/node_modules/axios/')) {
            return 'axios'
          }

          return 'vendor'
        }
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    allowedHosts: ['stat.icehe.life'],
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true
      }
    }
  }
})
