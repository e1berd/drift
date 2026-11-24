const SIZE_UNITS = ['byte', 'kilobyte', 'megabyte', 'gigabyte', 'terabyte'] as const

/** "12 MB", "1.5 GB": a byte count in decimal units, as file pickers and upload limits show it. */
export function formatFileSize(bytes: number, locale?: string): string {
    let value = bytes
    let unit = 0
    while (Math.round(value) >= 1000 && unit < SIZE_UNITS.length - 1) {
        value /= 1000
        unit++
    }
    return new Intl.NumberFormat(locale, {
        style: 'unit',
        unit: SIZE_UNITS[unit],
        unitDisplay: 'short',
        maximumFractionDigits: value < 10 && unit > 0 ? 1 : 0,
    }).format(value)
}
