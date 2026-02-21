-- Enable pgvector extension if not exists
CREATE EXTENSION IF NOT EXISTS vector;
-- Create schema_embeddings table to store vector representations of database schemas
CREATE TABLE IF NOT EXISTS schema_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    schema_name TEXT NOT NULL,
    table_name TEXT NOT NULL,
    column_name TEXT,
    -- NULL if this embedding represents the whole table
    description TEXT,
    -- A semantic description generated for the LLM
    data_type TEXT,
    -- NULL if representing a table
    embedding vector(1536) NOT NULL,
    -- 1536 is standard for text-embedding-ada-002 and text-embedding-3-small
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Ensure we don't have exact duplicate embeddings for the same item
    CONSTRAINT unique_schema_item UNIQUE (
        connection_id,
        schema_name,
        table_name,
        column_name
    )
);
-- Essential indexes for vector search and relational joins
CREATE INDEX IF NOT EXISTS schema_embeddings_connection_id_idx ON schema_embeddings(connection_id);
-- HNSW index is the state-of-the-art for vector similarity search in pgvector
CREATE INDEX IF NOT EXISTS schema_embeddings_vector_idx ON schema_embeddings USING hnsw (embedding vector_cosine_ops);