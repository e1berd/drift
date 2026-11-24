import { afterEach, describe, expect, it } from 'vitest'

import { frontendId, frontendName, pickFrontend, pickedFrontend } from '../frontendChoice'

describe('frontend choice', () => {
    afterEach(() => {
        document.cookie = 'ui=; path=/; max-age=0'
    })

    it('follows the instance default until the browser picks an interface', () => {
        expect(pickedFrontend()).toBeNull()
    })

    it('keeps the pick in the ui cookie the server reads', () => {
        pickFrontend('frontends/vuetify')
        expect(document.cookie).toContain('ui=vuetify')
        expect(pickedFrontend()).toBe('frontends/vuetify')
    })

    it('converts between ids and resource names', () => {
        expect(frontendName('default')).toBe('frontends/default')
        expect(frontendId('frontends/vuetify')).toBe('vuetify')
    })
})
