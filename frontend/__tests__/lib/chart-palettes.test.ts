import { describe, it, expect } from 'vitest'
import {
    getPaletteById,
    getPalettesByType,
    getColorFromPalette,
    getColorForValue,
    interpolateColor,
    createGradientPalette,
    isValidHexColor,
    rgbToHex,
    hexToRgb,
    getContrastColor,
    ALL_PALETTES,
    SEQUENTIAL_PALETTES,
    DIVERGING_PALETTES,
    CATEGORICAL_PALETTES,
    GRADIENT_PALETTES,
    DEFAULT_PALETTES,
} from '@/lib/chart-palettes'

// ─────────────────────────────────────────────────────────────────────────────
// Palette Constants
// ─────────────────────────────────────────────────────────────────────────────

describe('Palette Constants', () => {
    it('ALL_PALETTES is the union of all categories', () => {
        const expected =
            SEQUENTIAL_PALETTES.length +
            DIVERGING_PALETTES.length +
            CATEGORICAL_PALETTES.length +
            GRADIENT_PALETTES.length
        expect(ALL_PALETTES.length).toBe(expected)
    })

    it('every palette has a unique id', () => {
        const ids = ALL_PALETTES.map((p) => p.id)
        const unique = new Set(ids)
        expect(unique.size).toBe(ids.length)
    })

    it('DEFAULT_PALETTES maps each chart type to a valid palette id', () => {
        const defaultIds = Object.values(DEFAULT_PALETTES)
        defaultIds.forEach((id) => {
            expect(getPaletteById(id)).toBeDefined()
        })
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// getPaletteById
// ─────────────────────────────────────────────────────────────────────────────

describe('getPaletteById', () => {
    it('should return palette for valid id', () => {
        const palette = getPaletteById('blues')
        expect(palette).toBeDefined()
        expect(palette!.id).toBe('blues')
        expect(palette!.type).toBe('sequential')
    })

    it('should return undefined for invalid id', () => {
        expect(getPaletteById('nonexistent-palette-xyz')).toBeUndefined()
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// getPalettesByType
// ─────────────────────────────────────────────────────────────────────────────

describe('getPalettesByType', () => {
    it('should return only sequential palettes', () => {
        const result = getPalettesByType('sequential')
        expect(result.length).toBe(SEQUENTIAL_PALETTES.length)
        result.forEach((p) => expect(p.type).toBe('sequential'))
    })

    it('should return only categorical palettes', () => {
        const result = getPalettesByType('categorical')
        expect(result.length).toBe(CATEGORICAL_PALETTES.length)
        result.forEach((p) => expect(p.type).toBe('categorical'))
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// getColorFromPalette
// ─────────────────────────────────────────────────────────────────────────────

describe('getColorFromPalette', () => {
    it('should return first color for index 0', () => {
        const color = getColorFromPalette('blues', 0)
        expect(color).toBe('#f0f9ff')
    })

    it('should wrap around when index exceeds colors length', () => {
        const palette = getPaletteById('blues')!
        const wrapIndex = palette.colors.length + 2
        const color = getColorFromPalette('blues', wrapIndex)
        expect(color).toBe(palette.colors[2])
    })

    it('should return default blue for unknown palette', () => {
        const color = getColorFromPalette('nonexistent', 0)
        expect(color).toBe('#3b82f6')
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// getColorForValue
// ─────────────────────────────────────────────────────────────────────────────

describe('getColorForValue', () => {
    it('should return first color for min value', () => {
        const color = getColorForValue('blues', 0, 0, 100)
        expect(color).toBe('#f0f9ff')
    })

    it('should return last color for max value', () => {
        const palette = getPaletteById('blues')!
        const color = getColorForValue('blues', 100, 0, 100)
        expect(color).toBe(palette.colors[palette.colors.length - 1])
    })

    it('should return middle color for midpoint value', () => {
        const color = getColorForValue('blues', 50, 0, 100)
        expect(color).toBeDefined()
        expect(typeof color).toBe('string')
    })

    it('should return default blue for unknown palette', () => {
        expect(getColorForValue('nonexistent', 50, 0, 100)).toBe('#3b82f6')
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// interpolateColor
// ─────────────────────────────────────────────────────────────────────────────

describe('interpolateColor', () => {
    it('should return first color at factor 0', () => {
        const result = interpolateColor('#000000', '#ffffff', 0)
        expect(result).toBe('#000000')
    })

    it('should return second color at factor 1', () => {
        const result = interpolateColor('#000000', '#ffffff', 1)
        expect(result).toBe('#ffffff')
    })

    it('should return midpoint color at factor 0.5', () => {
        const result = interpolateColor('#000000', '#ffffff', 0.5)
        // Midpoint of black and white should be gray ~#808080
        expect(result).toMatch(/^#[0-9a-f]{6}$/i)
        const rgb = hexToRgb(result)!
        expect(rgb.r).toBeGreaterThanOrEqual(125)
        expect(rgb.r).toBeLessThanOrEqual(130)
    })

    it('should produce valid hex colors', () => {
        const result = interpolateColor('#ff0000', '#0000ff', 0.3)
        expect(isValidHexColor(result)).toBe(true)
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// createGradientPalette
// ─────────────────────────────────────────────────────────────────────────────

describe('createGradientPalette', () => {
    it('should produce the requested number of steps', () => {
        const palette = createGradientPalette('#000000', '#ffffff', 5)
        expect(palette.length).toBe(5)
    })

    it('should start with startColor and end with endColor', () => {
        const palette = createGradientPalette('#ff0000', '#0000ff', 9)
        expect(palette[0]).toBe('#ff0000')
        expect(palette[palette.length - 1]).toBe('#0000ff')
    })

    it('should default to 9 steps', () => {
        const palette = createGradientPalette('#000000', '#ffffff')
        expect(palette.length).toBe(9)
    })

    it('all colors should be valid hex', () => {
        const palette = createGradientPalette('#ff6600', '#0066ff', 7)
        palette.forEach((color) => {
            expect(isValidHexColor(color)).toBe(true)
        })
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// isValidHexColor
// ─────────────────────────────────────────────────────────────────────────────

describe('isValidHexColor', () => {
    it('should accept valid 6-digit hex color', () => {
        expect(isValidHexColor('#ff0000')).toBe(true)
        expect(isValidHexColor('#AABBCC')).toBe(true)
    })

    it('should reject short hex', () => {
        expect(isValidHexColor('#fff')).toBe(false)
    })

    it('should reject missing hash', () => {
        expect(isValidHexColor('ff0000')).toBe(false)
    })

    it('should reject invalid hex chars', () => {
        expect(isValidHexColor('#gggggg')).toBe(false)
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// rgbToHex / hexToRgb (round-trip)
// ─────────────────────────────────────────────────────────────────────────────

describe('rgbToHex', () => {
    it('should convert (255, 0, 0) to #ff0000', () => {
        expect(rgbToHex(255, 0, 0)).toBe('#ff0000')
    })

    it('should pad single-digit channels', () => {
        expect(rgbToHex(0, 0, 0)).toBe('#000000')
        expect(rgbToHex(1, 2, 3)).toBe('#010203')
    })
})

describe('hexToRgb', () => {
    it('should parse #ff0000 to { r:255, g:0, b:0 }', () => {
        expect(hexToRgb('#ff0000')).toEqual({ r: 255, g: 0, b: 0 })
    })

    it('should handle uppercase hex', () => {
        expect(hexToRgb('#AABBCC')).toEqual({ r: 170, g: 187, b: 204 })
    })

    it('should handle hex without hash', () => {
        expect(hexToRgb('ff0000')).toEqual({ r: 255, g: 0, b: 0 })
    })

    it('should return null for invalid hex', () => {
        expect(hexToRgb('invalid')).toBeNull()
        expect(hexToRgb('#ff')).toBeNull()
    })
})

describe('rgbToHex / hexToRgb round-trip', () => {
    it('should survive round-trip conversion', () => {
        const original = { r: 128, g: 64, b: 200 }
        const hex = rgbToHex(original.r, original.g, original.b)
        const back = hexToRgb(hex)
        expect(back).toEqual(original)
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// getContrastColor
// ─────────────────────────────────────────────────────────────────────────────

describe('getContrastColor', () => {
    it('should return black text for light backgrounds', () => {
        expect(getContrastColor('#ffffff')).toBe('#000000')
        expect(getContrastColor('#f0f0f0')).toBe('#000000')
    })

    it('should return white text for dark backgrounds', () => {
        expect(getContrastColor('#000000')).toBe('#ffffff')
        expect(getContrastColor('#1a1a1a')).toBe('#ffffff')
    })

    it('should return black for invalid color', () => {
        expect(getContrastColor('not-a-color')).toBe('#000000')
    })
})
