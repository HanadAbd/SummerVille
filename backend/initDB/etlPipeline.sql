CREATE SCHEMA IF NOT EXISTS prod;
CREATE SCHEMA IF NOT EXISTS dev;
CREATE SCHEMA IF NOT EXISTS staging;
CREATE SCHEMA IF NOT EXISTS test;

CREATE TABLE IF NOT EXISTS prod.etl_pipeline (
    pipeline_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.etl_pipeline_metadata(
    pipeline_metadata_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pipeline_id INT NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pipeline_id) REFERENCES prod.etl_pipeline(pipeline_id)
);

CREATE TABLE IF NOT EXISTS prod.etl_steps(
    step_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pipeline_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    query TEXT,
    step_order INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pipeline_id) REFERENCES prod.etl_pipeline(pipeline_id)
);

CREATE TABLE IF NOT EXISTS prod.etl_step_metadata(
    step_metadata_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    step_id INT NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (step_id) REFERENCES prod.etl_steps(step_id)
);

CREATE TABLE IF NOT EXISTS prod.etl_pipeline_data_sources(
    pipeline_data_source_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    pipeline_id INT NOT
    data_source_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    schema BLOB,
    data_source_type_id int,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (data_source_type_id) REFERENCES prod.data_source_types(data_source_type_id)
)

CREATE TABLE IF NOT EXISTS prod.data_source_types(
    data_source_type_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
)

CREATE TABLE IF NOT EXISTS prod.data_sources_conditions(
    data_source_condition_id GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    data_source_id INT NOT NULL,
    refresh_interval INT NOT NULL,
    append_only BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (data_source_id) REFERENCES prod.data_sources(data_source_id)
)