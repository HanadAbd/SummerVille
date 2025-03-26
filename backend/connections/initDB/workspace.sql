CREATE SCHEMA IF NOT EXISTS prod;
CREATE SCHEMA IF NOT EXISTS dev;
CREATE SCHEMA IF NOT EXISTS staging;
CREATE SCHEMA IF NOT EXISTS test;

CREATE TYPE user_role AS ENUM ('admin', 'developer', 'user');

CREATE TABLE IF NOT EXISTS prod.workspaces (
    workspace_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.users (
    user_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    default_workspace_id INTEGER REFERENCES prod.workspaces(workspace_id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prod.workspace_members (
    workspace_id INTEGER REFERENCES prod.workspaces(workspace_id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES prod.users(user_id) ON DELETE CASCADE,
    role user_role NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX idx_workspace_members_user_id ON prod.workspace_members(user_id);
CREATE INDEX idx_workspace_members_workspace_id ON prod.workspace_members(workspace_id);


INSERT INTO prod.workspaces (workspace_name, description) VALUES ('Default Workspace', 'This is the default workspace for all users');
INSERT INTO prod.users (username, email, default_workspace_id) VALUES ('admin', 'email', 1);
