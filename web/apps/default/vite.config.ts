import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'

import { driftAppConfig, driftLocales } from '../../vite.drift.ts'

export default defineConfig({
    ...driftAppConfig(import.meta.url, 'default', {
        app: './index.html',
        auth: './src/auth/main.ts',
    }),
    plugins: [
        tailwindcss(),
        vue(),
        VueI18nPlugin({
            include: [fileURLToPath(new URL('./src/locales/**', import.meta.url)), driftLocales],
        }),
        vueDevTools(),
    ],
    server: {
        port: 5173,
        strictPort: true,
    },
})
