<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

import DraftItem from '@/components/sidebar/DraftItem.vue'
import SessionItem from '@/components/sidebar/SessionItem.vue'
import { STANDALONE_DRAFT_KEY, useDraftIndex } from '@/composables/useDrafts'
import { useStandaloneSessions } from '@/shared/queries'

const { t } = useI18n()
const route = useRoute()
const drafts = useDraftIndex()
const { sessions, isPending, hasNextPage, isFetchingNextPage, fetchNextPage } =
    useStandaloneSessions()

const viewingNew = computed(() => route.name === 'new' && !route.query.project)
const showDraft = computed(() => drafts.has(STANDALONE_DRAFT_KEY) && !viewingNew.value)
</script>

<template>
    <section :aria-label="t('nav.chats')">
        <div class="px-3 pt-3 pb-1 text-muted">{{ t('nav.chats') }}</div>
        <div v-auto-animate class="flex flex-col gap-px px-2">
            <DraftItem
                v-if="showDraft"
                :key="STANDALONE_DRAFT_KEY"
                :draft-key="STANDALONE_DRAFT_KEY"
            />
            <SessionItem v-for="session in sessions" :key="session.id" :session="session" />
            <div v-if="isPending" key="skeleton" class="skeleton my-px h-8" />
            <p
                v-else-if="!sessions.length && !showDraft"
                key="empty"
                class="m-0 px-2 py-1 text-muted"
            >
                {{ t('nav.noChats') }}
            </p>
            <div v-if="isFetchingNextPage" key="next" class="skeleton my-px h-8" />
            <Button v-else-if="hasNextPage" key="more" @click="fetchNextPage()">
                {{ t('nav.showMore') }}
            </Button>
        </div>
    </section>
</template>
