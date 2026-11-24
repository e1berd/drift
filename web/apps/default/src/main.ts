import '@/styles/base.css'

import { createApp } from 'vue'
import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query'
import { createVAutoAnimate } from '@formkit/auto-animate/vue'

import App from './App.vue'
import router from './router'
import { setupI18n } from './i18n'
import { vTooltip } from '@/shared/tooltip/tooltip'
import { applyColorScheme } from '@/shared/preferences'
import { toastNotifier } from '@/shared/notifier'
import { createNotifierPlugin } from '@drift/vue/notifier'

applyColorScheme()

const i18n = setupI18n()
document.documentElement.lang = i18n.global.locale.value

const queryClient = new QueryClient({
    defaultOptions: {
        queries: { staleTime: 60_000, refetchOnWindowFocus: false, retry: 1 },
    },
})

const app = createApp(App)

app.use(createNotifierPlugin(toastNotifier))
app.use(router)
app.use(i18n)
app.use(VueQueryPlugin, { queryClient })
app.directive('tooltip', vTooltip)
app.directive(
    'auto-animate',
    createVAutoAnimate({ duration: 320, easing: 'cubic-bezier(0.65, 0, 0.35, 1)' }),
)

app.mount('#app')
