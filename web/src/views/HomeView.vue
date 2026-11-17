<script setup lang="ts">
import { useSession } from '@/composables/useSession'

const { session } = useSession()
</script>

<template>
  <section class="flex flex-col gap-2">
    <h1 class="text-2xl font-semibold">Drift</h1>
    <p class="text-(--ui-muted-foreground)">
      Вы вошли как {{ session?.identity.traits.name || session?.identity.traits.email }}.
    </p>
  </section>
</template>
