import { describe, expect, it } from 'vitest'

import { laneRanks, routeLinks } from '../linkRoutes'
import type { LinkEnds } from '../linkRoutes'

const gutter = { left: 100, right: 200 }

function link(id: string, sourceY: number, targetY: number, sourceHidden = false): LinkEnds {
    return { id, sourceY, targetY, sourceHidden, targetHidden: false }
}

describe('routeLinks', () => {
    it('draws an orthogonal route from the card to the row', () => {
        const [route] = routeLinks([link('a', 300, 120)], gutter)
        expect(route).toEqual({ id: 'a', d: 'M200 300 H150 V120 H100', dashed: false })
    })

    it('starts a hidden card at its edge, dashed and without the card leg', () => {
        const [route] = routeLinks([link('a', 0, 120, true)], gutter)
        expect(route?.d).toBe('M150 0 V120 H100')
        expect(route?.dashed).toBe(true)
    })

    it('orders lanes so partly overlapping links do not cross', () => {
        const links = [link('a', 100, 300), link('b', 200, 400)]
        const routes = routeLinks(links, gutter)
        expect(laneRanks(links)).toEqual([0, 1])
        expect(routes.every((route) => !route.d.includes(' M'))).toBe(true)
    })

    it('breaks the vertical where a crossing cannot be avoided', () => {
        const links = [link('outer', 300, 150), link('inner', 180, 230)]
        const routes = routeLinks(links, gutter)
        const broken = routes.filter((route) => route.d.includes(' M'))
        expect(broken).toHaveLength(1)
    })

    it('keeps separate lanes for links whose cards are all above the screen', () => {
        const routes = routeLinks([link('a', 0, 100, true), link('b', 0, 140, true)], gutter)
        const lanes = routes.map((route) => route.d.split(' ').at(0))
        expect(new Set(lanes).size).toBe(2)
    })
})
