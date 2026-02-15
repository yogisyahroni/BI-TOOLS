/**
 * Query Builder Type Definitions
 * 
 * Defines the complete state shape for the visual query builder
 */

export type AggregationFunction =
    | 'SUM'
    | 'AVG'
    | 'COUNT'
    | 'MIN'
    | 'MAX'
    | 'COUNT_DISTINCT';

export type ComparisonOperator =
    | '='
    | '!='
    | '>'
    | '<'
    | '>='
    | '<='
    | 'LIKE'
    | 'IN'
    | 'NOT IN'
    | 'IS NULL'
    | 'IS NOT NULL'
    | 'BETWEEN';

export type LogicalOperator = 'AND' | 'OR';

export type SortDirection = 'ASC' | 'DESC';

/**
 * Represents a selected column in the query
 */
export interface ColumnSelection {
    table: string;
    column: string;
    alias?: string;
    aggregation?: AggregationFunction;
}

/**
 * Represents a single filter condition
 */
export interface FilterCondition {
    id: string; // For React keys
    column: string;
    operator: ComparisonOperator;
    value: string | number | string[] | null;
}

/**
 * Represents a group of filters with logical operator
 * Supports nesting for complex AND/OR logic
 */
export interface FilterGroup {
    id: string; // For React keys
    operator: LogicalOperator;
    conditions: (FilterCondition | FilterGroup)[];
}

/**
 * Represents a sort rule
 */
export interface SortRule {
    id: string; // For React keys
    column: string;
    direction: SortDirection;
}

/**
 * Complete query builder state
 */
export interface QueryBuilderState {
    connectionId: string;
    table: string | null;
    columns: ColumnSelection[];
    filters: FilterGroup;
    sorts: SortRule[];
    limit: number;
}

/**
 * Represents a table selection in visual query
 */
export interface TableSelection {
    id: string; // For React keys
    name: string;
    schema?: string;
    alias?: string;
    position?: { x: number; y: number }; // Canvas position
}

/**
 * Join types supported
 */
export type JoinType = 'INNER' | 'LEFT' | 'RIGHT' | 'FULL OUTER';

/**
 * Represents a join configuration
 */
export interface JoinConfig {
    id: string; // For React keys
    type: JoinType;
    leftTable: string;
    leftColumn: string;
    rightTable: string;
    rightColumn: string;
    confidence?: 'high' | 'medium' | 'low'; // For auto-suggested joins
}

/**
 * Complete visual query configuration
 * Extends QueryBuilderState with tables, joins, and aggregations
 */
export interface VisualQueryConfig {
    connectionId: string;
    tables: TableSelection[];
    joins: JoinConfig[];
    columns: ColumnSelection[];
    filters: FilterGroup;
    sorts: SortRule[];
    groupBy: string[];
    aggregations: ColumnSelection[];
    having: FilterGroup;
    limit: number;
}

/**
 * Saved visual query metadata
 */
export interface SavedVisualQuery {
    id: string;
    name: string;
    description?: string;
    config: VisualQueryConfig;
    workspaceId: string;
    createdBy: string;
    createdAt: string;
    updatedAt: string;
    tags?: string[];
}

/**
 * Schema information for a table column
 */
export interface ColumnSchema {
    name: string;
    type: 'string' | 'number' | 'date' | 'boolean' | 'unknown';
    nullable: boolean;
}

/**
 * Schema information for a database table
 */
export interface TableSchema {
    name: string;
    schema?: string;
    columns: ColumnSchema[];
    rowCount?: number;
}

/**
 * Database schema (all tables)
 */
export interface DatabaseSchema {
    tables: TableSchema[];
}

/**
 * Query execution result
 */
export interface QueryResult {
    success: boolean;
    data?: any[];
    columns?: string[];
    rowCount?: number;
    executionTime?: number;
    error?: string;
}

/**
 * Helper to create initial empty state
 */
export function createInitialState(connectionId: string): QueryBuilderState {
    return {
        connectionId,
        table: null,
        columns: [],
        filters: {
            id: 'root',
            operator: 'AND',
            conditions: [],
        },
        sorts: [],
        limit: 100,
    };
}

/**
 * Helper to create initial visual query config
 */
export function createInitialVisualQueryConfig(connectionId: string): VisualQueryConfig {
    return {
        connectionId,
        tables: [],
        joins: [],
        columns: [],
        filters: {
            id: 'root-filters',
            operator: 'AND',
            conditions: [],
        },
        sorts: [],
        groupBy: [],
        aggregations: [],
        having: {
            id: 'root-having',
            operator: 'AND',
            conditions: [],
        },
        limit: 100,
    };
}

/**
 * Helper to check if a condition is a group
 */
export function isFilterGroup(
    condition: FilterCondition | FilterGroup
): condition is FilterGroup {
    return 'conditions' in condition;
}

/**
 * Helper to check if a condition is a simple condition
 */
export function isFilterCondition(
    condition: FilterCondition | FilterGroup
): condition is FilterCondition {
    return 'column' in condition;
}
