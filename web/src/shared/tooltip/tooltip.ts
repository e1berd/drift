import type { Directive } from 'vue'
import { createTooltipController } from '@ui-sdk/vue/tooltip'

const OPEN_DELAY_MS = 500

export const tooltipController = createTooltipController<string>()

interface TooltipBinding {
    text: string | undefined
    cancel: () => void
    detach: () => void
}

const bindings = new WeakMap<HTMLElement, TooltipBinding>()

function attach(el: HTMLElement): TooltipBinding {
    let timer: ReturnType<typeof setTimeout> | undefined
    const binding: TooltipBinding = { text: undefined, cancel, detach }

    function cancel() {
        clearTimeout(timer)
        if (tooltipController.isOpen()) tooltipController.close()
    }

    function show(delay: number) {
        clearTimeout(timer)
        timer = setTimeout(() => {
            if (binding.text) tooltipController.open({ payload: binding.text, anchor: el })
        }, delay)
    }

    const onPointerEnter = () => show(OPEN_DELAY_MS)
    const onFocusIn = (event: FocusEvent) => {
        if (event.target instanceof HTMLElement && event.target.matches(':focus-visible')) show(0)
    }
    const onLeave = () => cancel()
    const onKeyDown = (event: KeyboardEvent) => event.key === 'Escape' && cancel()

    el.addEventListener('pointerenter', onPointerEnter)
    el.addEventListener('pointerleave', onLeave)
    el.addEventListener('pointerdown', onLeave)
    el.addEventListener('focusin', onFocusIn)
    el.addEventListener('focusout', onLeave)
    el.addEventListener('keydown', onKeyDown)

    function detach() {
        cancel()
        el.removeEventListener('pointerenter', onPointerEnter)
        el.removeEventListener('pointerleave', onLeave)
        el.removeEventListener('pointerdown', onLeave)
        el.removeEventListener('focusin', onFocusIn)
        el.removeEventListener('focusout', onLeave)
        el.removeEventListener('keydown', onKeyDown)
    }

    return binding
}

export const vTooltip: Directive<HTMLElement, string | undefined> = {
    mounted(el, { value }) {
        const binding = attach(el)
        binding.text = value
        bindings.set(el, binding)
    },
    updated(el, { value }) {
        const binding = bindings.get(el)
        if (binding) binding.text = value
    },
    beforeUnmount(el) {
        bindings.get(el)?.detach()
        bindings.delete(el)
    },
}
