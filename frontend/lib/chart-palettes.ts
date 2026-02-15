// Chart Color Palettes - TASK-047
// Comprehensive color palette library untuk semua chart types

/**
 * Color Scale Type
 */
export type ColorScaleType = 'sequential' | 'diverging' | 'categorical' | 'gradient'

/**
 * Color Palette Definition
 */
export interface ColorPalette {
    id: string
    name: string
    type: ColorScaleType
    colors: string[]
    description?: string
    preview?: string
}

/**
 * Sequential Color Palettes (Low to High)
 * Best for: Heatmaps, Choropleth maps, single-metric visualization
 */
export const SEQUENTIAL_PALETTES: ColorPalette[] = [
    {
        id: 'blues',
        name: 'Blues',
        type: 'sequential',
        colors: ['#f0f9ff', '#bae6fd', '#7dd3fc', '#38bdf8', '#0ea5e9', '#0284c7', '#0369a1', '#075985', '#0c4a6e'],
        description: 'Light blue to dark blue progression'
    },
    {
        id: 'greens',
        name: 'Greens',
        type: 'sequential',
        colors: ['#f0fdf4', '#bbf7d0', '#86efac', '#4ade80', '#22c55e', '#16a34a', '#15803d', '#166534', '#14532d'],
        description: 'Light green to dark green progression'
    },
    {
        id: 'oranges',
        name: 'Oranges',
        type: 'sequential',
        colors: ['#fff7ed', '#fed7aa', '#fdba74', '#fb923c', '#f97316', '#ea580c', '#c2410c', '#9a3412', '#7c2d12'],
        description: 'Light orange to dark orange progression'
    },
    {
        id: 'purples',
        name: 'Purples',
        type: 'sequential',
        colors: ['#faf5ff', '#e9d5ff', '#d8b4fe', '#c084fc', '#a855f7', '#9333ea', '#7e22ce', '#6b21a8', '#581c87'],
        description: 'Light purple to dark purple progression'
    },
    {
        id: 'reds',
        name: 'Reds',
        type: 'sequential',
        colors: ['#fef2f2', '#fecaca', '#fca5a5', '#f87171', '#ef4444', '#dc2626', '#b91c1c', '#991b1b', '#7f1d1d'],
        description: 'Light red to dark red progression'
    },
    {
        id: 'grays',
        name: 'Grays',
        type: 'sequential',
        colors: ['#f9fafb', '#f3f4f6', '#e5e7eb', '#d1d5db', '#9ca3af', '#6b7280', '#4b5563', '#374151', '#1f2937'],
        description: 'Light gray to dark gray progression'
    }
]

/**
 * Diverging Color Palettes (Low - Neutral - High)
 * Best for: Variance analysis, positive/negative values, correlation matrices
 */
export const DIVERGING_PALETTES: ColorPalette[] = [
    {
        id: 'red-blue',
        name: 'Red-Blue',
        type: 'diverging',
        colors: ['#b91c1c', '#dc2626', '#ef4444', '#f87171', '#fecaca', '#f3f4f6', '#bae6fd', '#7dd3fc', '#38bdf8', '#0ea5e9', '#0369a1'],
        description: 'Red (negative) to blue (positive) with gray neutral'
    },
    {
        id: 'red-green',
        name: 'Red-Green',
        type: 'diverging',
        colors: ['#b91c1c', '#dc2626', '#ef4444', '#f87171', '#fecaca', '#f3f4f6', '#bbf7d0', '#86efac', '#4ade80', '#22c55e', '#16a34a'],
        description: 'Red (negative) to green (positive) with gray neutral'
    },
    {
        id: 'purple-green',
        name: 'Purple-Green',
        type: 'diverging',
        colors: ['#7e22ce', '#9333ea', '#a855f7', '#c084fc', '#e9d5ff', '#f3f4f6', '#bbf7d0', '#86efac', '#4ade80', '#22c55e', '#15803d'],
        description: 'Purple (negative) to green (positive) with gray neutral'
    },
    {
        id: 'cool-warm',
        name: 'Cool-Warm',
        type: 'diverging',
        colors: ['#0c4a6e', '#0369a1', '#0284c7', '#0ea5e9', '#7dd3fc', '#f3f4f6', '#fed7aa', '#fb923c', '#f97316', '#ea580c', '#9a3412'],
        description: 'Cool blues to warm oranges'
    }
]

