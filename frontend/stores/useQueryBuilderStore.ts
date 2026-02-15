import { create } from 'zustand';
import {
    VisualQueryConfig,
    TableSelection,
    JoinConfig,
    FilterCondition,
    ColumnSelection,
    SchemaTable,
    Aggregation // Import SchemaTable
} from '@/types/visual-query';
import { nanoid } from 'nanoid';

interface QueryBuilderState {
    config: VisualQueryConfig;

    // Metadata (not persisted directly in config, but used for UI)
    tableSchemas: Record<string, SchemaTable>;

    // Actions
    setTables: (tables: TableSelection[]) => void;
    addTable: (table: TableSelection, schema?: SchemaTable) => void;
    removeTable: (tableName: string) => void;

    addJoin: (join: JoinConfig) => void;
    removeJoin: (index: number) => void;
    updateJoin: (index: number, join: JoinConfig) => void;

    setColumns: (columns: ColumnSelection[]) => void;
    toggleColumn: (column: ColumnSelection) => void;

    addFilter: (filter: Omit<FilterCondition, 'id'>) => void;
    removeFilter: (id: string) => void;
    updateFilter: (id: string, filter: Partial<FilterCondition>) => void;

    setLimit: (limit: number | undefined) => void;
    reset: () => void;
}

const initialConfig: VisualQueryConfig = {
    tables: [],
    joins: [],
    columns: [],
    filters: [],
    aggregations: [],
    groupBy: [],
    orderBy: [],
    limit: 100
};

export const useQueryBuilderStore = create<QueryBuilderState>((set) => ({
    config: initialConfig,
    tableSchemas: {},

    setTables: (tables) => set((state) => ({
        config: { ...state.config, tables }
    })),

    addTable: (table, schema) => set((state) => {
        const newSchemas = { ...state.tableSchemas };
        if (schema) {
            newSchemas[table.name] = schema;
        }
        return {
            config: { ...state.config, tables: [...state.config.tables, table] },
            tableSchemas: newSchemas
        };
    }),

    removeTable: (tableName) => set((state) => {
        const newSchemas = { ...state.tableSchemas };
        delete newSchemas[tableName]; // Optional: keep it cached? No, better clear to avoid staleness if re-added
        return {
            config: {
                ...state.config,
                tables: state.config.tables.filter(t => t.name !== tableName),
                // Also remove related joins and columns
                joins: state.config.joins.filter(j => j.leftTable !== tableName && j.rightTable !== tableName),
                columns: state.config.columns.filter(c => c.table !== tableName)
            },
            tableSchemas: newSchemas
        };
    }),

    addJoin: (join) => set((state) => ({
        config: { ...state.config, joins: [...state.config.joins, join] }
    })),

    removeJoin: (index) => set((state) => {
        const newJoins = [...state.config.joins];
        newJoins.splice(index, 1);
        return { config: { ...state.config, joins: newJoins } };
    }),

    updateJoin: (index, join) => set((state) => {
        const newJoins = [...state.config.joins];
        newJoins[index] = join;
        return { config: { ...state.config, joins: newJoins } };
    }),

    setColumns: (columns) => set((state) => ({
        config: { ...state.config, columns }
    })),

    toggleColumn: (column) => set((state) => {
        const exists = state.config.columns.find(c => c.table === column.table && c.column === column.column);
        let newColumns;
        if (exists) {
            newColumns = state.config.columns.filter(c => !(c.table === column.table && c.column === column.column));
        } else {
            newColumns = [...state.config.columns, column];
        }
        return { config: { ...state.config, columns: newColumns } };
    }),

    addFilter: (filter) => set((state) => ({
        config: {
            ...state.config,
            filters: [...state.config.filters, { ...filter, id: nanoid() }]
        }
    })),

    removeFilter: (id) => set((state) => ({
        config: {
            ...state.config,
            filters: state.config.filters.filter(f => f.id !== id)
        }
    })),

    updateFilter: (id, filterUpdate) => set((state) => ({
        config: {
            ...state.config,
            filters: state.config.filters.map(f => f.id === id ? { ...f, ...filterUpdate } : f)
        }
    })),

    setLimit: (limit) => set((state) => ({
        config: { ...state.config, limit }
    })),

    reset: () => set({ config: initialConfig, tableSchemas: {} })
}));
