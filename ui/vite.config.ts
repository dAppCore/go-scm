// SPDX-License-Identifier: EUPL-1.2

import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'CoreScm',
      fileName: 'core-scm',
      formats: ['es'],
    },
    outDir: resolve(__dirname, '../pkg/api/ui/dist'),
    emptyOutDir: true,
    rollupOptions: {
      output: {
        entryFileNames: 'core-scm.js',
      },
    },
  },
});
