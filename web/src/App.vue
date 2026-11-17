<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useI18n } from 'vue-i18n'

import { useSession } from '@/composables/useSession'

const { t } = useI18n()
const { session, refresh } = useSession()

onMounted(refresh)
</script>

<template>
  <div class="min-h-dvh bg-(--ui-background) text-(--ui-foreground)">
    <header class="flex items-center justify-between gap-4 border-b border-(--ui-border) px-4 py-3">
      <nav class="flex items-center gap-4 text-sm">
        <RouterLink to="/" class="font-semibold">Drift</RouterLink>
        <RouterLink
          to="/settings"
          class="text-(--ui-muted-foreground) hover:text-(--ui-foreground)"
        >
          {{ t('passkeys.title') }}
        </RouterLink>
      </nav>
      <div v-if="session" class="flex items-center gap-3 text-sm">
        <span class="text-(--ui-muted-foreground)">{{ session.identity.traits.email }}</span>
        <a href="/logout" class="underline underline-offset-4">{{ t('actions.sign_out') }}</a>
      </div>
    </header>
    <main class="px-4 py-8">
      <RouterView v-if="session" />
    </main>
  </div>
</template>
