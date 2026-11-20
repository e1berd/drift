<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Input } from '@ui-sdk/vue/input'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRouter } from 'vue-router'

import ScrollPane from '@/components/ScrollPane.vue'
import BrandMark from '@/components/BrandMark.vue'
import DeliveryVersion from '@/components/delivery/DeliveryVersion.vue'
import SessionSearchResults from '@/components/sidebar/SessionSearchResults.vue'
import SidebarFooter from '@/components/sidebar/SidebarFooter.vue'
import WorkspaceNav from '@/components/sidebar/WorkspaceNav.vue'
import { useDebounced } from '@/composables/useDebounced'
import { useShell } from '@/composables/useShell'

const { t } = useI18n()
const router = useRouter()
const { closeOnNarrow } = useShell()

const query = ref('')
const searchText = useDebounced(() => query.value.trim(), 250)
const searching = computed(() => !!searchText.value)

function startNewChat() {
    closeOnNarrow()
    void router.push('/')
}
</script>

<template>
    <div class="flex h-full flex-col">
        <div class="sidebar-head">
            <span class="flex min-w-0 items-center gap-2">
                <RouterLink
                    to="/"
                    class="flex flex-none items-center gap-2 font-semibold no-underline"
                    ><BrandMark />drift</RouterLink
                >
                <DeliveryVersion />
            </span>
            <Button
                class="flex-none"
                :aria-label="t('nav.newChat')"
                v-tooltip="t('nav.newChat')"
                @click="startNewChat"
                ><PhPlus /><span class="new-chat-label">{{ t('nav.newChat') }}</span></Button
            >
        </div>

        <div class="px-3 pt-3">
            <Input
                v-model="query"
                type="search"
                maxlength="200"
                :placeholder="t('nav.search')"
                :aria-label="t('nav.search')"
                @keydown.esc="query = ''"
            />
        </div>

        <ScrollPane class="min-h-0 flex-1">
            <div v-auto-animate>
                <SessionSearchResults v-if="searching" key="search" :text="searchText" />
                <WorkspaceNav v-else key="browse" />
            </div>
        </ScrollPane>

        <SidebarFooter />
    </div>
</template>

<style scoped>
.sidebar-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    height: var(--topbar-h);
    flex: none;
    padding: 0 10px 0 14px;
    border-bottom: 1px solid var(--ui-border);
    container-type: inline-size;
}

@container (max-width: 300px) {
    .new-chat-label {
        display: none;
    }
}
</style>
