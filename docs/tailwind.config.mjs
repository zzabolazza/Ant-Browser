/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,ts}'],
  theme: {
    extend: {
      colors: {
        surface: {
          DEFAULT: '#070912',
          card: '#0f1421',
          elevated: '#151b2b',
          border: 'rgba(163,177,221,0.12)',
        },
        accent: {
          from: '#6d5dfc',
          to: '#25c9e8',
          blue: '#4f8cff',
        },
        muted: '#8992aa',
        text: '#f4f7ff',
      },
    },
  },
  plugins: [],
}
