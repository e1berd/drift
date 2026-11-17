<script setup lang="ts">
import { Switch, SwitchThumb } from '@ui-sdk/vue/switch'

import AgentLogo from '@/components/AgentLogo.vue'
import { useAgents } from '@/composables/useAgents'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const agents = useAgents()
</script>

<template>
    <section class="flex flex-col gap-4">
        <header class="flex flex-col gap-1">
            <h2 class="m-0 text-base font-semibold">{{ t('settings.agents') }}</h2>
            <p class="m-0 text-muted">{{ t('settings.agentsLead') }}</p>
        </header>
        <div v-for="agent in agents.all" :key="agent.id" class="flex items-center gap-3">
            <AgentLogo :agent="agent.id" :size="20" decorative />
            <span class="flex-1"
                >{{ agent.name }} <code class="text-muted">{{ agent.cli }}</code></span
            >
            <Switch
                :checked="agents.isConnected(agent.id)"
                :aria-label="agent.name"
                @checked-change="(on: boolean) => agents.toggle(agent.id, on)"
            >
                <SwitchThumb />
            </Switch>
        </div>
    </section>
</template>
