import '@/styles/base.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { setupI18n } from './i18n'
import { applyColorScheme } from '@/shared/preferences'

applyColorScheme()

const i18n = setupI18n()
document.documentElement.lang = i18n.global.locale.value

const app = createApp(App)

app.use(router)
app.use(i18n)

app.mount('#app')
