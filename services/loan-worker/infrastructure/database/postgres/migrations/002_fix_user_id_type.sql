-- Migration: 002_fix_user_id_type.sql
-- Description: Fix user_id type mismatch between loan_applications and users tables

-- First, drop the existing foreign key constraint if it exists
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_loan_applications_user_id'
    ) THEN
        ALTER TABLE loan_applications 
        DROP CONSTRAINT fk_loan_applications_user_id;
    END IF;
END $$;

-- Update the user_id column in loan_applications to be UUID type
ALTER TABLE loan_applications 
ALTER COLUMN user_id TYPE UUID USING user_id::uuid;

-- Now add the foreign key constraint with correct types
ALTER TABLE loan_applications 
ADD CONSTRAINT fk_loan_applications_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Update the sample data to use proper UUID format
-- First, delete existing sample data
DELETE FROM loan_applications WHERE user_id IN ('user123', 'user456');

-- Insert a proper user first
INSERT INTO users (
    id,
    first_name, 
    last_name, 
    email, 
    phone_number, 
    date_of_birth, 
    ssn,
    street_address, 
    city, 
    state, 
    zip_code, 
    country, 
    residence_type, 
    time_at_address_months,
    employer_name, 
    job_title, 
    time_employed_months, 
    work_phone, 
    work_email,
    bank_name, 
    account_type, 
    account_number, 
    routing_number
) VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'John', 'Doe', 'john.doe@example.com', '+1234567890', '1990-01-01', '123456789', '123 Main St', 'New York', 'NY', '10001', 'USA', 'own', 24, 'ABC Company', 'Software Engineer', 36, '+1234567890', 'john.doe@abccompany.com', 'Chase Bank', 'checking', '1234567890', '021000021'),
    ('550e8400-e29b-41d4-a716-446655440002', 'Jane', 'Smith', 'jane.smith@example.com', '+1234567891', '1985-05-15', '987654321', '456 Oak Ave', 'Los Angeles', 'CA', '90210', 'USA', 'rent', 12, 'XYZ Corp', 'Product Manager', 24, '+1234567891', 'jane.smith@xyzcorp.com', 'Wells Fargo', 'savings', '0987654321', '121000248')
ON CONFLICT (email) DO NOTHING;

-- Now insert loan applications with proper UUID references
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
    ('550e8400-e29b-41d4-a716-446655440001', 'LOAN123456', 25000.00, 'debt_consolidation', 60, 75000.00, 6250.00, 'full_time', 1500.00, 'initiated', 'draft'),
    ('550e8400-e29b-41d4-a716-446655440002', 'LOAN789012', 35000.00, 'home_improvement', 48, 85000.00, 7083.33, 'full_time', 2000.00, 'pre_qualified', 'pending')
ON CONFLICT (application_number) DO NOTHING;
