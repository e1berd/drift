import { meter } from '@ui-sdk/core/meter'
import { createParts } from '@ui-sdk/vue/primitive'

const {
    Root: InlineMeter,
    Track: InlineMeterTrack,
    Indicator: InlineMeterIndicator,
} = createParts(meter, {
    root: 'span',
    track: 'span',
    indicator: 'span',
})

export { InlineMeter, InlineMeterIndicator, InlineMeterTrack }
