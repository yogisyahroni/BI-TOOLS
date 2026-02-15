import { format as sqlFormat, FormatOptions as SQLFormatOptions } from 'sql-formatter';

/**
 * SQL Formatter Configuration
 * Wraps sql-formatter library with app-specific defaults
 */

export type SQLDialect =
    | 'postgresql'
    | 'mysql'
    | 'sqlite'
    | 'bigquery'
    | 'snowflake'
    | 'redshift'
    | 'mariadb'
    | 'plsql';

export interface FormatOptions {
    /**
     * SQL dialect to use for formatting
     * @default 'postgresql'
     */
    dialect?: SQLDialect;

    /**
     * Number of spaces for indentation
     * @default 2
     */
    tabWidth?: number;

    /**
     * Use tabs instead of spaces
     * @default false
     */
    useTabs?: boolean;

    /**
     * Convert keywords to uppercase
     * @default true
     */
    keywordCase?: 'upper' | 'lower' | 'preserve';

    /**
     * Convert data types to uppercase
     * @default true
     */
    dataTypeCase?: 'upper' | 'lower' | 'preserve';

    /**
     * Convert function names to uppercase
     * @default false
     */
    functionCase?: 'upper' | 'lower' | 'preserve';

    /**
     * Maximum line length before wrapping
     * @default 80
     */
    linesBetweenQueries?: number;

    /**
     * Add semicolon at the end of statements
     * @default true
     */
    ensureSemicolon?: boolean;
}

/**
 * Default formatting options
 */
const DEFAULT_OPTIONS: FormatOptions = {
    dialect: 'postgresql',
    tabWidth: 2,
    useTabs: false,
    keywordCase: 'upper',
    dataTypeCase: 'upper',
    functionCase: 'preserve',
    linesBetweenQueries: 1,
    ensureSemicolon: true,
};

/**
 * Format SQL query with configurable options
 *
 * @param sql - SQL query string to format
 * @param options - Formatting options
 * @returns Formatted SQL string
 *
 * @example
 * ```typescript
 * const formatted = formatSQL('select * from users where id=1');
 * // Returns:
 * // SELECT
 * //   *
 * // FROM
 * //   users
 * // WHERE
 * //   id = 1;
 * ```
 */
export function formatSQL(sql: string, options: FormatOptions = {}): string {
    const mergedOptions = { ...DEFAULT_OPTIONS, ...options };

    try {
        // sql-formatter accepts these options
        const formatted = sqlFormat(sql, {
            language: mergedOptions.dialect || 'postgresql',
            tabWidth: mergedOptions.tabWidth,
            useTabs: mergedOptions.useTabs,
            keywordCase: mergedOptions.keywordCase,
            dataTypeCase: mergedOptions.dataTypeCase,
            functionCase: mergedOptions.functionCase,
            linesBetweenQueries: mergedOptions.linesBetweenQueries,
        });

        // Ensure semicolon at the end if requested
        if (mergedOptions.ensureSemicolon && !formatted.trim().endsWith(';')) {
            return formatted.trim() + ';';
        }

        return formatted;
    } catch (error) {
        console.error('SQL formatting error:', error);
        // Return original SQL if formatting fails
        return sql;
    }
}

/**
 * Format SQL with PostgreSQL-specific settings
 */
export function formatPostgreSQL(sql: string, options: Partial<FormatOptions> = {}): string {
    return formatSQL(sql, {
        ...options,
        dialect: 'postgresql',
    });
}

/**
 * Format SQL with MySQL-specific settings
 */
export function formatMySQL(sql: string, options: Partial<FormatOptions> = {}): string {
    return formatSQL(sql, {
        ...options,
        dialect: 'mysql',
    });
}

/**
 * Format SQL with SQLite-specific settings
 */
export function formatSQLite(sql: string, options: Partial<FormatOptions> = {}): string {
    return formatSQL(sql, {
        ...options,
        dialect: 'sqlite',
    });
}

/**
 * Format SQL with BigQuery-specific settings
 */
export function formatBigQuery(sql: string, options: Partial<FormatOptions> = {}): string {
    return formatSQL(sql, {
        ...options,
        dialect: 'bigquery',
    });
}

/**
 * Compact SQL formatter - single line, minimal whitespace
 * Useful for logging or display in constrained spaces
 */
export function compactSQL(sql: string): string {
    return sql
        .replace(/\s+/g, ' ') // Replace multiple spaces with single space
        .replace(/\s*([,;()=<>])\s*/g, '$1') // Remove spaces around operators
        .trim();
}

/**
 * Validate if SQL can be parsed/formatted
 * Returns true if sql-formatter can handle it
 */
export function canFormat(sql: string): boolean {
    try {
        formatSQL(sql);
        return true;
    } catch {
        return false;
    }
}

/**
 * Extract tables from formatted SQL (simple regex-based)
 * Note: This is a simple heuristic, not a full SQL parser
 */
export function extractTables(sql: string): string[] {
    const tables = new Set<string>();
    const fromRegex = /FROM\s+([a-zA-Z_][a-zA-Z0-9_]*)/gi;
    const joinRegex = /JOIN\s+([a-zA-Z_][a-zA-Z0-9_]*)/gi;

    let match;
    while ((match = fromRegex.exec(sql)) !== null) {
        tables.add(match[1]);
    }
    while ((match = joinRegex.exec(sql)) !== null) {
        tables.add(match[1]);
    }

    return Array.from(tables);
}
