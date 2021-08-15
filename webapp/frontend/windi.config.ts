import { defineConfig } from 'vite-plugin-windicss'
import forms from 'windicss/plugin/forms'

export default defineConfig({
  plugins: [forms],
  theme: {
    backgroundColor: theme => ({
      ...theme('colors'),
      primary: '#F3F4F5',
      secondary: '#FFFFFF',
      teritary: '#F1F1F1',
      button: '#21394B',
      'accent-primary': '#FF6433',
      'status-info': '#94EFBC',
      'status-warning': '#FFEF5B',
      'status-sitting': '#FFBC7E',
      'status-critical': '#F69898'
    }),
    textColor: theme => ({
      ...theme('colors'),
      primary: '#241E12',
      secondary: '#6B6965',
      teritary: '#CCCCCC',
      error: '#CF1717',
      'white-primary': '#FFFFFF',
      'white-secondary': '#FFFFFF',
      'accent-primary': '#FF6433',
      'status-info': '#22623E',
      'status-warning': '#605910',
      'status-sitting': '#603A18',
      'status-critical': '#512424'
    }),
    borderColor: theme => ({
      ...theme('colors'),
      outline: '#E1E5E6',
      error: '#CF1717',
      'accent-primary': '#FF6433'
    }),
    extend: {
      gridTemplateColumns: {
        isus: 'repeat(auto-fill,minmax(10rem,1fr))',
        trend: '10rem 1fr'
      },
      keyframes: {
        scale: {
          '0%': { transform: 'scaley(1.0)' },
          '50%': { transform: 'scaley(0.4)' },
          '100%': { transform: 'scaley(1.0)' }
        }
      },
      animation: {
        loader0: 'scale 1s infinite cubic-bezier(0.2, 0.68, 0.18, 1.08)',
        loader1: 'scale 1s 0.1s infinite cubic-bezier(0.2, 0.68, 0.18, 1.08)',
        loader2: 'scale 1s 0.2s infinite cubic-bezier(0.2, 0.68, 0.18, 1.08)',
        loader3: 'scale 1s 0.3s infinite cubic-bezier(0.2, 0.68, 0.18, 1.08)',
        loader4: 'scale 1s 0.4s infinite cubic-bezier(0.2, 0.68, 0.18, 1.08)'
      }
    }
  }
})
