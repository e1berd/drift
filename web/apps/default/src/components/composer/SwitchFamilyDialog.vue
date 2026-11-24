<script setup lang="ts">
import {
    AlertDialog,
    AlertDialogBackdrop,
    AlertDialogClose,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogTitle,
    createConfirm,
} from '@ui-sdk/vue/alert-dialog'
import { useI18n } from 'vue-i18n'

import type { AgentId } from '@drift/sdk/agents'

import AgentLogo from '@/components/AgentLogo.vue'
import { AGENTS } from '@drift/sdk/agents'

interface SwitchQuestion {
    agent: AgentId
}

const { t } = useI18n()
const confirm = createConfirm<SwitchQuestion>()
const SwitchDialogContent = AlertDialogContent<SwitchQuestion>

defineExpose({ ask: (agent: AgentId) => confirm({ agent }) })
</script>

<template>
    <AlertDialog :controller="confirm.controller">
        <AlertDialogBackdrop />
        <SwitchDialogContent v-slot="{ state }">
            <div v-if="state.payload" class="flex flex-col gap-2">
                <AlertDialogTitle class="flex items-center gap-2">
                    <AgentLogo :agent="state.payload.agent" :size="20" decorative />
                    {{
                        t('composer.switchFamily.title', {
                            agent: AGENTS[state.payload.agent].name,
                        })
                    }}
                </AlertDialogTitle>
                <AlertDialogDescription>{{
                    t('composer.switchFamily.body')
                }}</AlertDialogDescription>
            </div>
            <div class="flex justify-end gap-2">
                <AlertDialogClose>{{ t('composer.switchFamily.cancel') }}</AlertDialogClose>
                <AlertDialogClose data-variant="solid" value="confirm">{{
                    t('composer.switchFamily.confirm')
                }}</AlertDialogClose>
            </div>
        </SwitchDialogContent>
    </AlertDialog>
</template>
