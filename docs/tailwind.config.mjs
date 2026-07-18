/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,ts}'],
  theme: {
    extend: {
      colors: {
        surface: {
          DEFAULT: '#0f1117',
          card: '#1a1d27',
          border: 'rgba(255,255,255,0.06)',
        },
        accent: {
          from: '#4b6eff',
          to: '#7c5cff',
        },
        muted: '#6d7186',
        text: '#e7e8f2',
      },
    },
  },
  plugins: [],
}
