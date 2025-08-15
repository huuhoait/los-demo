-- Database initialization script for Loan Service
-- This script creates the necessary databases and tables

-- Create conductor database for Netflix Conductor
CREATE DATABASE conductor;

-- Connect to loan_service database (created by POSTGRES_DB environment variable)
\c loan_service;

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create loan applications table
CREATE TABLE IF NOT EXISTS loan_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    application_number VARCHAR(50) UNIQUE NOT NULL,
    loan_amount DECIMAL(15,2) NOT NULL,
    loan_purpose VARCHAR(50) NOT NULL,
    requested_term_months INTEGER NOT NULL,
    annual_income DECIMAL(15,2) NOT NULL,
    monthly_income DECIMAL(15,2) NOT NULL,
    employment_status VARCHAR(50) NOT NULL,
    monthly_debt_payments DECIMAL(15,2) NOT NULL,
    current_state VARCHAR(50) NOT NULL DEFAULT 'initiated',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    risk_score INTEGER,
    workflow_id VARCHAR(255),
    correlation_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create loan offers table
CREATE TABLE IF NOT EXISTS loan_offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES loan_applications(id) ON DELETE CASCADE,
    offer_amount DECIMAL(15,2) NOT NULL,
    interest_rate DECIMAL(5,2) NOT NULL,
    term_months INTEGER NOT NULL,
    monthly_payment DECIMAL(15,2) NOT NULL,
    total_interest DECIMAL(15,2) NOT NULL,
    apr DECIMAL(5,2) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create state transitions table
CREATE TABLE IF NOT EXISTS state_transitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES loan_applications(id) ON DELETE CASCADE,
    from_state VARCHAR(50) NOT NULL,
    to_state VARCHAR(50) NOT NULL,
    transition_reason TEXT,
    triggered_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create workflow executions table
CREATE TABLE IF NOT EXISTS workflow_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workflow_id VARCHAR(255) NOT NULL,
    application_id UUID NOT NULL REFERENCES loan_applications(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_loan_applications_user_id ON loan_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_loan_applications_status ON loan_applications(status);
CREATE INDEX IF NOT EXISTS idx_loan_applications_current_state ON loan_applications(current_state);
CREATE INDEX IF NOT EXISTS idx_loan_applications_created_at ON loan_applications(created_at);

CREATE INDEX IF NOT EXISTS idx_loan_offers_application_id ON loan_offers(application_id);
CREATE INDEX IF NOT EXISTS idx_loan_offers_status ON loan_offers(status);

CREATE INDEX IF NOT EXISTS idx_state_transitions_application_id ON state_transitions(application_id);
CREATE INDEX IF NOT EXISTS idx_state_transitions_created_at ON state_transitions(created_at);

CREATE INDEX IF NOT EXISTS idx_workflow_executions_application_id ON workflow_executions(application_id);
CREATE INDEX IF NOT EXISTS idx_workflow_executions_status ON workflow_executions(status);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_loan_applications_updated_at 
    BEFORE UPDATE ON loan_applications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_loan_offers_updated_at 
    BEFORE UPDATE ON loan_offers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_workflow_executions_updated_at 
    BEFORE UPDATE ON workflow_executions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data for testing
INSERT INTO loan_applications (
    user_id, 
    application_number, 
    loan_amount, 
    loan_purpose, 
    requested_term_months, 
    annual_income, 
    monthly_income, 
    employment_status, 
    monthly_debt_payments, 
    current_state, 
    status
) VALUES 
    ('user123', 'LOAN123456', 25000.00, 'debt_consolidation', 60, 75000.00, 6250.00, 'full_time', 1500.00, 'initiated', 'draft'),
    ('user456', 'LOAN789012', 35000.00, 'home_improvement', 48, 85000.00, 7083.33, 'full_time', 2000.00, 'pre_qualified', 'pending')
ON CONFLICT (application_number) DO NOTHING;

-- Grant permissions to postgres user (already done by default)
-- In production, you might want to create a specific user for the application

-- Print success message
SELECT 'Database initialization completed successfully!' as status;
