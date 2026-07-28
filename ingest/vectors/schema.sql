-- Enables the pgvector extension and adds the embeddings table, reusing the
-- same Postgres instance as the rule-engine schema (schema.sql). We use one
-- generic table for both course descriptions and FAQ answers rather than two
-- separate tables, since both are queried the same way: "find nearest by
-- meaning" -- source_type distinguishes them for filtering.

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE embeddings (
    id              SERIAL PRIMARY KEY,
    source_type     TEXT NOT NULL,        -- 'course_description' or 'faq'
    source_id       TEXT NOT NULL,        -- course code, or a stable FAQ id
    content         TEXT NOT NULL,        -- the exact text that was embedded
    embedding       VECTOR(384) NOT NULL  -- all-MiniLM-L6-v2 output dimension
);

-- Small corpus (~180 rows total: 146 courses + 31 FAQs) -- brute-force
-- cosine search is fast enough here, no ANN index needed yet. If this
-- grows into the thousands, add:
--   CREATE INDEX ON embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 10);
CREATE INDEX idx_embeddings_source_type ON embeddings(source_type);