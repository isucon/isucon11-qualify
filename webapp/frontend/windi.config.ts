import { defineConfig } from 'vite-plugin-windicss'
import forms from 'windicss/plugin/forms'
import colors from 'windicss/colors'

const mainColor = '#FF6433'

export default defineConfig({
  plugins: [forms],
  theme: {
    extend: {
      colors: {
        main: mainColor
      },
      backgroundColor: {
        primary: colors.coolGray[100],
        secondary: colors.light[50],
        button: colors.blueGray[900],
        line: colors.trueGray[300],
        'accent-primary': mainColor,
        'status-info': colors.green[300],
        'status-warning': colors.yellow[300],
        'status-sitting': colors.orange[300],
        'status-critical': colors.red[300]
      },
      textColor: {
        primary: colors.trueGray[900],
        secondary: colors.trueGray[500],
        white: colors.light[50],
        'accent-primary': mainColor,
        'status-info': colors.green[900],
        'status-warning': colors.yellow[900],
        'status-sitting': colors.orange[900],
        'status-critical': colors.red[900]
      },
      borderColor: {
        primary: colors.trueGray[300],
        'accent-primary': mainColor
      },
      gridTemplateColumns: {
        list: 'repeat(auto-fill, minmax(12rem,1fr))',
        trend: '10rem 1fr'
      },
      keyframes: {
        scale: {
          '0%, 100%': { transform: 'scaley(1.0)' },
          '50%': { transform: 'scaley(0.4)' }
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
