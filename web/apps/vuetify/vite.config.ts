import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'

import { driftAppConfig, driftLocales } from '../../vite.drift.ts'

export default defineConfig({
    ...driftAppConfig(import.meta.url, 'vuetify', { app: './index.html' }),
    plugins: [
        vue(),
        vuetify({ autoImport: true }),
        VueI18nPlugin({
            include: [fileURLToPath(new URL('./src/locales/**', import.meta.url)), driftLocales],
        }),
    ],
    server: {
        port: 5174,
        strictPort: true,
    },
})
