CREATE SCHEMA IF NOT EXISTS prod;
CREATE SCHEMA IF NOT EXISTS dev;
CREATE SCHEMA IF NOT EXISTS staging;
CREATE SCHEMA IF NOT EXISTS test;

CREATE TYPE etl_layers AS ENUM ('raw', 'staging', 'transformed', 'final');


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
    layer etl_layers NOT NULL,
    step_order INT NOT NULL,
    child_step_id INTEGER REFERENCES prod.etl_steps(step_id) ON DELETE SET NULL,
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
    pipeline_id INTEGER REFERENCES prod.etl_pipeline(pipeline_id) ON DELETE CASCADE,
    step_child_id INTEGER REFERENCES prod.etl_steps(step_id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    layer etl_layers NOT NULL DEFAULT 'raw',
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


INSERT INTO prod.etl_pipeline (name, description, workspace_id) VALUES ('Default Pipeline', 'This is the default pipeline for all users', 1);

INSERT INTO prod.data_sources (pipeline_id, name, description, schema, data_source_type_id) VALUES (1, 'Default Data Source', 'This is the default data source for all users', '{"type": "file", "path": "/path/to/file"}', 1);

INSERT INTO prod.data_source_types (name, description) VALUES ('file', 'File data source type');

INSERT INTO prod.data_sources_conditions (data_source_id, refresh_interval, append_only) VALUES (1, 60, TRUE);

INSERT INTO prod.etl_steps (pipeline_id, name, description, query, layer, step_order) VALUES (1, 'Default Step', 'This is the default step for all users', 'SELECT * FROM data_source', 'raw', 1);

INSERT INTO prod.etl_step_metadata (step_id, key, value) VALUES (1, 'type', 'sql');

INSERT INTO prod.etl_steps (pipeline_id, name, description, query, layer, step_order, child_step_id) VALUES (1, 'Default Step 2', 'This is the default step 2 for all users', 'SELECT * FROM step_1', 'staging', 2, 1);

INSERT INTO prod.etl_step_metadata (step_id, key, value) VALUES (2, 'type', 'sql');

INSERT INTO prod.etl_steps (pipeline_id, name, description, query, layer, step_order, child_step_id) VALUES (1, 'Default Step 3', 'This is the default step 3 for all users', 'SELECT * FROM step_2', 'transformed', 3, 2);

INSERT INTO prod.etl_step_metadata (step_id, key, value) VALUES (3, 'type', 'sql');

INSERT INTO prod.etl_steps (pipeline_id, name, description, query, layer, step_order, child_step_id) VALUES (1, 'Default Step 4', 'This is the default step 4 for all users', 'SELECT * FROM step_3', 'final', 4, 3);

INSERT INTO prod.etl_step_metadata (step_id, key, value) VALUES (4, 'type', 'sql');
