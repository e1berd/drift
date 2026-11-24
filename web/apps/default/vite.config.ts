import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'

export default defineConfig({
    plugins: [
        tailwindcss(),
        vue(),
        VueI18nPlugin({
            include: [
                fileURLToPath(new URL('./src/locales/**', import.meta.url)),
                fileURLToPath(new URL('../../packages/vue/src/locales/**', import.meta.url)),
            ],
        }),
        vueDevTools(),
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    server: {
        port: 5173,
        strictPort: true,
    },
    build: {
        outDir: fileURLToPath(new URL('../../../app/internal/frontend/dist', import.meta.url)),
        emptyOutDir: true,
        manifest: true,
        rolldownOptions: {
            input: {
                app: fileURLToPath(new URL('./index.html', import.meta.url)),
                auth: fileURLToPath(new URL('./src/auth/main.ts', import.meta.url)),
            },
        },
    },
})
