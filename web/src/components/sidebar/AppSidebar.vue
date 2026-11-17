<script setup lang="ts">
import { computed, ref } from 'vue'
import { PhMagnifyingGlass, PhPlus } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

import BrandMark from '@/components/BrandMark.vue'
import NewProjectForm from '@/components/sidebar/NewProjectForm.vue'
import ProjectItem from '@/components/sidebar/ProjectItem.vue'
import SidebarFooter from '@/components/sidebar/SidebarFooter.vue'
import { useShell } from '@/composables/useShell'
import { useI18n } from '@/shared/i18n'
import { useProjects } from '@/shared/queries'

const { t } = useI18n()
const { closeOnNarrow } = useShell()
const { projects, isPending } = useProjects()

const query = ref('')
const creating = ref(false)
const filter = computed(() => query.value.trim().toLowerCase())
</script>

<template>
    <div class="flex h-full flex-col">
        <div class="sb-head">
            <RouterLink
                to="/"
                class="flex items-center gap-2 text-[15px] font-semibold tracking-tight"
            >
                <BrandMark :size="20" />drift
            </RouterLink>
            <RouterLink to="/" class="btn h-[30px]" @click="closeOnNarrow">
                <PhPlus :size="15" />{{ t('nav.newChat') }}
            </RouterLink>
        </div>

        <label class="search">
            <PhMagnifyingGlass :size="15" class="search-icon" />
            <span class="sr-only">{{ t('nav.search') }}</span>
            <input
                v-model="query"
                class="field pl-8"
                :placeholder="t('nav.search')"
                autocomplete="off"
            />
        </label>

        <div class="sec-h">
            <span>{{ t('nav.projects') }}</span>
            <button
                class="icon-btn small"
                :aria-label="t('nav.newProject')"
                :title="t('nav.newProject')"
                :aria-expanded="creating"
                @click="creating = !creating"
            >
                <PhPlus :size="15" />
            </button>
        </div>

        <div class="grid-reveal px-2" :data-open="creating || undefined">
            <div>
                <NewProjectForm v-if="creating" @done="creating = false" />
            </div>
        </div>

        <nav class="min-h-0 flex-1 overflow-auto px-2 pb-2.5" :aria-label="t('nav.projects')">
            <div v-if="isPending" class="flex flex-col gap-1 pt-1">
                <div
                    v-for="i in 3"
                    :key="i"
                    class="skeleton h-[34px]"
                    :style="{ opacity: 1 - i * 0.25 }"
                />
            </div>
            <p
                v-else-if="!projects.length && !creating"
                class="px-2.5 py-2 text-[12.5px] leading-relaxed text-faint"
            >
                {{ t('nav.noProjects') }}
            </p>
            <div v-auto-animate>
                <ProjectItem
                    v-for="(p, i) in projects"
                    :key="p.id"
                    :project="p"
                    :filter="filter"
                    :is-default="i === 0"
                />
            </div>
        </nav>

        <SidebarFooter />
    </div>
</template>

<style scoped>
.sb-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    height: var(--topbar-h);
    flex: none;
    padding: 0 10px 0 16px;
    border-bottom: 1px solid var(--ui-border);
}

.search {
    position: relative;
    margin: 10px 10px 4px;
}

.search-icon {
    position: absolute;
    top: 8px;
    left: 9px;
    color: var(--faint);
    pointer-events: none;
}

.sec-h {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 8px 4px 16px;
    color: var(--faint);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
}
</style>
