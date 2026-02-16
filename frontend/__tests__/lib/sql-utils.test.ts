import { describe, it, expect } from 'vitest'
import {
    formatSQL,
    minifySQL,
    extractTableNames,
    validateSQL,
    getQueryType,
    replaceQueryVariables,
    extractQueryVariables,
} from '@/lib/sql-utils'

// ─────────────────────────────────────────────────────────────────────────────
// formatSQL
// ─────────────────────────────────────────────────────────────────────────────

describe('formatSQL', () => {
    it('should uppercase SQL keywords', () => {
        const result = formatSQL('select id from users')
        expect(result).toContain('SELECT')
        expect(result).toContain('FROM')
    })

    it('should return empty/falsy input unchanged', () => {
        expect(formatSQL('')).toBe('')
        expect(formatSQL('   ')).toBe('   ')
    })

    it('should collapse excess whitespace', () => {
        const result = formatSQL('select   id    from    users')
        expect(result).not.toContain('   ')
    })

    it('should add newlines before major clauses', () => {
        const result = formatSQL('SELECT id FROM users WHERE active = true ORDER BY id')
        const lines = result.split('\n')
        expect(lines.length).toBeGreaterThan(1)
    })

    it('should indent AND/OR in WHERE clause', () => {
        const result = formatSQL('SELECT * FROM users WHERE a = 1 AND b = 2 OR c = 3')
        expect(result).toContain('AND')
        expect(result).toContain('OR')
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// minifySQL
// ─────────────────────────────────────────────────────────────────────────────

describe('minifySQL', () => {
    it('should collapse whitespace and trim', () => {
        expect(minifySQL('  SELECT   id \n  FROM   users  ')).toBe('SELECT id FROM users')
    })

    it('should return falsy input unchanged', () => {
        expect(minifySQL('')).toBe('')
    })

    it('should handle single-line SQL unchanged', () => {
        expect(minifySQL('SELECT 1')).toBe('SELECT 1')
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// extractTableNames
// ─────────────────────────────────────────────────────────────────────────────

describe('extractTableNames', () => {
    it('should extract single table from FROM clause', () => {
        const tables = extractTableNames('SELECT * FROM users')
        expect(tables).toEqual(['users'])
    })

    it('should extract table from JOIN clause', () => {
        const tables = extractTableNames('SELECT * FROM orders JOIN customers ON orders.cid = customers.id')
        expect(tables).toContain('orders')
        expect(tables).toContain('customers')
    })

    it('should deduplicate tables', () => {
        const tables = extractTableNames('SELECT * FROM users JOIN users ON a = b')
        expect(tables).toEqual(['users'])
    })

    it('should return empty array for no tables', () => {
        const tables = extractTableNames('SELECT 1 + 1')
        expect(tables).toEqual([])
    })

    it('should handle multiple joins', () => {
        const tables = extractTableNames(
            'SELECT * FROM orders LEFT JOIN customers ON a = b RIGHT JOIN products ON c = d'
        )
        expect(tables).toContain('orders')
        expect(tables).toContain('customers')
        expect(tables).toContain('products')
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// validateSQL
// ─────────────────────────────────────────────────────────────────────────────

describe('validateSQL', () => {
    it('should reject empty queries', () => {
        const result = validateSQL('')
        expect(result.valid).toBe(false)
        expect(result.error).toBe('Query cannot be empty')
    })

    it('should reject whitespace-only queries', () => {
        const result = validateSQL('   ')
        expect(result.valid).toBe(false)
    })

    it('should accept valid SELECT query', () => {
        const result = validateSQL('SELECT id FROM users')
        expect(result.valid).toBe(true)
        expect(result.error).toBeUndefined()
    })

    it('should warn on DROP statements', () => {
        const result = validateSQL('DROP TABLE users')
        expect(result.valid).toBe(true)
        expect(result.error).toContain('destructive')
    })

    it('should warn on TRUNCATE statements', () => {
        const result = validateSQL('TRUNCATE TABLE users')
        expect(result.valid).toBe(true)
        expect(result.error).toContain('destructive')
    })

    it('should detect unbalanced parentheses', () => {
        const result = validateSQL('SELECT * FROM (SELECT id FROM users')
        expect(result.valid).toBe(false)
        expect(result.error).toBe('Unbalanced parentheses')
    })

    it('should detect unclosed string literals', () => {
        const result = validateSQL("SELECT * FROM users WHERE name = 'hello")
        expect(result.valid).toBe(false)
        expect(result.error).toBe('Unclosed string literal')
    })

    it('should accept SELECT with expression (no FROM)', () => {
        const result = validateSQL('SELECT 1 + 1')
        expect(result.valid).toBe(true)
    })

    it('should accept subquery in SELECT', () => {
        const result = validateSQL('SELECT (SELECT COUNT(*) FROM users)')
        expect(result.valid).toBe(true)
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// getQueryType
// ─────────────────────────────────────────────────────────────────────────────

describe('getQueryType', () => {
    it('should detect SELECT queries', () => {
        expect(getQueryType('SELECT * FROM users')).toBe('SELECT')
    })

    it('should detect WITH (CTE) as SELECT', () => {
        expect(getQueryType('WITH cte AS (SELECT 1) SELECT * FROM cte')).toBe('SELECT')
    })

    it('should detect INSERT queries', () => {
        expect(getQueryType('INSERT INTO users (name) VALUES (\'test\')')).toBe('INSERT')
    })

    it('should detect UPDATE queries', () => {
        expect(getQueryType('UPDATE users SET name = \'new\'')).toBe('UPDATE')
    })

    it('should detect DELETE queries', () => {
        expect(getQueryType('DELETE FROM users WHERE id = 1')).toBe('DELETE')
    })

    it('should detect CREATE as DDL', () => {
        expect(getQueryType('CREATE TABLE users (id INT)')).toBe('DDL')
    })

    it('should detect ALTER as DDL', () => {
        expect(getQueryType('ALTER TABLE users ADD COLUMN age INT')).toBe('DDL')
    })

    it('should detect DROP as DDL', () => {
        expect(getQueryType('DROP TABLE users')).toBe('DDL')
    })

    it('should return OTHER for unknown', () => {
        expect(getQueryType('GRANT ALL ON users TO admin')).toBe('OTHER')
    })

    it('should be case-insensitive', () => {
        expect(getQueryType('select * from users')).toBe('SELECT')
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// replaceQueryVariables
// ─────────────────────────────────────────────────────────────────────────────

describe('replaceQueryVariables', () => {
    it('should replace {{variable}} syntax with string value', () => {
        const result = replaceQueryVariables('SELECT * FROM users WHERE name = {{name}}', {
            name: 'Alice',
        })
        expect(result).toContain("'Alice'")
        expect(result).not.toContain('{{name}}')
    })

    it('should replace :variable syntax with number value', () => {
        const result = replaceQueryVariables('SELECT * FROM users WHERE id = :id', { id: 42 })
        expect(result).toContain('42')
        expect(result).not.toContain(':id')
    })

    it('should replace null values with NULL', () => {
        const result = replaceQueryVariables('SELECT * FROM users WHERE deleted_at = {{deleted}}', {
            deleted: null,
        })
        expect(result).toContain('NULL')
    })

    it('should replace boolean values', () => {
        const result = replaceQueryVariables('SELECT * FROM users WHERE active = {{active}}', {
            active: true,
        })
        expect(result).toContain('TRUE')
    })

    it('should replace false boolean', () => {
        const result = replaceQueryVariables('SELECT * FROM users WHERE active = :active', {
            active: false,
        })
        expect(result).toContain('FALSE')
    })

    it('should escape single quotes in string values', () => {
        const result = replaceQueryVariables("SELECT * FROM users WHERE name = {{name}}", {
            name: "O'Brien",
        })
        expect(result).toContain("O''Brien")
    })

    it('should leave unmatched variables unchanged', () => {
        const result = replaceQueryVariables('SELECT * FROM users WHERE name = {{unknown}}', {})
        expect(result).toContain('{{unknown}}')
    })
})

// ─────────────────────────────────────────────────────────────────────────────
// extractQueryVariables
// ─────────────────────────────────────────────────────────────────────────────

describe('extractQueryVariables', () => {
    it('should extract {{variable}} syntax', () => {
        const vars = extractQueryVariables('SELECT * FROM users WHERE name = {{name}} AND age = {{age}}')
        expect(vars).toContain('name')
        expect(vars).toContain('age')
    })

    it('should extract :variable syntax', () => {
        const vars = extractQueryVariables('SELECT * FROM users WHERE id = :userId')
        expect(vars).toContain('userId')
    })

    it('should deduplicate variables', () => {
        const vars = extractQueryVariables('SELECT {{col}} FROM {{col}}')
        expect(vars).toEqual(['col'])
    })

    it('should return empty array for no variables', () => {
        const vars = extractQueryVariables('SELECT * FROM users')
        expect(vars).toEqual([])
    })

    it('should not extract Postgres type casts (::)', () => {
        const vars = extractQueryVariables("SELECT id::text FROM users")
        expect(vars).not.toContain('text')
    })
})
