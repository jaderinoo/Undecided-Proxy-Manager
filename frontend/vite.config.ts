import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 6071,
    host: true,
  },
  test: {
    setupFiles: ['./src/test/setup.ts'],
  },
});
