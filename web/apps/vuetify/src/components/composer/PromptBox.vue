<script setup lang="ts">
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import { mdiArrowUp, mdiPaperclip, mdiStop } from '@mdi/js'
import { useI18n } from 'vue-i18n'

import AgentControls from '@/components/composer/AgentControls.vue'
import ComposerFiles from '@/components/composer/ComposerFiles.vue'
import PromptEditor from '@/components/composer/PromptEditor.vue'
import { usePromptComposer } from '@drift/vue/chat/usePromptComposer'
import { AGENTS } from '@drift/sdk/agents'
import type { AgentId } from '@drift/sdk/agents'
import type { SentPrompt } from '@drift/sdk/attachments'

const props = withDefaults(
    defineProps<{
        draftKey: string
        submit: (prompt: SentPrompt) => Promise<unknown>
        agent: AgentId
        sessionId?: string
        placeholder: string
        history?: string[]
        busy?: boolean
        running?: boolean
        stopping?: boolean
        autofocus?: boolean
    }>(),
    { history: () => [], busy: false, running: false, stopping: false, autofocus: false },
)
const emit = defineEmits<{ stop: [] }>()

const { t } = useI18n()
const { draft, uploads, unavailable, canSend, send, back, forward, type } = usePromptComposer({
    draftKey: () => props.draftKey,
    agent: () => props.agent,
    submit: (prompt) => props.submit(prompt),
    history: () => props.history,
    busy: () => props.busy,
})
const editor = useTemplateRef<InstanceType<typeof PromptEditor>>('editor')
const filePicker = useTemplateRef<HTMLInputElement>('filePicker')

const sendLabel = computed(() => {
    if (uploads.uploading.value) return t('composer.waitForUploads')
    return props.running ? t('composer.queue') : t('composer.send')
})

function pickFiles(event: Event) {
    const input = event.target as HTMLInputElement
    uploads.add([...(input.files ?? [])])
    input.value = ''
}

watch(
    [() => props.draftKey, unavailable],
    async () => {
        if (!props.autofocus || unavailable.value) return
        await nextTick()
        editor.value?.focus()
    },
    { immediate: true, flush: 'post' },
)
</script>

<template>
    <v-card variant="outlined" class="prompt-box">
        <v-alert v-if="unavailable" type="warning" variant="tonal" density="compact" class="ma-2">
            {{ t('composer.notConnected', { agent: AGENTS[agent].name }) }}
            <router-link to="/settings">{{ t('composer.connect') }}</router-link>
        </v-alert>
        <ComposerFiles :uploads="uploads" />
        <PromptEditor
            ref="editor"
            :model-value="draft.text.value"
            :revision="draft.external.value"
            :placeholder="placeholder"
            :label="t('composer.label')"
            :disabled="unavailable"
            @update:model-value="type"
            @submit="send"
            @history-back="back"
            @files="uploads.add"
        />
        <v-card-actions class="flex-wrap ga-1">
            <input ref="filePicker" type="file" multiple hidden @change="pickFiles" />
            <v-btn
                :icon="mdiPaperclip"
                size="small"
                :aria-label="t('composer.attach')"
                :disabled="unavailable"
                @click="filePicker?.click()"
            />
            <slot name="start" />
            <AgentControls :agent="agent" :session-id="sessionId" />
            <v-spacer />
            <v-btn
                v-if="running"
                :icon="mdiStop"
                size="small"
                color="error"
                :loading="stopping"
                :aria-label="t('composer.stop')"
                @click="emit('stop')"
            />
            <v-btn
                :icon="mdiArrowUp"
                size="small"
                variant="flat"
                color="primary"
                :disabled="!canSend"
                :aria-label="sendLabel"
                :title="sendLabel"
                @click="send"
            />
        </v-card-actions>
    </v-card>
</template>

<style scoped>
.prompt-box {
    inline-size: 100%;
}
</style>
