<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { Input } from '@ui-sdk/vue/input'
import { useI18n } from 'vue-i18n'
import { RouterLink, useRouter } from 'vue-router'

import BrandMark from '@/components/BrandMark.vue'
import NewProjectForm from '@/components/sidebar/NewProjectForm.vue'
import ProjectItem from '@/components/sidebar/ProjectItem.vue'
import SidebarFooter from '@/components/sidebar/SidebarFooter.vue'
import { useShell } from '@/composables/useShell'
import { useProjects } from '@/shared/queries'

const { t } = useI18n()
const router = useRouter()
const { closeOnNarrow } = useShell()
const { projects, isPending } = useProjects()

const query = ref('')
const creating = ref(false)
const filter = computed(() => query.value.trim().toLowerCase())

function startNewChat() {
    closeOnNarrow()
    void router.push('/')
}
</script>

<template>
    <div class="flex h-full flex-col">
        <div class="sidebar-head">
            <RouterLink to="/" class="flex items-center gap-2 font-semibold no-underline"
                ><BrandMark />drift</RouterLink
            >
            <Button @click="startNewChat"><PhPlus />{{ t('nav.newChat') }}</Button>
        </div>

        <div class="px-3 pt-3">
            <Input
                v-model="query"
                type="search"
                :placeholder="t('nav.search')"
                :aria-label="t('nav.search')"
            />
        </div>

        <div class="flex items-center justify-between px-3 pt-3 pb-1 text-muted">
            <span>{{ t('nav.projects') }}</span>
            <Button
                :aria-label="t('nav.newProject')"
                :title="t('nav.newProject')"
                :aria-expanded="creating"
                @click="creating = !creating"
            >
                <PhPlus />
            </Button>
        </div>

        <div class="grid-reveal px-3" :data-open="creating || undefined">
            <div>
                <NewProjectForm v-if="creating" @done="creating = false" />
            </div>
        </div>

        <nav class="min-h-0 flex-1 overflow-auto px-2 pb-2" :aria-label="t('nav.projects')">
            <div v-if="isPending" class="flex flex-col gap-1 pt-1">
                <div v-for="row in 3" :key="row" class="skeleton h-8" />
            </div>
            <p v-else-if="!projects.length && !creating" class="px-2 py-2 text-muted">
                {{ t('nav.noProjects') }}
            </p>
            <div v-auto-animate>
                <ProjectItem
                    v-for="(project, index) in projects"
                    :key="project.id"
                    :project="project"
                    :filter="filter"
                    :is-default="index === 0"
                />
            </div>
        </nav>

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
}
</style>
