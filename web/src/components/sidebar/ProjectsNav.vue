<script setup lang="ts">
import { ref, useTemplateRef } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import { Button } from '@ui-sdk/vue/button'
import { useI18n } from 'vue-i18n'

import NewProjectForm from '@/components/sidebar/NewProjectForm.vue'
import ProjectItem from '@/components/sidebar/ProjectItem.vue'
import { useStickyProjects } from '@/composables/useStickyProjects'
import { useProjects } from '@/shared/queries'

const { t } = useI18n()
const { projects, isPending } = useProjects()
const creating = ref(false)
const { pinStyle } = useStickyProjects(useTemplateRef<HTMLElement>('projectList'))
</script>

<template>
    <div>
        <div class="px-3 pt-2">
            <Button class="w-full" :aria-expanded="creating" @click="creating = !creating">
                <PhPlus />{{ t('nav.newProject') }}
            </Button>
        </div>

        <div class="grid-reveal px-3" :data-open="creating || undefined">
            <div>
                <NewProjectForm v-if="creating" @done="creating = false" />
            </div>
        </div>

        <nav class="px-2 pt-2 pb-2" :aria-label="t('nav.projects')">
            <div v-if="isPending" class="flex flex-col gap-1 pt-1">
                <div v-for="row in 3" :key="row" class="skeleton h-8" />
            </div>
            <p v-else-if="!projects.length && !creating" class="px-2 py-2 text-muted">
                {{ t('nav.noProjects') }}
            </p>
            <div ref="projectList">
                <ProjectItem
                    v-for="(project, index) in projects"
                    :key="project.id"
                    :project="project"
                    :pin-style="pinStyle(index, projects.length)"
                />
            </div>
        </nav>
    </div>
</template>
