-- Migration: 000_create_loan_applications_table.sql
-- Description: Create loan_applications table for loan applications
-- This migration must run before 001_create_users_table.sql

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create loan_applications table
CREATE TABLE IF NOT EXISTS loan_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    application_number VARCHAR(50) UNIQUE NOT NULL,
    loan_amount DECIMAL(15,2) NOT NULL,
    loan_purpose VARCHAR(100) NOT NULL,
    requested_term_months INTEGER NOT NULL,
    annual_income DECIMAL(15,2) NOT NULL,
    monthly_income DECIMAL(15,2) NOT NULL,
    employment_status VARCHAR(50) NOT NULL,
    monthly_debt_payments DECIMAL(15,2) NOT NULL,
    current_state VARCHAR(50) NOT NULL DEFAULT 'initiated',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    risk_score INTEGER,
    workflow_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_loan_applications_user_id ON loan_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_loan_applications_application_number ON loan_applications(application_number);
CREATE INDEX IF NOT EXISTS idx_loan_applications_current_state ON loan_applications(current_state);
CREATE INDEX IF NOT EXISTS idx_loan_applications_status ON loan_applications(status);
CREATE INDEX IF NOT EXISTS idx_loan_applications_created_at ON loan_applications(created_at);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_loan_applications_updated_at 
    BEFORE UPDATE ON loan_applications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create state_transitions table for tracking state changes
CREATE TABLE IF NOT EXISTS state_transitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL,
    from_state VARCHAR(50),
    to_state VARCHAR(50) NOT NULL,
    transition_reason TEXT NOT NULL,
    automated BOOLEAN NOT NULL DEFAULT true,
    user_id UUID,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for state_transitions
CREATE INDEX IF NOT EXISTS idx_state_transitions_application_id ON state_transitions(application_id);
CREATE INDEX IF NOT EXISTS idx_state_transitions_to_state ON state_transitions(to_state);
CREATE INDEX IF NOT EXISTS idx_state_transitions_created_at ON state_transitions(created_at);

-- Create loan_offers table
CREATE TABLE IF NOT EXISTS loan_offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL,
    offer_amount DECIMAL(15,2) NOT NULL,
    interest_rate DECIMAL(5,2) NOT NULL,
    term_months INTEGER NOT NULL,
    monthly_payment DECIMAL(15,2) NOT NULL,
    total_interest DECIMAL(15,2) NOT NULL,
    apr DECIMAL(5,2) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for loan_offers
CREATE INDEX IF NOT EXISTS idx_loan_offers_application_id ON loan_offers(application_id);
CREATE INDEX IF NOT EXISTS idx_loan_offers_status ON loan_offers(status);
CREATE INDEX IF NOT EXISTS idx_loan_offers_expires_at ON loan_offers(expires_at);

-- Create trigger to automatically update updated_at for loan_applications
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';
