<script setup lang="ts">
import { useI18n } from 'vue-i18n'

import { useSession } from '@/composables/useSession'

const { t } = useI18n()
const { session } = useSession()
</script>

<template>
  <section class="flex flex-col gap-2">
    <h1 class="text-2xl font-semibold">Drift</h1>
    <p class="text-(--ui-muted-foreground)">
      {{
        t('signed_in_as', { name: session?.identity.traits.name || session?.identity.traits.email })
      }}
    </p>
  </section>
</template>

<i18n lang="yaml">
ru:
  signed_in_as: Вы вошли как {name}.
en:
  signed_in_as: Signed in as {name}.
</i18n>
