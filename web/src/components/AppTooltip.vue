<script setup lang="ts">
import { Tooltip, TooltipContent } from '@ui-sdk/vue/tooltip'

import { tooltipController } from '@/shared/tooltip/tooltip'

const HintContent = TooltipContent<string>
</script>

<template>
    <Tooltip :controller="tooltipController" type="label">
        <HintContent v-slot="{ state }" class="whitespace-pre-line">{{
            state.payload
        }}</HintContent>
    </Tooltip>
</template>
