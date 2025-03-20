CREATE SCHEMA IF NOT EXISTS prod;
CREATE SCHEMA IF NOT EXISTS dev;
CREATE SCHEMA IF NOT EXISTS staging;
CREATE SCHEMA IF NOT EXISTS test;

CREATE TABLE IF NOT EXISTS prod.etl_pipeline (
    pipeline_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    workspace_id INTEGER REFERENCES prod.workspaces(workspace_id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.etl_pipeline_metadata(
    pipeline_metadata_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pipeline_id INTEGER REFERENCES prod.etl_pipeline(pipeline_id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.etl_steps(
    step_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pipeline_id INTEGER REFERENCES prod.etl_pipeline(pipeline_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query TEXT,
    step_order INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.etl_step_metadata(
    step_metadata_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    step_id INTEGER REFERENCES prod.etl_steps(step_id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.data_sources(
    data_source_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema JSONB,
    data_source_type_id INTEGER REFERENCES prod.data_source_types(data_source_type_id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.data_source_types(
    data_source_type_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS prod.data_sources_conditions(
    data_source_condition_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    data_source_id INTEGER REFERENCES prod.data_sources(data_source_id) ON DELETE CASCADE,
    refresh_interval INT NOT NULL,
    append_only BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_data_sources_workspace_id ON prod.data_sources(workspace_id);
CREATE INDEX idx_etl_pipeline_workspace_id ON prod.etl_pipeline(workspace_id);
