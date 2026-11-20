export interface LinkEnds {
    id: string
    sourceY: number
    targetY: number
    sourceHidden: boolean
    targetHidden: boolean
}

export interface Gutter {
    left: number
    right: number
}

export interface LinkRoute {
    id: string
    d: string
    dashed: boolean
}

const EDGE_INSET = 8
const MAX_LANE_STEP = 8
const MIN_LANE_STEP = 3
const GAP_HALF = 3
const EPSILON = 0.5

function spanOf(link: LinkEnds): [number, number] {
    return [Math.min(link.sourceY, link.targetY), Math.max(link.sourceY, link.targetY)]
}

function strictlyInside(y: number, [low, high]: [number, number]): boolean {
    return y > low + EPSILON && y < high - EPSILON
}

function hasRightLeg(link: LinkEnds): boolean {
    return !link.sourceHidden
}

function leftOfPairs(links: LinkEnds[]): Array<[number, number]> {
    const pairs: Array<[number, number]> = []
    links.forEach((a, i) =>
        links.forEach((b, j) => {
            if (i === j) return
            const span = spanOf(b)
            if (hasRightLeg(a) && strictlyInside(a.sourceY, span)) pairs.push([j, i])
            if (strictlyInside(a.targetY, span)) pairs.push([i, j])
        }),
    )
    return pairs
}

function nextFree(remaining: Set<number>, incoming: number[], links: LinkEnds[]): number {
    let best = -1
    for (const index of remaining) {
        if (
            best === -1 ||
            incoming[index]! < incoming[best]! ||
            (incoming[index] === incoming[best] && links[index]!.sourceY < links[best]!.sourceY)
        )
            best = index
    }
    return best
}

/** Lane rank of every link, 0 nearest the info bar, chosen so legs cross other verticals as rarely as possible. */
export function laneRanks(links: LinkEnds[]): number[] {
    const pairs = leftOfPairs(links)
    const incoming = links.map(() => 0)
    for (const [, right] of pairs) incoming[right]! += 1
    const remaining = new Set(links.map((_, index) => index))
    const ranks = links.map(() => 0)
    let rank = 0
    while (remaining.size) {
        const index = nextFree(remaining, incoming, links)
        remaining.delete(index)
        ranks[index] = rank++
        for (const [left, right] of pairs) if (left === index) incoming[right]! -= 1
    }
    return ranks
}

function laneXs(links: LinkEnds[], gutter: Gutter): number[] {
    const room = Math.max(0, gutter.right - gutter.left - EDGE_INSET * 2)
    const gaps = Math.max(1, links.length - 1)
    const step = Math.max(MIN_LANE_STEP, Math.min(MAX_LANE_STEP, room / gaps))
    const bundle = step * (links.length - 1)
    const first = gutter.left + EDGE_INSET + Math.max(0, (room - bundle) / 2)
    return laneRanks(links).map((rank) =>
        Math.min(gutter.right - EDGE_INSET, Math.round(first + rank * step)),
    )
}

function crossingsOnVertical(index: number, links: LinkEnds[], lanes: number[]): number[] {
    const span = spanOf(links[index]!)
    const lane = lanes[index]!
    const ys: number[] = []
    links.forEach((other, j) => {
        if (j === index) return
        const otherLane = lanes[j]!
        if (hasRightLeg(other) && otherLane < lane && strictlyInside(other.sourceY, span))
            ys.push(other.sourceY)
        if (otherLane > lane && strictlyInside(other.targetY, span)) ys.push(other.targetY)
    })
    return ys
}

function verticalWithGaps(x: number, from: number, to: number, gaps: number[]): string {
    const direction = to >= from ? 1 : -1
    const ordered = gaps.toSorted((a, b) => (a - b) * direction)
    const parts = ordered.map(
        (y) => `V${y - direction * GAP_HALF} M${x} ${y + direction * GAP_HALF}`,
    )
    return [...parts, `V${to}`].join(' ')
}

function pathOf(link: LinkEnds, lane: number, gutter: Gutter, gaps: number[]): string {
    const start = link.sourceHidden
        ? `M${lane} ${link.sourceY}`
        : `M${gutter.right} ${link.sourceY} H${lane}`
    return `${start} ${verticalWithGaps(lane, link.sourceY, link.targetY, gaps)} H${gutter.left}`
}

/** Orthogonal routes from tool cards (right) to info bar rows (left) through the gutter between them. */
export function routeLinks(links: LinkEnds[], gutter: Gutter): LinkRoute[] {
    const lanes = laneXs(links, gutter)
    return links.map((link, index) => ({
        id: link.id,
        d: pathOf(link, lanes[index]!, gutter, crossingsOnVertical(index, links, lanes)),
        dashed: link.sourceHidden || link.targetHidden,
    }))
}
