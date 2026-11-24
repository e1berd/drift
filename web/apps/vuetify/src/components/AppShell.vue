<script setup lang="ts">
import { computed, ref } from 'vue'
import { useDisplay } from 'vuetify'

import NavDrawer from '@/components/nav/NavDrawer.vue'
import { provideShell } from '@drift/vue/composables/useShell'
import { useDriftStartup } from '@drift/vue/composables/useDriftStartup'

const { mobile } = useDisplay()
const drawerWanted = ref(true)
const drawer = computed({
    get: () => drawerWanted.value,
    set: (open: boolean) => (drawerWanted.value = open),
})

useDriftStartup()
provideShell({
    sidebarOpen: drawer,
    toggleSidebar: () => (drawerWanted.value = !drawerWanted.value),
    closeOnNarrow: () => {
        if (mobile.value) drawerWanted.value = false
    },
})
</script>

<template>
    <v-navigation-drawer v-model="drawer" :temporary="mobile" location="end" width="300">
        <NavDrawer />
    </v-navigation-drawer>
    <v-main class="shell-main">
        <slot />
    </v-main>
</template>

<style scoped>
.shell-main {
    block-size: 100dvh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}
</style>
