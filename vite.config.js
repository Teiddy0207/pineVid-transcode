import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig(() => {
  require('dotenv').config({ path: `./.env` });

  return defineConfig({
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, 'src'),
      },
    },
    server: {
      open: true,
      proxy: {
        '/api': {
          target: process.env.DOMAIN_API,
          changeOrigin: true,
          logLevel: 'info'
        },
        '/public': {
          target: process.env.DOMAIN_API,
          changeOrigin: true,
          logLevel: 'info'
        },
      },
    },
    build: {
      outDir: './layout',
    },
    publicDir: false
  })
})
