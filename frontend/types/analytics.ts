export type InsightType = 'trend' | 'anomaly' | 'correlation' | 'descriptive' | 'statistic' | 'forecast';

export interface Insight {
    id: string;
    type: InsightType;
    title: string;
    description: string;
    metric: string;
    value: any;
    confidence: number;
    metadata?: Record<string, any>;
    createdAt: string;
}

export interface CorrelationResult {
    variableA: string;
    variableB: string;
    coefficient: number;
    strength: 'Strong' | 'Moderate' | 'Weak' | 'None';
    significance?: number;
}

export interface GenerateInsightsRequest {
    data: Record<string, any>[];
    metricCol: string;
    timeCol?: string;
}

export interface CalculateCorrelationRequest {
    data: Record<string, any>[];
    cols: string[];
}
