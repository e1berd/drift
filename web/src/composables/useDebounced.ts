import { onScopeDispose, ref, toValue, watch } from 'vue'

import type { MaybeRefOrGetter, Ref } from 'vue'

/** A copy of `source` that follows it once it stops changing for `delay` ms; an empty value passes at once. */
export function useDebounced<T>(source: MaybeRefOrGetter<T>, delay: number): Readonly<Ref<T>> {
    const settled = ref(toValue(source)) as Ref<T>
    let timer: ReturnType<typeof setTimeout> | undefined

    watch(
        () => toValue(source),
        (value) => {
            clearTimeout(timer)
            if (!value) settled.value = value
            else timer = setTimeout(() => (settled.value = value), delay)
        },
    )
    onScopeDispose(() => clearTimeout(timer))

    return settled
}
