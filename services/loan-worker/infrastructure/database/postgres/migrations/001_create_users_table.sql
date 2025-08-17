-- Migration: 001_create_users_table.sql
-- Description: Create users table for loan applications

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    date_of_birth DATE NOT NULL,
    ssn VARCHAR(9) NOT NULL,
    
    -- Address information
    street_address VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(50) NOT NULL,
    zip_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL DEFAULT 'USA',
    residence_type VARCHAR(50) NOT NULL,
    time_at_address_months INTEGER NOT NULL DEFAULT 0,
    
    -- Employment information
    employer_name VARCHAR(255) NOT NULL,
    job_title VARCHAR(255) NOT NULL,
    time_employed_months INTEGER NOT NULL DEFAULT 0,
    work_phone VARCHAR(20) NOT NULL,
    work_email VARCHAR(255),
    
    -- Banking information
    bank_name VARCHAR(255) NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    routing_number VARCHAR(20) NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_ssn ON users(ssn);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add user_id foreign key constraint to loan_applications
ALTER TABLE loan_applications 
ADD CONSTRAINT fk_loan_applications_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
