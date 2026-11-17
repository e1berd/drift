<script setup lang="ts">
import { computed, ref, useTemplateRef } from 'vue'
import {
    PhArrowLeft,
    PhArrowRight,
    PhArrowUp,
    PhDeviceMobile,
    PhPaperclip,
} from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { FileInput, FileInputTrigger } from '@ui-sdk/vue/file-input'

import AttachmentChip from '@/components/composer/AttachmentChip.vue'
import ModeSelect from '@/components/composer/ModeSelect.vue'
import ModelSelect from '@/components/composer/ModelSelect.vue'
import PromptEditor from '@/components/composer/PromptEditor.vue'
import ImagePreview from '@/components/composer/ImagePreview.vue'
import { useDraft } from '@/composables/useDrafts'
import type { AgentId } from '@/shared/agents'
import { useI18n } from 'vue-i18n'

const props = withDefaults(
    defineProps<{
        draftKey: string
        agent: AgentId
        placeholder: string
        history?: string[]
        busy?: boolean
        showHistory?: boolean
    }>(),
    { history: () => [], busy: false, showHistory: true },
)

const emit = defineEmits<{ send: [payload: { text: string; files: File[] }] }>()
const { t } = useI18n()

const draft = useDraft(() => props.draftKey)
const files = ref<File[]>([])
const historyIndex = ref(-1)
const preview = ref<{ url: string; name: string } | null>(null)
const editor = useTemplateRef<InstanceType<typeof PromptEditor>>('editor')

const canSend = computed(() => !props.busy && (!!draft.text.value.trim() || files.value.length > 0))

function setText(value: string) {
    draft.set(value)
    draft.external.value++
}

function send() {
    if (!canSend.value) return
    emit('send', { text: draft.text.value.trim(), files: files.value })
    files.value = []
    historyIndex.value = -1
    draft.clear()
}

function back() {
    if (!props.history.length) return
    const index =
        historyIndex.value === -1 ? props.history.length - 1 : Math.max(0, historyIndex.value - 1)
    historyIndex.value = index
    setText(props.history.at(index) ?? '')
}

function forward() {
    if (historyIndex.value === -1) return
    const index = historyIndex.value + 1
    if (index >= props.history.length) {
        historyIndex.value = -1
        setText('')
        return
    }
    historyIndex.value = index
    setText(props.history.at(index) ?? '')
}

function addFiles(added: File[]) {
    files.value = [...files.value, ...added]
}

function onInput(value: string) {
    historyIndex.value = -1
    draft.set(value)
}

defineExpose({ focus: () => editor.value?.focus() })
</script>

<template>
    <FileInput v-model:value="files" class="max-w-none" multiple :max-size="25 * 1024 * 1024">
        <div class="composer" :data-remote="draft.remote.value.typing || undefined">
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

            <div v-if="files.length" v-auto-animate class="flex flex-wrap gap-2 px-3 pt-3">
                <AttachmentChip
                    v-for="(file, index) in files"
                    :key="`${file.name}-${file.size}-${file.lastModified}`"
                    :file="file"
                    removable
                    @remove="files.splice(index, 1)"
                    @open="(url) => (preview = { url, name: file.name })"
                />
            </div>

            <PromptEditor
                ref="editor"
                :model-value="draft.text.value"
                :revision="draft.external.value"
                :placeholder="placeholder"
                :label="t('composer.label')"
                @update:model-value="onInput"
                @submit="send"
                @history-back="back"
                @files="addFiles"
            />

            <div class="flex flex-wrap items-center justify-between gap-2 p-2">
                <div class="flex items-center gap-1">
                    <FileInputTrigger
                        :aria-label="t('composer.attach')"
                        :title="t('composer.attach')"
                    >
                        <PhPaperclip />
                    </FileInputTrigger>
                    <slot name="start">
                        <template v-if="showHistory">
                            <Button
                                :disabled="!history.length || historyIndex === 0"
                                :aria-label="t('composer.historyBack')"
                                :title="t('composer.historyBack')"
                                @click="back"
                            >
                                <PhArrowLeft />
                            </Button>
                            <Button
                                :disabled="historyIndex === -1"
                                :aria-label="t('composer.historyForward')"
                                :title="t('composer.historyForward')"
                                @click="forward"
                            >
                                <PhArrowRight />
                            </Button>
                        </template>
                    </slot>
                </div>
                <div class="flex items-center gap-1">
                    <ModelSelect :agent="agent" />
                    <ModeSelect />
                    <slot name="action">
                        <Button
                            data-variant="solid"
                            :disabled="!canSend"
                            :aria-label="t('composer.send')"
                            :title="t('composer.send')"
                            @click="send"
                        >
                            <PhArrowUp />
                        </Button>
                    </slot>
                </div>
            </div>
        </div>
    </FileInput>
    <ImagePreview v-if="preview" :url="preview.url" :name="preview.name" @close="preview = null" />
</template>

<style scoped>
.composer {
    width: 100%;
    max-width: 780px;
    margin-inline: auto;
    border: 1px solid var(--ui-input);
    border-radius: var(--ui-radius);
    background: var(--ui-background);

    &:focus-within {
        outline: 4px solid var(--ui-ring);
        border-color: var(--ui-ring-border);
    }

    &[data-remote] {
        border-style: dashed;
    }
}
</style>
