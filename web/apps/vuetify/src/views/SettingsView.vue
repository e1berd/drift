<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import ViewBar from '@/components/ViewBar.vue'
import AppearanceSettings from '@/components/settings/AppearanceSettings.vue'
import FrontendSettings from '@/components/settings/FrontendSettings.vue'

const { t } = useI18n()
</script>

<template>
    <ViewBar :title="t('nav.settings')" />
    <div class="settings">
        <AppearanceSettings />
        <FrontendSettings />
        <v-alert type="info" variant="tonal">{{ t('settings.moreInDefault') }}</v-alert>
    </div>
</template>

<style scoped>
.settings {
    display: flex;
    flex-direction: column;
    gap: 16px;
    inline-size: min(680px, 100%);
    margin-inline: auto;
    padding: 16px;
    overflow-y: auto;
}
</style>