/**
 * Categorical Color Palettes
 * Best for: Bar charts, pie charts, scatter plots, multiple series
 */
export const CATEGORICAL_PALETTES: ColorPalette[] = [
    {
        id: 'default',
        name: 'Default',
        type: 'categorical',
        colors: ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#14b8a6', '#f97316', '#06b6d4', '#84cc16'],
        description: 'Balanced, distinct colors for categories'
    },
    {
        id: 'vibrant',
        name: 'Vibrant',
        type: 'categorical',
        colors: ['#e60049', '#0bb4ff', '#50e991', '#e6d800', '#9b19f5', '#ffa300', '#dc0ab4', '#00bfa0', '#b3d4ff', '#fdcce5'],
        description: 'High contrast, vibrant colors'
    },
    {
        id: 'pastel',
        name: 'Pastel',
        type: 'categorical',
        colors: ['#fbb4ae', '#b3cde3', '#ccebc5', '#decbe4', '#fed9a6', '#ffffcc', '#e5d8bd', '#fddaec', '#f2f2f2', '#e0e0e0'],
        description: 'Soft, muted pastel colors'
    },
    {
        id: 'earth',
        name: 'Earth Tones',
        type: 'categorical',
        colors: ['#8b4513', '#a0522d', '#d2691e', '#cd853f', '#deb887', '#f4a460', '#d2b48c', '#bc8f8f', '#cd5c5c', '#f08080'],
        description: 'Natural, earthy colors'
    },
    {
        id: 'ocean',
        name: 'Ocean',
        type: 'categorical',
        colors: ['#006994', '#0582ca', '#00a6fb', '#0fc2c0', '#008dd5', '#007ea7', '#0a369d', '#4392f1', '#5b5f97', '#7a93ac'],
        description: 'Blue ocean-inspired palette'
    },
    {
        id: 'sunset',
        name: 'Sunset',
        type: 'categorical',
        colors: ['#ff6b35', '#f7931e', '#fdc500', '#c1a57b', '#d7263d', '#f46036', '#2e294e', '#011627', '#ff6d00', '#ffba08'],
        description: 'Warm sunset colors'
    },
    {
        id: 'forest',
        name: 'Forest',
        type: 'categorical',
        colors: ['#2d6a4f', '#40916c', '#52b788', '#74c69d', '#95d5b2', '#b7e4c7', '#8b9556', '#a68a59', '#68a357', '#94b49f'],
        description: 'Green forest tones'
    },
    {
        id: 'neon',
        name: 'Neon',
        type: 'categorical',
        colors: ['#ff006e', '#fb5607', '#ffbe0b', '#8338ec', '#3a86ff', '#06ffa5', '#ff006e', '#8338ec', '#ffbe0b', '#3a86ff'],
        description: 'Bright neon colors'
    },
    {
        id: 'corporate',
        name: 'Corporate',
        type: 'categorical',
        colors: ['#003f5c', '#2f4b7c', '#665191', '#a05195', '#d45087', '#f95d6a', '#ff7c43', '#ffa600', '#7f8c8d', '#95a5a6'],
        description: 'Professional, corporate colors'
    },
    {
        id: 'minimal',
        name: 'Minimal',
        type: 'categorical',
        colors: ['#000000', '#333333', '#666666', '#999999', '#cccccc', '#2c3e50', '#34495e', '#7f8c8d', '#95a5a6', '#bdc3c7'],
        description: 'Minimal grayscale with accents'
    }
]

/**
 * Gradient Palettes
 * Best for: Backgrounds, fills, modern visualizations
 */
export const GRADIENT_PALETTES: ColorPalette[] = [
    {
        id: 'sunset-gradient',
        name: 'Sunset Gradient',
        type: 'gradient',
        colors: ['#ff6b6b', '#ffa502', '#ffd93d', '#6bcf7f'],
        description: 'Sunset gradient from red to green'
    },
    {
        id: 'ocean-gradient',
        name: 'Ocean Gradient',
        type: 'gradient',
        colors: ['#667eea', '#764ba2', '#f093fb', '#4facfe'],
        description: 'Ocean waves gradient'
    },
    {
        id: 'fire-gradient',
        name: 'Fire Gradient',
        type: 'gradient',
        colors: ['#f2709c', '#ff9472', '#ffaa00', '#ffd600'],
        description: 'Fire gradient from pink to yellow'
    },
    {
        id: 'mint-gradient',
        name: 'Mint Gradient',
        type: 'gradient',
        colors: ['#00b4db', '#0083b0', '#00d2ff', '#3a7bd5'],
        description: 'Cool mint gradient'
    }
]

