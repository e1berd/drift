<script setup lang="ts">
import TopBar from '@/components/TopBar.vue'
import AgentSettings from '@/components/settings/AgentSettings.vue'
import InterfaceSettings from '@/components/settings/InterfaceSettings.vue'
import PasskeySettings from '@/components/settings/PasskeySettings.vue'
import { useI18n } from '@/shared/i18n'

const { t } = useI18n()
</script>

<template>
    <TopBar>
        <template #start
            ><span class="truncate font-medium">{{ t('settings.title') }}</span></template
        >
    </TopBar>
    <div class="min-h-0 flex-1 overflow-auto">
        <div class="motion-fade mx-auto flex max-w-[640px] flex-col gap-10 px-6 py-8 max-md:px-4">
            <InterfaceSettings />
            <AgentSettings />
            <PasskeySettings />
        </div>
    </div>
</template>
