/**
 * Drill-Through Configuration Library
 * 
 * Provides types and utilities for configuring drill-through navigation
 * in dashboards and charts.
 */

/**
 * Type of drill target
 */
export type DrillTargetType = 'dashboard' | 'page' | 'url' | 'modal';

/**
 * Parameter transformation function
 */
export type ParameterTransform = (value: any, context?: Record<string, any>) => any;

/**
 * Parameter mapping configuration
 */
export interface ParameterMapping {
    /** Source field name from clicked data */
    sourceField: string;

    /** Target parameter name in destination */
    targetParameter: string;

    /** Optional transformation function */
    transform?: ParameterTransform;

    /** Whether this parameter is required */
    required?: boolean;

    /** Default value if source is undefined */
    defaultValue?: any;
}

/**
 * Drill target configuration
 */
export interface DrillTarget {
    /** Unique identifier for this drill target */
    id: string;

    /** Type of drill target */
    type: DrillTargetType;

    /** Target identifier (dashboard ID, page path, or URL) */
    targetId: string;

    /** Display label for this drill target */
    label?: string;

    /** Parameter mappings */
    parameterMappings: ParameterMapping[];

    /** Whether to open in new tab/window */
    openInNewTab?: boolean;

    /** Additional metadata */
    metadata?: Record<string, any>;
}

/**
 * Drill level in a hierarchical path
 */
export interface DrillLevel {
    /** Level identifier */
    id: string;

    /** Display name for this level */
    name: string;

    /** Field name at this level */
    fieldName: string;

    /** Current value at this level (if drilled) */
    value?: any;

    /** Drill target for this level */
    drillTarget?: DrillTarget;

    /** Whether this level is the current active level */
    isActive?: boolean;
}

/**
 * Complete drill path configuration
 */
export interface DrillPath {
    /** Unique identifier for this drill path */
    id: string;

    /** Display name for this drill path */
    name: string;

    /** Ordered levels in the drill path */
    levels: DrillLevel[];

    /** Current drill level index (0-based) */
    currentLevel: number;

    /** Whether drill-through is enabled */
    enabled?: boolean;
}

/**
 * Drill configuration for a chart component
 */
export interface ChartDrillConfig {
    /** Chart identifier */
    chartId: string;

    /** Available drill paths */
    drillPaths: DrillPath[];

    /** Default drill path ID (if multiple paths available) */
    defaultDrillPathId?: string;

    /** Whether to show drill breadcrumb */
    showBreadcrumb?: boolean;

    /** Whether drill is enabled for this chart */
    enabled?: boolean;
}

/**
 * Built-in parameter transforms
 */
export const ParameterTransforms = {
    /** Convert to uppercase */
    toUpperCase: (value: any) => String(value).toUpperCase(),

    /** Convert to lowercase */
    toLowerCase: (value: any) => String(value).toLowerCase(),

    /** Convert to number */
    toNumber: (value: any) => Number(value),

    /** Convert to string */
    toString: (value: any) => String(value),

    /** Format as date (ISO) */
    toISODate: (value: any) => new Date(value).toISOString().split('T')[0],

    /** Trim whitespace */
    trim: (value: any) => String(value).trim(),

    /** URL encode */
    urlEncode: (value: any) => encodeURIComponent(String(value)),

    /** URL decode */
    urlDecode: (value: any) => decodeURIComponent(String(value)),

    /** JSON stringify */
    toJSON: (value: any) => JSON.stringify(value),

    /** JSON parse */
    fromJSON: (value: any) => JSON.parse(String(value)),
};

/**
 * Create a parameter mapping
 */
export function createParameterMapping(
    sourceField: string,
    targetParameter: string,
    options?: {
        transform?: ParameterTransform;
        required?: boolean;
        defaultValue?: any;
    }
): ParameterMapping {
    return {
        sourceField,
        targetParameter,
        transform: options?.transform,
        required: options?.required ?? false,
        defaultValue: options?.defaultValue,
    };
}

/**
 * Create a drill target
 */
export function createDrillTarget(
    id: string,
    type: DrillTargetType,
    targetId: string,
    parameterMappings: ParameterMapping[],
    options?: {
        label?: string;
        openInNewTab?: boolean;
        metadata?: Record<string, any>;
    }
): DrillTarget {
    return {
        id,
        type,
        targetId,
        label: options?.label,
        parameterMappings,
        openInNewTab: options?.openInNewTab ?? false,
        metadata: options?.metadata,
    };
}

/**
 * Create a drill level
 */
export function createDrillLevel(
    id: string,
    name: string,
    fieldName: string,
    options?: {
        value?: any;
        drillTarget?: DrillTarget;
        isActive?: boolean;
    }
): DrillLevel {
    return {
        id,
        name,
        fieldName,
        value: options?.value,
        drillTarget: options?.drillTarget,
        isActive: options?.isActive ?? false,
    };
}

/**
 * Create a drill path
 */
export function createDrillPath(
    id: string,
    name: string,
    levels: DrillLevel[],
    options?: {
        currentLevel?: number;
        enabled?: boolean;
    }
): DrillPath {
    return {
        id,
        name,
        levels,
        currentLevel: options?.currentLevel ?? 0,
        enabled: options?.enabled ?? true,
    };
}