/**
 * All Palettes Combined
 */
export const ALL_PALETTES: ColorPalette[] = [
    ...SEQUENTIAL_PALETTES,
    ...DIVERGING_PALETTES,
    ...CATEGORICAL_PALETTES,
    ...GRADIENT_PALETTES
]

/**
 * Get palette by ID
 */
export function getPaletteById(id: string): ColorPalette | undefined {
    return ALL_PALETTES.find(p => p.id === id)
}

/**
 * Get palettes by type
 */
export function getPalettesByType(type: ColorScaleType): ColorPalette[] {
    return ALL_PALETTES.filter(p => p.type === type)
}

/**
 * Generate color for index in palette
 */
export function getColorFromPalette(paletteId: string, index: number): string {
    const palette = getPaletteById(paletteId)
    if (!palette) return '#3b82f6' // Default blue

    const colors = palette.colors
    return colors[index % colors.length]
}

/**
 * Generate color for value in sequential/diverging palette
 */
export function getColorForValue(
    paletteId: string,
    value: number,
    min: number,
    max: number
): string {
    const palette = getPaletteById(paletteId)
    if (!palette) return '#3b82f6'

    const colors = palette.colors
    const normalized = (value - min) / (max - min)
    const index = Math.floor(normalized * (colors.length - 1))
    const clampedIndex = Math.max(0, Math.min(colors.length - 1, index))

    return colors[clampedIndex]
}

/**
 * Interpolate between two colors
 */
export function interpolateColor(color1: string, color2: string, factor: number): string {
    const hex2rgb = (hex: string) => {
        const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
        return result ? {
            r: parseInt(result[1], 16),
            g: parseInt(result[2], 16),
            b: parseInt(result[3], 16)
        } : { r: 0, g: 0, b: 0 }
    }

    const rgb2hex = (r: number, g: number, b: number) => {
        return '#' + [r, g, b].map(x => {
            const hex = Math.round(x).toString(16)
            return hex.length === 1 ? '0' + hex : hex
        }).join('')
    }

    const c1 = hex2rgb(color1)
    const c2 = hex2rgb(color2)

    const r = c1.r + factor * (c2.r - c1.r)
    const g = c1.g + factor * (c2.g - c1.g)
    const b = c1.b + factor * (c2.b - c1.b)

    return rgb2hex(r, g, b)
}

/**
 * Generate custom gradient palette
 */
export function createGradientPalette(
    startColor: string,
    endColor: string,
    steps: number = 9
): string[] {
    const palette: string[] = []

    for (let i = 0; i < steps; i++) {
        const factor = i / (steps - 1)
        palette.push(interpolateColor(startColor, endColor, factor))
    }

    return palette
}

/**
 * Validate hex color
 */
export function isValidHexColor(color: string): boolean {
    return /^#[0-9A-F]{6}$/i.test(color)
}

/**
 * Convert RGB to Hex
 */
export function rgbToHex(r: number, g: number, b: number): string {
    return '#' + [r, g, b].map(x => {
        const hex = x.toString(16)
        return hex.length === 1 ? '0' + hex : hex
    }).join('')
}

/**
 * Convert Hex to RGB
 */
export function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
    return result ? {
        r: parseInt(result[1], 16),
        g: parseInt(result[2], 16),
        b: parseInt(result[3], 16)
    } : null
}

/**
 * Get contrasting text color (black or white) for background
 */
export function getContrastColor(backgroundColor: string): string {
    const rgb = hexToRgb(backgroundColor)
    if (!rgb) return '#000000'

    // Calculate relative luminance
    const luminance = (0.299 * rgb.r + 0.587 * rgb.g + 0.114 * rgb.b) / 255

    return luminance > 0.5 ? '#000000' : '#ffffff'
}

/**
 * Default palette IDs untuk different chart types
 */
export const DEFAULT_PALETTES = {
    bar: 'default',
    line: 'default',
    pie: 'vibrant',
    scatter: 'ocean',
    heatmap: 'blues',
    choropleth: 'greens',
    treemap: 'forest',
    sankey: 'sunset',
    waterfall: 'red-green',
    funnel: 'purples'
} as const
