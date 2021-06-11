import { defineConfig } from 'vite'
import reactRefresh from '@vitejs/plugin-react-refresh'
import WindiCSS from 'vite-plugin-windicss'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [reactRefresh(), WindiCSS()],
  esbuild: {
    jsxInject: `import React from 'react'`
  }
})
