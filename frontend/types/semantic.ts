export interface SemanticModel {
    id: string;
    name: string;
    description?: string;
    dataSourceId: string;
    tableName: string; // The physical table name
    workspaceId: string;
    createdBy: string;
    dimensions: SemanticDimension[];
    metrics: SemanticMetric[];
    createdAt: string;
    updatedAt: string;
}

export interface SemanticDimension {
    id: string;
    modelId: string;
    name: string;
    columnName: string;
    dataType: string;
    description?: string;
    isHidden: boolean;
    createdAt: string;
    updatedAt: string;
}

export interface SemanticMetric {
    id: string;
    modelId: string;
    name: string;
    formula: string;
    description?: string;
    format?: string;
    createdAt: string;
    updatedAt: string;
}

// Request Types

export interface CreateDimensionRequest {
    name: string;
    columnName: string;
    dataType: string;
    description?: string;
    isHidden: boolean;
}

export interface CreateMetricRequest {
    name: string;
    formula: string;
    description?: string;
    format?: string;
}

export interface CreateSemanticModelRequest {
    name: string;
    description?: string;
    dataSourceId: string;
    tableName: string;
    dimensions: CreateDimensionRequest[];
    metrics: CreateMetricRequest[];
}

export type UpdateSemanticModelRequest = CreateSemanticModelRequest;

export interface SemanticQueryRequest {
    modelId: string;
    dimensions: string[];
    metrics: string[];
    filters?: Record<string, any>;
    limit?: number;
}

export interface SemanticQueryResponse {
    sql: string;
    args: any[];
    dimensions: string[];
    metrics: string[];
    rowCount: number;
    // We might add actual data results here later if we execute it
}
