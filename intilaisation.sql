CREATE TABLE prod.queries (
	query_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	query_name VARCHAR(100) NOT NULL,
	query TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	created_by VARCHAR(30) NOT NULL DEFAULT 'me',
	updated_by VARCHAR(30) NOT NULL DEFAULT 'me',
	query_text TEXT NOT NULL
);

CREATE TABLE prod.input_parameters (
	parameter_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	query_id INT REFERENCES prod.queries(query_id) ON DELETE CASCADE,
	parameter_name VARCHAR(100) NOT NULL,
	parameter_type VARCHAR(20) NOT NULL
);

CREATE TABLE prod.sources (
	source_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	source_type VARCHAR(20) NOT NULL,
	source_name VARCHAR(100) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	created_by VARCHAR(30) NOT NULL DEFAULT 'me'
);

CREATE TABLE prod.source_credentials (
	id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	datasource_id INT REFERENCES prod.sources(source_id) ON DELETE CASCADE,
	credentials JSONB NOT NULL 
);

CREATE TABLE prod.etl_processes (
	etl_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	status VARCHAR(20) NOT NULL,
	started_at TIMESTAMP,
	ended_at TIMESTAMP,
	full_refresh_time TIMESTAMP,
	incremental_refresh_time TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	created_by VARCHAR(30) NOT NULL DEFAULT 'me'
);

CREATE TABLE prod.etl_queries (
	query_id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	etl_id INTEGER REFERENCES prod.etl_processes(etl_id) ON DELETE CASCADE,
	query_text TEXT NOT NULL
);

CREATE TABLE prod.etl_sources (
	etl_id INTEGER REFERENCES prod.etl_processes(etl_id) ON DELETE CASCADE,
	source_id INTEGER REFERENCES prod.sources(source_id) ON DELETE CASCADE,
	PRIMARY KEY (etl_id, source_id)
);

DROP TABLE IF EXISTS prod.queries;
DROP TABLE IF EXISTS prod.input_parameters;
DROP TABLE IF EXISTS prod.sources;
DROP TABLE IF EXISTS prod.source_credentials;
DROP TABLE IF EXISTS prod.etl_processes;
DROP TABLE IF EXISTS prod.etl_queries;
DROP TABLE IF EXISTS prod.etl_sources;