/// <reference types="vite/client" />

import type { Directive } from 'vue'
import type { AutoAnimateOptions } from '@formkit/auto-animate'

declare module 'vue' {
    interface GlobalDirectives {
        vAutoAnimate: Directive<HTMLElement, Partial<AutoAnimateOptions> | undefined>
    }
}
