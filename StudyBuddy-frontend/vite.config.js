import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: true,
    proxy: {
      '/api/v1/courses': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/users': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/interests': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/availability': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/matching': {
        target: 'http://localhost:8084',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/groups': {
        target: 'http://localhost:8085',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/reviews': {
        target: 'http://localhost:8086',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/points': {
        target: 'http://localhost:8087',
        changeOrigin: true,
        secure: false,
      },
      '/api/v1/sessions': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        secure: false,
      },
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
