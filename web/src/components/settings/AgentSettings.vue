<script setup lang="ts">
import { Switch, SwitchThumb } from '@ui-sdk/vue/switch'

import AgentLogo from '@/components/AgentLogo.vue'
import { useAgents } from '@/composables/useAgents'
import { useI18n } from '@/shared/i18n'

const { t } = useI18n()
const agents = useAgents()
</script>

<template>
    <section class="flex flex-col gap-4">
        <header class="flex flex-col gap-1">
            <h2 class="m-0 text-base font-semibold">{{ t('settings.agents') }}</h2>
            <p class="m-0 text-sm text-muted">{{ t('settings.agentsLead') }}</p>
        </header>
        <ul
            class="m-0 flex list-none flex-col overflow-hidden rounded-xl border border-line bg-surface p-0"
        >
            <li
                v-for="a in agents.all"
                :key="a.id"
                class="flex items-center gap-3 px-3.5 py-3 not-first:border-t not-first:border-line"
            >
                <AgentLogo :agent="a.id" :size="22" decorative />
                <span class="flex-1 text-[13.5px]">
                    {{ a.name }} <code class="ml-1 text-[11.5px] text-faint">{{ a.cli }}</code>
                </span>
                <Switch
                    class="agent-switch"
                    :checked="agents.isConnected(a.id)"
                    :aria-label="a.name"
                    @checked-change="(on: boolean) => agents.toggle(a.id, on)"
                >
                    <SwitchThumb class="agent-switch-thumb" />
                </Switch>
            </li>
        </ul>
    </section>
</template>

<style>
.agent-switch {
    position: relative;
    width: 36px;
    height: 20px;
    flex: none;
    padding: 0;
    border: 0;
    border-radius: 999px;
    background: var(--surface-3);
    cursor: pointer;
    transition: background var(--dur-2);

    &[data-checked] {
        background: var(--accent);
    }
}

.agent-switch-thumb {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.3);
    transition: translate var(--dur-3) var(--ease-out);

    &[data-checked] {
        translate: 16px 0;
    }
}
</style>
