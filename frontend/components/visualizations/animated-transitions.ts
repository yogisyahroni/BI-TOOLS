'use client';

/**
 * TASK-CHART-023: Animated Transitions
 * Animation presets and utilities for chart transitions.
 */

export interface AnimationPreset {
    animationDuration: number;
    animationEasing: string;
    animationDelay?: number | ((idx: number) => number);
    animationDurationUpdate?: number;
    animationEasingUpdate?: string;
    animationDelayUpdate?: number | ((idx: number) => number);
}

/**
 * Pre-defined animation presets for common use cases.
 */
export const animationPresets: Record<string, AnimationPreset> = {
    /** Smooth entrance with cubic ease-out */
    smooth: {
        animationDuration: 700,
        animationEasing: 'cubicOut',
        animationDurationUpdate: 500,
        animationEasingUpdate: 'cubicInOut',
    },

    /** Fast snap for real-time data */
    snap: {
        animationDuration: 200,
        animationEasing: 'linear',
        animationDurationUpdate: 200,
        animationEasingUpdate: 'linear',
    },

    /** Dramatic bounce entrance */
    bounce: {
        animationDuration: 1000,
        animationEasing: 'elasticOut',
        animationDurationUpdate: 500,
        animationEasingUpdate: 'cubicInOut',
    },

    /** Spring physics feel */
    spring: {
        animationDuration: 800,
        animationEasing: 'backOut',
        animationDurationUpdate: 400,
        animationEasingUpdate: 'cubicOut',
    },

    /** Staggered cascade for lists */
    stagger: {
        animationDuration: 600,
        animationEasing: 'cubicOut',
        animationDelay: (idx: number) => idx * 50,
        animationDurationUpdate: 400,
        animationEasingUpdate: 'cubicInOut',
        animationDelayUpdate: (idx: number) => idx * 30,
    },

    /** Slow cinematic reveal */
    cinematic: {
        animationDuration: 1500,
        animationEasing: 'sinusoidalInOut',
        animationDurationUpdate: 800,
        animationEasingUpdate: 'sinusoidalInOut',
    },

    /** No animation for performance-critical views */
    none: {
        animationDuration: 0,
        animationEasing: 'linear',
        animationDurationUpdate: 0,
        animationEasingUpdate: 'linear',
    },
};

/**
 * Get the ECharts animation config from a preset name.
 */
export function getAnimationConfig(preset: keyof typeof animationPresets | AnimationPreset): AnimationPreset {
    if (typeof preset === 'string') {
        return animationPresets[preset] ?? animationPresets.smooth;
    }
    return preset;
}

/**
 * Merge animation preset into an ECharts option object.
 */
export function applyAnimation(
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    option: Record<string, any>,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    preset: keyof typeof animationPresets | AnimationPreset,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
): Record<string, any> {
    const config = getAnimationConfig(preset);
    return {
        ...option,
        ...config,
    };
}

/**
 * Generate stagger delays for series items.
 */
export function staggerDelay(options: {
    baseDelay?: number;
    increment?: number;
    maxDelay?: number;
}): (idx: number) => number {
    const { baseDelay = 0, increment = 40, maxDelay = 2000 } = options;
    return (idx: number) => Math.min(baseDelay + idx * increment, maxDelay);
}

/**
 * CSS transition class builder for React components with chart-like transitions.
 */
export function chartTransitionClasses(options?: {
    duration?: number;
    easing?: string;
    properties?: string[];
}): string {
    const { duration = 300, easing = 'ease-out', properties = ['all'] } = options ?? {};
    const props = properties.join(', ');
    return `transition-[${props}] duration-[${duration}ms] ease-[${easing}]`;
}

/**
 * ECharts series-level animation for progressive data loading.
 */
export function progressiveAnimation(totalItems: number): {
    progressive: number;
    progressiveThreshold: number;
    animationDuration: number;
    animationEasing: string;
} {
    return {
        progressive: Math.max(100, Math.floor(totalItems / 10)),
        progressiveThreshold: Math.max(500, totalItems / 2),
        animationDuration: Math.min(2000, 300 + totalItems * 2),
        animationEasing: 'cubicOut',
    };
}
