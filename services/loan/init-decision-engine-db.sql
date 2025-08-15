-- Initialize Decision Engine database
-- This script creates the necessary database and tables for the Decision Engine Service

-- Create decision_engine database
CREATE DATABASE decision_engine;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE decision_engine TO postgres;

-- Connect to decision_engine database
\c decision_engine;

-- Create decision_requests table
CREATE TABLE IF NOT EXISTS decision_requests (
    id BIGSERIAL PRIMARY KEY,
    application_id VARCHAR(255) UNIQUE NOT NULL,
    customer_id VARCHAR(255) NOT NULL,
    loan_amount DECIMAL(15,2) NOT NULL,
    loan_purpose VARCHAR(100) NOT NULL,
    loan_term_months INTEGER NOT NULL,
    annual_income DECIMAL(15,2) NOT NULL,
    monthly_income DECIMAL(15,2) NOT NULL,
    credit_score INTEGER NOT NULL,
    employment_type VARCHAR(50) NOT NULL,
    requested_amount DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create indexes for decision_requests
CREATE INDEX IF NOT EXISTS idx_decision_requests_customer_id ON decision_requests(customer_id);
CREATE INDEX IF NOT EXISTS idx_decision_requests_application_id ON decision_requests(application_id);
CREATE INDEX IF NOT EXISTS idx_decision_requests_created_at ON decision_requests(created_at);

-- Create decisions table
CREATE TABLE IF NOT EXISTS decisions (
    id BIGSERIAL PRIMARY KEY,
    application_id VARCHAR(255) NOT NULL,
    decision VARCHAR(20) NOT NULL,
    confidence_score DECIMAL(5,4) NOT NULL,
    interest_rate DECIMAL(5,4),
    max_amount DECIMAL(15,2),
    reason TEXT,
    risk_assessment JSONB,
    applied_rules JSONB,
    recommendations JSONB,
    decision_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create indexes for decisions
CREATE INDEX IF NOT EXISTS idx_decisions_application_id ON decisions(application_id);
CREATE INDEX IF NOT EXISTS idx_decisions_decision ON decisions(decision);
CREATE INDEX IF NOT EXISTS idx_decisions_decision_date ON decisions(decision_date);
CREATE INDEX IF NOT EXISTS idx_decisions_created_at ON decisions(created_at);

-- Add foreign key constraint
ALTER TABLE decisions 
ADD CONSTRAINT fk_decisions_application_id 
FOREIGN KEY (application_id) REFERENCES decision_requests(application_id);

-- Create info table to track initialization
CREATE TABLE IF NOT EXISTS decision_engine_info (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50),
    initialized_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert initialization record
INSERT INTO decision_engine_info (version) VALUES ('1.0.0');
