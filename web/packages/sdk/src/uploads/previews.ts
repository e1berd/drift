const previews = new Map<string, string>()

/** A local URL showing an image while it uploads, kept until the attachment leaves the composer. */
export function localPreview(key: string, data: unknown): string | undefined {
    const known = previews.get(key)
    if (known) return known
    if (!(data instanceof Blob) || typeof URL.createObjectURL !== 'function') return undefined
    const url = URL.createObjectURL(data)
    previews.set(key, url)
    return url
}

export function knownPreview(key: string | undefined): string | undefined {
    return key ? previews.get(key) : undefined
}

export function forgetPreview(key: string | undefined): void {
    const url = key ? previews.get(key) : undefined
    if (!key || !url) return
    URL.revokeObjectURL(url)
    previews.delete(key)
}
