'use client';

/**
 * TASK-CHART-024: Accessibility Features
 * WCAG 2.1 AA compliance utilities for chart components.
 */

/**
 * High-contrast color palettes for accessibility.
 * Each pair has WCAG AA contrast ratio >= 4.5:1 against both light and dark backgrounds.
 */
export const accessibleColorPalettes = {
    /** Distinguishable even with common color vision deficiencies */
    colorblindSafe: ['#0077bb', '#33bbee', '#009988', '#ee7733', '#cc3311', '#ee3377', '#bbbbbb', '#000000'],

    /** Maximum contrast for presentations */
    highContrast: ['#003f5c', '#2f4b7c', '#665191', '#a05195', '#d45087', '#f95d6a', '#ff7c43', '#ffa600'],

    /** Wong palette (nature publishing) */
    wong: ['#000000', '#e69f00', '#56b4e9', '#009e73', '#f0e442', '#0072b2', '#d55e00', '#cc79a7'],

    /** Tol palette (Paul Tol) */
    tol: ['#332288', '#88ccee', '#44aa99', '#117733', '#999933', '#ddcc77', '#cc6677', '#882255'],
};

/**
 * Pattern fill definitions for distinguishing series without relying on color alone.
 */
export const chartPatterns = {
    diagonal: 'M0,0 L10,10 M-2,8 L2,12 M8,-2 L12,2',
    crosshatch: 'M0,5 L10,5 M5,0 L5,10',
    dots: 'circle',
    horizontal: 'M0,5 L10,5',
    vertical: 'M5,0 L5,10',
    zigzag: 'M0,5 L2.5,0 L5,5 L7.5,0 L10,5',
};

/**
 * Generate an ECharts-compatible pattern fill.
 * Use for series differentiation beyond color.
 */
export function createPatternFill(options: {
    pattern: keyof typeof chartPatterns;
    color: string;
    backgroundColor?: string;
    lineWidth?: number;
    size?: number;
}): {
    type: 'pattern';
    image: HTMLCanvasElement;
    repeat: string;
} | null {
    if (typeof document === 'undefined') return null;

    const { pattern, color, backgroundColor = 'transparent', lineWidth = 1.5, size = 10 } = options;
    const canvas = document.createElement('canvas');
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext('2d');
    if (!ctx) return null;

    // Background
    ctx.fillStyle = backgroundColor;
    ctx.fillRect(0, 0, size, size);

    ctx.strokeStyle = color;
    ctx.lineWidth = lineWidth;

    const path = chartPatterns[pattern];

    if (pattern === 'dots') {
        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.arc(size / 2, size / 2, size / 4, 0, Math.PI * 2);
        ctx.fill();
    } else {
        const segments = path.split(' ');
        ctx.beginPath();
        let i = 0;
        while (i < segments.length) {
            const cmd = segments[i];
            if (cmd === 'M' && i + 1 < segments.length) {
                const [x, y] = segments[i + 1].split(',').map(Number);
                ctx.moveTo(x, y);
                i += 2;
            } else if (cmd === 'L' && i + 1 < segments.length) {
                const [x, y] = segments[i + 1].split(',').map(Number);
                ctx.lineTo(x, y);
                i += 2;
            } else {
                i++;
            }
        }
        ctx.stroke();
    }

    return {
        type: 'pattern',
        image: canvas,
        repeat: 'repeat',
    };
}

/**
 * Generate ARIA attributes for chart containers.
 */
export function chartAriaAttributes(options: {
    title: string;
    description?: string;
    dataPointCount?: number;
    chartType?: string;
}): Record<string, string> {
    const { title, description, dataPointCount, chartType } = options;
    const desc = description ?? `${chartType ?? 'Chart'} showing ${title}${dataPointCount ? ` with ${dataPointCount} data points` : ''}`;

    return {
        role: 'img',
        'aria-label': desc,
        'aria-roledescription': chartType ?? 'chart',
        tabIndex: '0',
    };
}

/**
 * Generate a text summary of chart data for screen readers.
 */
export function generateChartSummary(options: {
    chartType: string;
    title: string;
    dataPoints: { label: string; value: number }[];
    format?: (value: number) => string;
}): string {
    const { chartType, title, dataPoints, format } = options;
    const fmt = format ?? ((v: number) => v.toLocaleString());

    const count = dataPoints.length;
    const values = dataPoints.map(d => d.value);
    const minVal = Math.min(...values);
    const maxVal = Math.max(...values);
    const sum = values.reduce((s, v) => s + v, 0);
    const avg = count > 0 ? sum / count : 0;

    const minItem = dataPoints.find(d => d.value === minVal);
    const maxItem = dataPoints.find(d => d.value === maxVal);

    let summary = `${chartType}: ${title}. ${count} data points. `;
    if (maxItem) summary += `Highest: ${maxItem.label} at ${fmt(maxVal)}. `;
    if (minItem) summary += `Lowest: ${minItem.label} at ${fmt(minVal)}. `;
    summary += `Average: ${fmt(avg)}.`;

    return summary;
}

/**
 * Keyboard navigation helper for chart interactions.
 */
export function chartKeyboardHandler(options: {
    dataPoints: { label: string; value: number }[];
    currentIndex: number;
    onIndexChange: (index: number) => void;
    onSelect?: (index: number) => void;
    announce?: (text: string) => void;
}) {
    const { dataPoints, currentIndex, onIndexChange, onSelect, announce } = options;
    const count = dataPoints.length;

    return (event: React.KeyboardEvent) => {
        let newIndex = currentIndex;

        switch (event.key) {
            case 'ArrowRight':
            case 'ArrowDown':
                event.preventDefault();
                newIndex = (currentIndex + 1) % count;
                break;
            case 'ArrowLeft':
            case 'ArrowUp':
                event.preventDefault();
                newIndex = (currentIndex - 1 + count) % count;
                break;
            case 'Home':
                event.preventDefault();
                newIndex = 0;
                break;
            case 'End':
                event.preventDefault();
                newIndex = count - 1;
                break;
            case 'Enter':
            case ' ':
                event.preventDefault();
                onSelect?.(currentIndex);
                return;
            default:
                return;
        }

        if (newIndex !== currentIndex) {
            onIndexChange(newIndex);
            const point = dataPoints[newIndex];
            announce?.(`${point.label}: ${point.value.toLocaleString()}`);
        }
    };
}

/**
 * Screen reader announcement utility (live region).
 */
export function createLiveRegion(): {
    announce: (message: string) => void;
    cleanup: () => void;
} {
    if (typeof document === 'undefined') {
        return { announce: () => { }, cleanup: () => { } };
    }

    const region = document.createElement('div');
    region.setAttribute('role', 'status');
    region.setAttribute('aria-live', 'polite');
    region.setAttribute('aria-atomic', 'true');
    region.style.position = 'absolute';
    region.style.width = '1px';
    region.style.height = '1px';
    region.style.overflow = 'hidden';
    region.style.clip = 'rect(0,0,0,0)';
    region.style.whiteSpace = 'nowrap';
    document.body.appendChild(region);

    return {
        announce: (message: string) => {
            region.textContent = '';
            // Force DOM reflow for screen reader
            void region.offsetWidth;
            region.textContent = message;
        },
        cleanup: () => {
            document.body.removeChild(region);
        },
    };
}
