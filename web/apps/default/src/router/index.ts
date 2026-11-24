import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: '/', name: 'new', component: () => import('@/views/NewChatView.vue') },
        {
            path: '/s/:sessionId/:tab(cli|files|git|terminal)?',
            name: 'session',
            component: () => import('@/views/SessionView.vue'),
        },
        {
            path: '/settings',
            name: 'settings',
            component: () => import('@/views/SettingsView.vue'),
        },
        { path: '/:pathMatch(.*)*', redirect: '/' },
    ],
})

export default router
