<script setup lang="ts">
import { ref } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import NewProjectForm from '@/components/sidebar/NewProjectForm.vue'
import ProjectItem from '@/components/sidebar/ProjectItem.vue'
import StandaloneChats from '@/components/sidebar/StandaloneChats.vue'
import { useProjects } from '@/shared/queries'

const { t } = useI18n()
const { projects, isPending } = useProjects()
const creating = ref(false)
</script>

<template>
    <div>
        <StandaloneChats />

        <div class="flex items-center justify-between px-3 pt-3 pb-1 text-muted">
            <span>{{ t('nav.projects') }}</span>
            <Button
                :aria-label="t('nav.newProject')"
                v-tooltip="t('nav.newProject')"
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

        <nav class="px-2 pb-2" :aria-label="t('nav.projects')">
            <div v-if="isPending" class="flex flex-col gap-1 pt-1">
                <div v-for="row in 3" :key="row" class="skeleton h-8" />
            </div>
            <p v-else-if="!projects.length && !creating" class="px-2 py-2 text-muted">
                {{ t('nav.noProjects') }}
            </p>
            <div v-auto-animate>
                <ProjectItem v-for="project in projects" :key="project.id" :project="project" />
            </div>
        </nav>
    </div>
</template>
