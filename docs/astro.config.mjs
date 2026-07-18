import { defineConfig } from 'astro/config'
import tailwind from '@astrojs/tailwind'

export default defineConfig({
  site: 'https://zzabolazza.github.io',
  base: '/Facade/',
  integrations: [tailwind()],
})
