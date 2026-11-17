<script setup lang="ts">
import { computed, ref, useTemplateRef } from 'vue'
import {
    PhArrowLeft,
    PhArrowRight,
    PhArrowUp,
    PhDeviceMobile,
    PhPaperclip,
} from '@phosphor-icons/vue'
import { FileInput, FileInputDropzone, FileInputTrigger } from '@ui-sdk/vue/file-input'

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
    const i =
        historyIndex.value === -1 ? props.history.length - 1 : Math.max(0, historyIndex.value - 1)
    historyIndex.value = i
    setText(props.history[i] ?? '')
}

function forward() {
    if (historyIndex.value === -1) return
    const i = historyIndex.value + 1
    if (i >= props.history.length) {
        historyIndex.value = -1
        setText('')
        return
    }
    historyIndex.value = i
    setText(props.history[i] ?? '')
}

function onInput(value: string) {
    historyIndex.value = -1
    draft.set(value)
}

defineExpose({ focus: () => editor.value?.focus() })
</script>

<template>
    <FileInput v-model:value="files" class="block w-full" multiple :max-size="25 * 1024 * 1024">
        <FileInputDropzone class="composer" :data-remote="draft.remote.value.typing || undefined">
            <div
                v-if="draft.remote.value.typing || draft.remote.value.fromOther"
                class="remote motion-fade"
            >
                <PhDeviceMobile :size="13" />
                <span class="flex-1">
                    {{
                        draft.remote.value.typing
                            ? t('composer.remoteTyping')
                            : t('composer.remoteDraft')
                    }}
                    <span v-if="draft.remote.value.typing" class="typing-dots ml-0.5"
                        ><i /><i /><i
                    /></span>
                </span>
                <button
                    v-if="!draft.remote.value.typing"
                    type="button"
                    class="link-btn"
                    @click="draft.clear()"
                >
                    {{ t('composer.clear') }}
                </button>
            </div>

            <div v-if="files.length" v-auto-animate class="flex flex-wrap gap-2 px-3 pt-3">
                <AttachmentChip
                    v-for="(f, i) in files"
                    :key="`${f.name}-${f.size}-${f.lastModified}`"
                    :file="f"
                    removable
                    @remove="files.splice(i, 1)"
                    @open="(url) => (preview = { url, name: f.name })"
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
            />

            <div class="bar">
                <div class="flex items-center gap-0.5">
                    <FileInputTrigger
                        class="icon-btn"
                        :aria-label="t('composer.attach')"
                        :title="t('composer.attach')"
                    >
                        <PhPaperclip :size="18" />
                    </FileInputTrigger>
                    <span class="mx-1 h-4 w-px bg-line-strong" />
                    <slot name="start">
                        <template v-if="showHistory">
                            <button
                                type="button"
                                class="icon-btn"
                                :disabled="!history.length || historyIndex === 0"
                                :aria-label="t('composer.historyBack')"
                                :title="t('composer.historyBack')"
                                @click="back"
                            >
                                <PhArrowLeft :size="18" />
                            </button>
                            <button
                                type="button"
                                class="icon-btn"
                                :disabled="historyIndex === -1"
                                :aria-label="t('composer.historyForward')"
                                :title="t('composer.historyForward')"
                                @click="forward"
                            >
                                <PhArrowRight :size="18" />
                            </button>
                        </template>
                    </slot>
                </div>
                <div class="flex items-center gap-1">
                    <ModelSelect :agent="agent" />
                    <ModeSelect />
                    <slot name="action">
                        <button
                            type="button"
                            class="send"
                            :disabled="!canSend"
                            :aria-label="t('composer.send')"
                            :title="t('composer.send')"
                            @click="send"
                        >
                            <PhArrowUp :size="17" weight="bold" />
                        </button>
                    </slot>
                </div>
            </div>
        </FileInputDropzone>
    </FileInput>
    <ImagePreview v-if="preview" :url="preview.url" :name="preview.name" @close="preview = null" />
</template>

<style scoped>
.composer {
    position: relative;
    width: 100%;
    max-width: 780px;
    margin: 0 auto;
    border: 1px solid var(--border-strong);
    border-radius: 16px;
    background: var(--surface);
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.12);
    transition:
        border-color var(--dur-2),
        box-shadow var(--dur-2);

    &:focus-within {
        border-color: var(--accent);
        box-shadow: 0 0 0 4px var(--accent-soft);
    }

    &[data-remote] {
        border-style: dashed;
    }

    &[data-dragging] {
        border-color: var(--accent);
        border-style: dashed;
        background: var(--accent-soft);
    }
}

.remote {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px 0 14px;
    color: var(--ui-muted-foreground);
    font-size: 12px;
}

.link-btn {
    padding: 2px 4px;
    border: 0;
    border-radius: 5px;
    background: none;
    color: var(--faint);
    font-size: 12px;
    cursor: pointer;

    &:hover {
        background: var(--surface-2);
        color: var(--ui-foreground);
    }
}

.bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 6px 8px 8px;
}

.send {
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border: 0;
    border-radius: 9px;
    background: var(--accent);
    color: var(--accent-ink);
    cursor: pointer;
    transition:
        background var(--dur-2),
        color var(--dur-2),
        transform var(--dur-1);

    &:disabled {
        background: var(--surface-3);
        color: var(--faint);
        cursor: not-allowed;
    }

    &:not(:disabled):active {
        transform: scale(0.93);
    }
}
</style>
