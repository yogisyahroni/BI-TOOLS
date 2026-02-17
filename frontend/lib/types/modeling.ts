// Modeling API Types
// Purpose: TypeScript types for Modeling API (metric definitions for governance)

export interface ModelDefinition {
    id: string;
    name: string;
    description: string;
    type: 'table' | 'view' | 'query';
    sourceTable?: string;
    sourceQuery?: string;
    workspaceId: string;
    createdBy: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
    metrics?: MetricDefinition[];
    createdAt: string;
    updatedAt: string;
}

export interface MetricDefinition {
    id: string;
    name: string;
    description: string;
    formula: string;
    modelId?: string;
    dataType: 'number' | 'currency' | 'percentage' | 'count' | 'decimal';
    format?: string;
    aggregationType?: 'sum' | 'avg' | 'count' | 'min' | 'max' | 'count_distinct';
    workspaceId: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    createdBy: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
    createdAt: string;
    updatedAt: string;
}

// Request Types

export interface CreateModelDefinitionRequest {
    name: string;
    description?: string;
    type: 'table' | 'view' | 'query';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    sourceTable?: string;
    sourceQuery?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
}

export interface UpdateModelDefinitionRequest {
    name?: string;
    description?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    type?: 'table' | 'view' | 'query';
    sourceTable?: string;
    sourceQuery?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
}

export interface CreateMetricDefinitionRequest {
    name: string;
    description?: string;
    formula: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    modelId?: string;
    dataType: 'number' | 'currency' | 'percentage' | 'count' | 'decimal';
    format?: string;
    aggregationType?: 'sum' | 'avg' | 'count' | 'min' | 'max' | 'count_distinct';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
}

export interface UpdateMetricDefinitionRequest {
    name?: string;
    description?: string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    formula?: string;
    modelId?: string;
    dataType?: 'number' | 'currency' | 'percentage' | 'count' | 'decimal';
    format?: string;
    aggregationType?: 'sum' | 'avg' | 'count' | 'min' | 'max' | 'count_distinct';
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    metadata?: Record<string, any>;
}
