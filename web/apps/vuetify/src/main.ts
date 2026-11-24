import 'vuetify/styles'
import '@/styles/markdown.css'

import { createApp } from 'vue'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'

import App from './App.vue'
import router from './router'
import { setupI18n } from './i18n'
import { createDriftVuetify } from './shared/vuetify'
import { snackbarNotifier } from './shared/snackbarNotifier'
import { createNotifierPlugin } from '@drift/vue/notifier'

const i18n = setupI18n()
document.documentElement.lang = i18n.global.locale.value

const queryClient = new QueryClient({
    defaultOptions: {
        queries: { staleTime: 60_000, refetchOnWindowFocus: false, retry: 1 },
    },
})

createApp(App)
    .use(createNotifierPlugin(snackbarNotifier))
    .use(createDriftVuetify())
    .use(router)
    .use(i18n)
    .use(VueQueryPlugin, { queryClient })
    .mount('#app')