/**
 * Apply parameter mappings to source data
 */
export function applyParameterMappings(
    sourceData: Record<string, any>,
    mappings: ParameterMapping[],
    context?: Record<string, any>
): Record<string, any> {
    const result: Record<string, any> = {};

    for (const mapping of mappings) {
        let value = sourceData[mapping.sourceField];

        // Use default value if source is undefined
        if (value === undefined && mapping.defaultValue !== undefined) {
            value = mapping.defaultValue;
        }

        // Check required constraint
        if (mapping.required && (value === undefined || value === null)) {
            throw new Error(
                `Required parameter '${mapping.sourceField}' is missing or null`
            );
        }

        // Apply transformation if provided
        if (value !== undefined && mapping.transform) {
            try {
                value = mapping.transform(value, context);
            } catch (error) {
                console.error(
                    `Error transforming parameter '${mapping.sourceField}':`,
                    error
                );
                // Continue with untransformed value
            }
        }

        // Only add to result if value is defined
        if (value !== undefined) {
            result[mapping.targetParameter] = value;
        }
    }

    return result;
}

/**
 * Build drill URL with parameters
 */
export function buildDrillUrl(
    baseUrl: string,
    parameters: Record<string, any>,
    options?: {
        includeHash?: boolean;
        hashParams?: Record<string, any>;
    }
): string {
    const url = new URL(baseUrl, window.location.origin);

    // Add query parameters
    Object.entries(parameters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
            url.searchParams.set(key, String(value));
        }
    });

    // Add hash parameters if specified
    if (options?.includeHash && options.hashParams) {
        const hashParams = new URLSearchParams();
        Object.entries(options.hashParams).forEach(([key, value]) => {
            if (value !== undefined && value !== null) {
                hashParams.set(key, String(value));
            }
        });
        url.hash = hashParams.toString();
    }

    return url.toString();
}

/**
 * Validate drill configuration
 */
export function validateDrillConfig(config: ChartDrillConfig): {
    valid: boolean;
    errors: string[];
} {
    const errors: string[] = [];

    if (!config.chartId) {
        errors.push('Chart ID is required');
    }

    if (!config.drillPaths || config.drillPaths.length === 0) {
        errors.push('At least one drill path is required');
    }

    config.drillPaths.forEach((path, pathIndex) => {
        if (!path.id) {
            errors.push(`Drill path ${pathIndex} is missing an ID`);
        }

        if (!path.levels || path.levels.length === 0) {
            errors.push(`Drill path '${path.id}' has no levels`);
        }

        path.levels.forEach((level, levelIndex) => {
            if (!level.id) {
                errors.push(
                    `Level ${levelIndex} in drill path '${path.id}' is missing an ID`
                );
            }
            if (!level.fieldName) {
                errors.push(
                    `Level '${level.id}' in drill path '${path.id}' is missing a field name`
                );
            }
        });
    });

    return {
        valid: errors.length === 0,
        errors,
    };
}

/**
 * Get next drill level
 */
export function getNextDrillLevel(drillPath: DrillPath): DrillLevel | null {
    if (drillPath.currentLevel >= drillPath.levels.length - 1) {
        return null; // Already at the last level
    }
    return drillPath.levels[drillPath.currentLevel + 1];
}

/**
 * Get previous drill level
 */
export function getPreviousDrillLevel(drillPath: DrillPath): DrillLevel | null {
    if (drillPath.currentLevel <= 0) {
        return null; // Already at the first level
    }
    return drillPath.levels[drillPath.currentLevel - 1];
}

/**
 * Check if can drill down
 */
export function canDrillDown(drillPath: DrillPath): boolean {
    return (
        drillPath.enabled !== false &&
        drillPath.currentLevel < drillPath.levels.length - 1
    );
}

/**
 * Check if can drill up
 */
export function canDrillUp(drillPath: DrillPath): boolean {
    return drillPath.currentLevel > 0;
}

/**
 * Get breadcrumb trail
 */
export function getBreadcrumbTrail(drillPath: DrillPath): DrillLevel[] {
    return drillPath.levels.slice(0, drillPath.currentLevel + 1);
}

/**
 * Example: Create a simple hierarchical drill path
 */
export function createHierarchicalDrillPath(
    id: string,
    name: string,
    hierarchy: Array<{
        id: string;
        name: string;
        fieldName: string;
        targetDashboardId?: string;
    }>
): DrillPath {
    const levels: DrillLevel[] = hierarchy.map((item, index) => {
        const drillTarget =
            item.targetDashboardId && index < hierarchy.length - 1
                ? createDrillTarget(
                    `${item.id}-target`,
                    'dashboard',
                    item.targetDashboardId,
                    [
                        createParameterMapping(item.fieldName, 'filter_value', {
                            required: true,
                        }),
                    ]
                )
                : undefined;

        return createDrillLevel(item.id, item.name, item.fieldName, {
            drillTarget,
            isActive: index === 0,
        });
    });

    return createDrillPath(id, name, levels);
}
