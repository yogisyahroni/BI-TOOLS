export interface DataClassification {
    id: number;
    name: string;
    description?: string;
    color: string;
    created_at: string;
    updated_at: string;
}

export interface ColumnMetadata {
    id: number;
    datasource_id: string;
    table_name: string;
    column_name: string;
    data_classification_id?: number;
    data_classification?: DataClassification;
    alias?: string;
    description?: string;
    created_at: string;
    updated_at: string;
}

export interface ColumnPermission {
    id: number;
    role_id: number;
    column_metadata_id: number;
    column_metadata?: ColumnMetadata;
    is_hidden: boolean;
    masking_type: 'none' | 'full' | 'email' | 'last4' | 'partial';
    created_at: string;
    updated_at: string;
}
