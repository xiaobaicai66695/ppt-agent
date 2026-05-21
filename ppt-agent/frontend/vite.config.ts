import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // SSE (text/event-stream) 需要禁用代理缓冲，以支持分块传输
        configure: (proxy) => {
          proxy.on('proxyRes', (proxyRes) => {
            // 删除可能导致代理缓冲的 header
            delete proxyRes.headers['content-length'];
            delete proxyRes.headers['content-encoding'];
          });
        },
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
