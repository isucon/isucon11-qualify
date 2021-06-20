import { defineConfig } from 'vite'
import reactRefresh from '@vitejs/plugin-react-refresh'
import WindiCSS from 'vite-plugin-windicss'
import path from 'path'

const srcPath = path.resolve(__dirname, 'src').replace(/\\/g, '/')

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '/@': srcPath
    }
  },
  plugins: [reactRefresh(), WindiCSS()],
  esbuild: {
    jsxInject: `import React from 'react'`
  }
})
