-- Create data_classifications table
CREATE TABLE IF NOT EXISTS data_classifications (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    color VARCHAR(20) DEFAULT '#808080',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- Seed default classifications
INSERT INTO data_classifications (name, description, color)
VALUES (
        'Public',
        'Data available to everyone',
        '#22c55e'
    ),
    (
        'Internal',
        'Data for internal use only',
        '#3b82f6'
    ),
    (
        'Confidential',
        'Sensitive business data',
        '#f59e0b'
    ),
    (
        'PII',
        'Personally Identifiable Information',
        '#ef4444'
    ) ON CONFLICT (name) DO NOTHING;
-- Create column_metadata table
CREATE TABLE IF NOT EXISTS column_metadata (
    id SERIAL PRIMARY KEY,
    datasource_id VARCHAR(36) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    column_name VARCHAR(255) NOT NULL,
    data_classification_id INT REFERENCES data_classifications(id),
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_col_meta_datasource ON column_metadata(datasource_id);
CREATE INDEX idx_col_meta_table ON column_metadata(table_name);
CREATE UNIQUE INDEX idx_col_meta_unique ON column_metadata(datasource_id, table_name, column_name);
-- Create column_permissions table
CREATE TABLE IF NOT EXISTS column_permissions (
    id SERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    column_metadata_id INT NOT NULL REFERENCES column_metadata(id) ON DELETE CASCADE,
    is_hidden BOOLEAN DEFAULT FALSE,
    masking_type VARCHAR(20) DEFAULT 'none',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_col_perm_role ON column_permissions(role_id);
CREATE INDEX idx_col_perm_meta ON column_permissions(column_metadata_id);
CREATE UNIQUE INDEX idx_col_perm_unique ON column_permissions(role_id, column_metadata_id);