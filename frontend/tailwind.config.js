/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        tavern: {
          50: '#f0f0f5',
          100: '#e0e0eb',
          200: '#c2c2d6',
          300: '#a3a3c2',
          400: '#8585ad',
          500: '#6b6b99',
          600: '#55557a',
          700: '#3f3f5c',
          800: '#2a2a3d',
          900: '#1a1a2e',
          950: '#0f0f1f',
        },
        accent: {
          DEFAULT: '#5b86e5',
          light: '#7ba0f0',
          dark: '#3d5fa8',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
    },
  },
  plugins: [],
}
