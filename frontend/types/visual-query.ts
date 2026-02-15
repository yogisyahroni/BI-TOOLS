export type JoinType = 'INNER' | 'LEFT' | 'RIGHT' | 'FULL';
export type FilterOperator = '=' | '!=' | '>' | '<' | '>=' | '<=' | 'LIKE' | 'IN' | 'BETWEEN';
export type LogicOperator = 'AND' | 'OR';
export type AggregationFunction = 'SUM' | 'AVG' | 'COUNT' | 'MIN' | 'MAX';

export interface SchemaColumn {
    name: string;
    type: string;
}

export interface SchemaTable {
    name: string;
    columns: SchemaColumn[];
}

export type SortDirection = 'ASC' | 'DESC';

export interface TableSelection {
    name: string;
    alias: string;
}

export interface JoinConfig {
    type: JoinType;
    leftTable: string;
    rightTable: string;
    leftColumn: string;
    rightColumn: string;
}

export interface ColumnSelection {
    table: string;
    column: string;
    alias?: string;
    aggregation?: AggregationFunction;
}

export interface FilterCondition {
    id: string; // Frontend-only for React keys
    column: string;
    operator: FilterOperator;
    value: any;
    logic: LogicOperator;
}

export interface Aggregation {
    function: AggregationFunction;
    column: string;
    alias: string;
}

export interface OrderByClause {
    column: string;
    direction: SortDirection;
}

export interface VisualQueryConfig {
    tables: TableSelection[];
    joins: JoinConfig[];
    columns: ColumnSelection[];
    filters: FilterCondition[];
    aggregations: Aggregation[];
    groupBy: string[];
    orderBy: OrderByClause[];
    limit?: number;
    cursor?: string;
}

export interface QueryResult {
    columns: string[];
    rows: any[];
    executionTime?: number;
    rowCount?: number;
    metadata?: any;
    nextCursor?: string;
}
