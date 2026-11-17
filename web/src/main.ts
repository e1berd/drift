import '@/styles/base.css'

import { createApp } from 'vue'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { createVAutoAnimate } from '@formkit/auto-animate/vue'

import App from './App.vue'
import router from './router'
import { setupI18n } from './i18n'
import { applyColorScheme } from '@/shared/preferences'

applyColorScheme()

const i18n = setupI18n()
document.documentElement.lang = i18n.global.locale.value

const queryClient = new QueryClient({
    defaultOptions: {
        queries: { staleTime: 60_000, refetchOnWindowFocus: false, retry: 1 },
    },
})

const app = createApp(App)

app.use(router)
app.use(i18n)
app.use(VueQueryPlugin, { queryClient })
app.directive(
    'auto-animate',
    createVAutoAnimate({ duration: 320, easing: 'cubic-bezier(0.65, 0, 0.35, 1)' }),
)

app.mount('#app')
