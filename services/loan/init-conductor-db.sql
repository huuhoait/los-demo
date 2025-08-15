-- Initialize Conductor database
-- This script creates the necessary database for Netflix Conductor

-- Create conductor database
CREATE DATABASE conductor;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE conductor TO postgres;

-- Connect to conductor database and create basic schema
\c conductor;

-- Create basic tables that Conductor will need
-- Note: Conductor will create its own tables on startup, this is just to ensure the database exists

-- Create a basic info table to track initialization
CREATE TABLE IF NOT EXISTS conductor_info (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50),
    initialized_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert initialization record
INSERT INTO conductor_info (version) VALUES ('3.15.0');
