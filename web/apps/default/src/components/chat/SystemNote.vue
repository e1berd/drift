<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'

import type { MessageBody } from '@drift/vue/api'

const props = defineProps<{ body: MessageBody }>()
const { t } = useI18n()

function removalKey(body: MessageBody): string {
    if (body.target === 'file') return 'chat.notes.removedFile'
    return body.branch ? 'chat.notes.removedWorktreeBranch' : 'chat.notes.removedWorktree'
}

const sentence = computed(() => {
    const body = props.body
    if (!body.code) return ''
    if (body.code === 'removed')
        return t(removalKey(body), { path: body.path, branch: body.branch })
    return t(`chat.notes.${body.code}`)
})
</script>

<template>
    <div class="flex flex-col items-center gap-1 self-center text-center text-muted">
        <p class="m-0">
            {{ sentence }}
            <RouterLink
                v-if="body.code === 'noSubscription'"
                :to="{ name: 'settings' }"
                class="ms-1"
            >
                {{ t('chat.notes.openSettings') }}
            </RouterLink>
        </p>
        <p v-if="body.text" class="m-0 font-mono whitespace-pre-wrap">{{ body.text }}</p>
    </div>
</template>
