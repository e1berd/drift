<script setup lang="ts">
import { mdiClose } from '@mdi/js'
import { useI18n } from 'vue-i18n'

import { closeNotice, notices, pressNoticeAction } from '@/shared/snackbarNotifier'
import type { NoticeColor } from '@/shared/snackbarNotifier'

const { t } = useI18n()

const ALERT_TYPES: Record<NoticeColor, 'info' | 'success' | 'error'> = {
    info: 'info',
    success: 'success',
    error: 'error',
    loading: 'info',
}
</script>

<template>
    <div class="snackbar-host" role="region" :aria-label="t('notices')">
        <TransitionGroup name="notice">
            <v-alert
                v-for="notice in notices"
                :key="notice.key"
                :type="ALERT_TYPES[notice.color]"
                :title="notice.title"
                :text="notice.description"
                variant="elevated"
                density="compact"
                :role="notice.color === 'error' ? 'alert' : 'status'"
            >
                <template v-if="notice.color === 'loading'" #prepend>
                    <v-progress-circular indeterminate size="20" width="2" />
                </template>
                <template #append>
                    <v-btn v-if="notice.action" size="small" @click="pressNoticeAction(notice)">{{
                        notice.action.label
                    }}</v-btn>
                    <v-btn
                        :icon="mdiClose"
                        size="small"
                        :aria-label="t('close')"
                        @click="closeNotice(notice.key, true)"
                    />
                </template>
            </v-alert>
        </TransitionGroup>
    </div>
</template>

<style scoped>
.snackbar-host {
    position: fixed;
    inset-block-end: 16px;
    inset-inline-end: 16px;
    z-index: 2400;
    display: flex;
    flex-direction: column;
    gap: 8px;
    inline-size: min(420px, calc(100vw - 32px));
}

.notice-enter-active,
.notice-leave-active {
    transition:
        opacity 0.2s,
        translate 0.2s;
}

.notice-enter-from,
.notice-leave-to {
    opacity: 0;
    translate: 0 8px;
}
</style>
