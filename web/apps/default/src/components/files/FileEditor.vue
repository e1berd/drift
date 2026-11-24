<script setup lang="ts">
import { computed } from 'vue'
import { PhFloppyDisk } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useQueryClient } from '@tanstack/vue-query'
import { useI18n } from 'vue-i18n'

import type { FileContent } from '@drift/vue/api'

import CodeView from '@/components/files/CodeView.vue'
import { useSessionFiles } from '@drift/vue/composables/useSessionFiles'
import { readSessionFileQueryKey, useReadSessionFile, useUpdateSessionFile } from '@drift/vue/api'
import { useErrorNotice } from '@drift/vue/notifier'

const props = defineProps<{ sessionId: string }>()
const { t } = useI18n()
const qc = useQueryClient()
const { selected, drafts } = useSessionFiles()

const target = computed(() => ({
    path: { sessionId: props.sessionId },
    query: { path: selected.value ?? '' },
}))
const file = useReadSessionFile(
    { path: () => target.value.path, query: () => target.value.query },
    { query: { enabled: () => !!selected.value } },
)

const editable = computed(() => !!file.data.value && !file.data.value.truncated)
const draft = computed(() => (selected.value ? drafts.get(selected.value) : undefined))
const content = computed(() => draft.value ?? file.data.value?.content ?? '')
const dirty = computed(() => draft.value !== undefined && draft.value !== file.data.value?.content)
const lineCount = computed(() => content.value.split('\n').length)

const save = useUpdateSessionFile({
    mutation: {
        onSuccess: (saved: FileContent) => {
            qc.setQueryData(
                readSessionFileQueryKey({
                    path: { sessionId: props.sessionId },
                    query: { path: saved.path },
                }),
                saved,
            )
            if (drafts.get(saved.path) === saved.content) drafts.delete(saved.path)
        },
        onError: useErrorNotice(),
    },
})

function remember(text: string) {
    if (selected.value) drafts.set(selected.value, text)
}

function saveDraft() {
    if (!selected.value || !dirty.value || save.isPending.value) return
    save.mutate({ ...target.value, body: { content: content.value } })
}
</script>

<template>
    <section class="flex min-h-0 min-w-0 flex-col">
        <div class="editor-head">
            <b class="truncate font-mono font-medium text-fg">{{ selected ?? t('files.pick') }}</b>
            <span v-if="dirty" class="shrink-0" v-tooltip="t('unsaved')" aria-hidden="true">●</span>
            <span v-if="file.data.value && !file.data.value.binary" class="ml-auto shrink-0">
                {{ t('files.lines', { count: lineCount }) }}
                <template v-if="!editable"> · {{ t('files.readOnly') }}</template>
            </span>
            <Button
                v-if="editable && !file.data.value?.binary"
                :disabled="!dirty || save.isPending.value"
                v-tooltip="t('saveHint')"
                @click="saveDraft"
            >
                <PhFloppyDisk />{{ t('save') }}
            </Button>
        </div>
        <div v-if="!selected" class="grid flex-1 place-items-center text-muted">
            {{ t('files.pick') }}
        </div>
        <div v-else-if="file.isPending.value" class="flex flex-col gap-1.5 p-4">
            <div
                v-for="i in 8"
                :key="i"
                class="skeleton h-3.5"
                :style="{ width: `${40 + ((i * 37) % 50)}%` }"
            />
        </div>
        <div v-else-if="file.data.value?.binary" class="grid flex-1 place-items-center text-muted">
            {{ t('files.binary') }}
        </div>
        <template v-else-if="file.data.value">
            <p v-if="file.data.value.truncated" class="m-0 px-4 pt-2 text-muted">
                {{ t('files.truncated') }}
            </p>
            <CodeView
                :path="file.data.value.path"
                :content="content"
                :read-only="!editable"
                @change="remember"
                @save="saveDraft"
            />
        </template>
    </section>
</template>

<style scoped>
.editor-head {
    display: flex;
    height: 40px;
    flex: none;
    align-items: center;
    gap: 8px;
    padding: 0 6px 0 14px;
    border-bottom: 1px solid var(--ui-border);
    color: var(--ui-muted-foreground);
    font-size: 12.5px;
}
</style>

<i18n lang="yaml">
ru:
    save: Сохранить
    saveHint: Сохранить (Ctrl+S)
    unsaved: Есть несохранённые изменения
en:
    save: Save
    saveHint: Save (Ctrl+S)
    unsaved: Unsaved changes
</i18n>
