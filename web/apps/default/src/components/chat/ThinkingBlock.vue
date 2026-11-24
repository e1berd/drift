<script setup lang="ts">
import { ref } from 'vue'
import { Collapsible, CollapsiblePanel, CollapsibleTrigger } from '@ui-sdk/vue/collapsible'
import { useI18n } from 'vue-i18n'

import type { MessageBody } from '@drift/vue/api'

const props = defineProps<{ body: MessageBody }>()
const { t } = useI18n()
const open = ref(props.body.status === 'running')
</script>

<template>
    <Collapsible v-model:open="open" class="ms-9 max-md:ms-0">
        <CollapsibleTrigger>
            <span v-if="body.status === 'running'" class="motion-shine">{{
                t('chat.thinking')
            }}</span>
            <span v-else>{{ t('chat.thought') }} {{ body.meta }}</span>
        </CollapsibleTrigger>
        <CollapsiblePanel>
            <p class="m-0 whitespace-pre-wrap text-muted italic">{{ body.text }}</p>
        </CollapsiblePanel>
    </Collapsible>
</template>
