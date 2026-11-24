<script setup lang="ts">
import { computed, nextTick, ref, useTemplateRef, watch } from 'vue'
import { PhArrowLeft, PhArrowRight, PhDeviceMobile, PhPaperclip } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { FileInput, FileInputTrigger } from '@ui-sdk/vue/file-input'
import { RouterLink } from 'vue-router'

import ComposerAttachments from '@/components/composer/ComposerAttachments.vue'
import ComposerSendControls from '@/components/composer/ComposerSendControls.vue'
import ModeSelect from '@/components/composer/ModeSelect.vue'
import ModelSelect from '@/components/composer/ModelSelect.vue'
import PromptEditor from '@/components/composer/PromptEditor.vue'
import UsageIndicator from '@/components/composer/usage/UsageIndicator.vue'
import ImagePreview from '@/components/composer/ImagePreview.vue'
import { usePromptComposer } from '@drift/vue/chat/usePromptComposer'
import type { AgentId } from '@drift/sdk/agents'
import { AGENTS } from '@drift/sdk/agents'
import type { SentPrompt } from '@drift/sdk/attachments'
import { useI18n } from 'vue-i18n'

const props = withDefaults(
    defineProps<{
        draftKey: string
        submit: (payload: SentPrompt) => Promise<unknown>
        agent: AgentId
        sessionId?: string
        placeholder: string
        history?: string[]
        busy?: boolean
        running?: boolean
        stopping?: boolean
        showHistory?: boolean
        autofocus?: boolean
    }>(),
    {
        history: () => [],
        busy: false,
        running: false,
        stopping: false,
        showHistory: true,
        autofocus: false,
    },
)

const emit = defineEmits<{
    switchAgent: [choice: { agent: AgentId; model: string }]
    stop: []
}>()
const { t } = useI18n()
const composer = usePromptComposer({
    draftKey: () => props.draftKey,
    agent: () => props.agent,
    submit: (prompt) => props.submit(prompt),
    history: () => props.history,
    busy: () => props.busy,
})
const { draft, uploads, historyIndex, unavailable, canSend, send, back, forward } = composer
const onInput = composer.type
const noFiles: File[] = []
const preview = ref<{ url: string; name: string } | null>(null)
const editor = useTemplateRef<InstanceType<typeof PromptEditor>>('editor')

const sendTitle = computed(() => {
    if (uploads.uploading.value) return t('composer.waitForUploads')
    return props.running ? t('composer.queue') : t('composer.send')
})

watch(
    [() => props.draftKey, unavailable],
    async () => {
        if (!props.autofocus || unavailable.value) return
        await nextTick()
        editor.value?.focus()
    },
    { immediate: true, flush: 'post' },
)

defineExpose({ focus: () => editor.value?.focus() })
</script>

<template>
    <FileInput
        :value="noFiles"
        class="max-w-none"
        multiple
        :disabled="unavailable"
        @update:value="uploads.add"
    >
        <div
            class="composer"
            :data-remote="draft.remote.value.typing || undefined"
            :data-unavailable="unavailable || undefined"
        >
            <p v-if="unavailable" class="motion-fade m-0 px-3 pt-3 text-muted">
                {{ t('composer.unavailable', { agent: AGENTS[agent].name }) }}
                <RouterLink to="/settings" class="text-fg underline underline-offset-4">{{
                    t('composer.setUp')
                }}</RouterLink>
            </p>
            <div
                v-if="draft.remote.value.typing || draft.remote.value.fromOther"
                class="status-tag motion-fade px-3 pt-2"
            >
                <PhDeviceMobile />
                <span class="flex-1">
                    {{
                        draft.remote.value.typing
                            ? t('composer.remoteTyping')
                            : t('composer.remoteDraft')
                    }}
                    <span v-if="draft.remote.value.typing" class="typing-dots"
                        ><i /><i /><i
                    /></span>
                </span>
                <Button v-if="!draft.remote.value.typing" @click="draft.clear()">{{
                    t('composer.clear')
                }}</Button>
            </div>

            <ComposerAttachments
                :uploads="uploads"
                @open="(url, name) => (preview = { url, name })"
            />

            <PromptEditor
                ref="editor"
                :model-value="draft.text.value"
                :revision="draft.external.value"
                :placeholder="placeholder"
                :label="t('composer.label')"
                :disabled="unavailable"
                @update:model-value="onInput"
                @submit="send"
                @history-back="back"
                @files="uploads.add"
            />

            <div class="composer-bar">
                <FileInputTrigger
                    :aria-label="t('composer.attach')"
                    v-tooltip="t('composer.attach')"
                >
                    <PhPaperclip />
                </FileInputTrigger>
                <ModelSelect
                    :agent="agent"
                    :session-id="sessionId"
                    @switch-agent="(choice) => emit('switchAgent', choice)"
                />
                <ModeSelect :agent="agent" />
                <slot name="start">
                    <div v-if="showHistory" class="composer-history">
                        <Button
                            :disabled="unavailable || !history.length || historyIndex === 0"
                            :aria-label="t('composer.historyBack')"
                            v-tooltip="t('composer.historyBack')"
                            @click="back"
                        >
                            <PhArrowLeft />
                        </Button>
                        <Button
                            :disabled="historyIndex === -1"
                            :aria-label="t('composer.historyForward')"
                            v-tooltip="t('composer.historyForward')"
                            @click="forward"
                        >
                            <PhArrowRight />
                        </Button>
                    </div>
                </slot>
                <span class="composer-bar-spacer" />
                <UsageIndicator :agent="agent" :session-id="sessionId" :running="running" />
                <ComposerSendControls
                    :running="running"
                    :stopping="stopping"
                    :can-send="canSend"
                    :send-title="sendTitle"
                    @send="send"
                    @stop="emit('stop')"
                />
            </div>
        </div>
    </FileInput>
    <ImagePreview v-if="preview" :url="preview.url" :name="preview.name" @close="preview = null" />
</template>

<style scoped>
.composer {
    container: composer / inline-size;
    width: 100%;
    max-width: 780px;
    margin-inline: auto;
    border: 1px solid var(--ui-input);
    border-radius: var(--ui-radius);
    background: var(--ui-background);

    &:focus-within {
        border-color: var(--ui-ring-border);
    }

    &[data-remote] {
        border-style: dashed;
    }

    &[data-unavailable] {
        background: var(--ui-muted);
    }
}
</style>
