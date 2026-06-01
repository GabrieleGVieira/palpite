import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'node:path';

export default defineConfig({
  base: process.env.VITE_BASE_PATH ?? '/',
  build: {
    rollupOptions: {
      input: {
        accountDeletion: resolve(__dirname, 'account-deletion/index.html'),
        main: resolve(__dirname, 'index.html'),
        privacy: resolve(__dirname, 'privacy/index.html'),
        terms: resolve(__dirname, 'terms/index.html'),
      },
    },
  },
  plugins: [react()],
});
