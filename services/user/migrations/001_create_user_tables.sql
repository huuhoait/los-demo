-- User Service Database Schema
-- This file contains all table definitions for the User Service microservice

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create custom types
CREATE TYPE user_role AS ENUM ('customer', 'agent', 'underwriter', 'admin', 'super_admin');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'pending_verification');
CREATE TYPE kyc_status AS ENUM ('not_started', 'in_progress', 'pending_review', 'approved', 'rejected', 'expired');
CREATE TYPE verification_type AS ENUM ('identity', 'address', 'employment', 'income', 'assets');
CREATE TYPE document_type AS ENUM ('passport', 'driver_license', 'national_id', 'utility_bill', 'bank_statement', 'employment_letter', 'pay_stub', 'tax_return', 'other');
CREATE TYPE marital_status AS ENUM ('single', 'married', 'divorced', 'widowed', 'separated');
CREATE TYPE employment_type AS ENUM ('full_time', 'part_time', 'contract', 'self_employed', 'unemployed', 'retired', 'student');

-- Users table - Core user entity
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'customer',
    status user_status DEFAULT 'pending_verification',
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    date_of_birth DATE,
    
    -- Security and verification
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    last_login_at TIMESTAMP,
    login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP,
    email_verification_token VARCHAR(255),
    email_verification_expires TIMESTAMP,
    phone_verification_code VARCHAR(10),
    phone_verification_expires TIMESTAMP,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    version INTEGER DEFAULT 1,
    
    -- Indexes will be created separately
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT valid_phone CHECK (phone IS NULL OR phone ~* '^\+?[1-9]\d{1,14}$')
);

-- User profiles table - Extended user information
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Personal information
    middle_name VARCHAR(100),
    suffix VARCHAR(10),
    preferred_name VARCHAR(100),
    gender VARCHAR(20),
    nationality VARCHAR(3), -- ISO 3166-1 alpha-3
    marital_status marital_status,
    dependents INTEGER DEFAULT 0,
    
    -- Contact information (encrypted)
    primary_address_encrypted TEXT,
    secondary_address_encrypted TEXT,
    emergency_contact_encrypted TEXT,
    
    -- Professional information
    employment_type employment_type,
    employer_name_encrypted TEXT,
    job_title_encrypted TEXT,
    work_address_encrypted TEXT,
    annual_income_encrypted TEXT,
    monthly_income_encrypted TEXT,
    employment_start_date DATE,
    years_at_current_job INTEGER,
    previous_employer_encrypted TEXT,
    
    -- Financial information (encrypted)
    bank_account_encrypted TEXT,
    credit_score INTEGER,
    credit_score_date DATE,
    assets_encrypted TEXT,
    liabilities_encrypted TEXT,
    monthly_expenses_encrypted TEXT,
    
    -- Preferences
    language_preference VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    currency_preference VARCHAR(3) DEFAULT 'USD',
    communication_preferences JSONB DEFAULT '{}',
    marketing_consent BOOLEAN DEFAULT FALSE,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 1,
    
    UNIQUE(user_id)
);

-- KYC verifications table - Know Your Customer verification records
CREATE TABLE kyc_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    verification_type verification_type NOT NULL,
    status kyc_status DEFAULT 'not_started',
    
    -- Verification details
    provider VARCHAR(100), -- External KYC provider used
    provider_reference VARCHAR(255), -- Provider's reference ID
    verification_data_encrypted TEXT, -- Encrypted verification response
    confidence_score DECIMAL(5,4), -- 0.0000 to 1.0000
    risk_score DECIMAL(5,4), -- 0.0000 to 1.0000
    
    -- Review information
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    rejection_reason VARCHAR(500),
    
    -- Expiration
    expires_at TIMESTAMP,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 1,
    
    UNIQUE(user_id, verification_type)
);

-- Documents table - User uploaded documents
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kyc_verification_id UUID REFERENCES kyc_verifications(id) ON DELETE SET NULL,
    
    -- Document details
    document_type document_type NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    storage_key VARCHAR(500) NOT NULL, -- S3 key or file path
    storage_bucket VARCHAR(100),
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_hash VARCHAR(64), -- SHA-256 hash for integrity
    
    -- Encryption
    encryption_key_id VARCHAR(255), -- Reference to encryption key
    is_encrypted BOOLEAN DEFAULT TRUE,
    
    -- Processing status
    processing_status VARCHAR(50) DEFAULT 'uploaded', -- uploaded, processing, processed, failed
    ocr_text TEXT, -- Extracted text from OCR
    extracted_data JSONB, -- Structured data extracted from document
    
    -- Security and compliance
    virus_scan_status VARCHAR(20) DEFAULT 'pending', -- pending, clean, infected
    virus_scan_date TIMESTAMP,
    compliance_flags JSONB DEFAULT '[]',
    
    -- Metadata
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    version INTEGER DEFAULT 1,
    
    CONSTRAINT valid_file_size CHECK (file_size > 0 AND file_size <= 52428800), -- Max 50MB
    CONSTRAINT valid_content_type CHECK (content_type IN (
        'image/jpeg', 'image/png', 'image/gif', 'image/webp',
        'application/pdf', 'application/msword', 
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
    ))
);

-- User sessions table - Track user sessions for security
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE,
    
    -- Session details
    ip_address INET,
    user_agent TEXT,
    device_info JSONB,
    location_info JSONB,
    
    -- Session lifecycle
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    revoked_by UUID REFERENCES users(id),
    revoke_reason VARCHAR(255),
    
    -- Security flags
    is_suspicious BOOLEAN DEFAULT FALSE,
    risk_score DECIMAL(5,4) DEFAULT 0.0000
);

-- User preferences table - Additional user configuration
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference_key VARCHAR(100) NOT NULL,
    preference_value TEXT,
    preference_type VARCHAR(20) DEFAULT 'string', -- string, number, boolean, json
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(user_id, preference_key)
);

-- Audit log table - Track all user-related activities
CREATE TABLE user_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Event details
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL, -- authentication, profile, security, document, kyc
    event_description TEXT,
    
    -- Context
    ip_address INET,
    user_agent TEXT,
    session_id UUID REFERENCES user_sessions(id),
    request_id VARCHAR(100),
    
    -- Data changes
    old_values JSONB,
    new_values JSONB,
    affected_fields TEXT[],
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    severity VARCHAR(20) DEFAULT 'info', -- debug, info, warning, error, critical
    source VARCHAR(50) DEFAULT 'user-service'
);

-- Create indexes for performance

-- Users table indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_users_phone_verified ON users(phone_verified);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_last_login_at ON users(last_login_at);

-- User profiles table indexes
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_employment_type ON user_profiles(employment_type);
CREATE INDEX idx_user_profiles_marital_status ON user_profiles(marital_status);

-- KYC verifications table indexes
CREATE INDEX idx_kyc_verifications_user_id ON kyc_verifications(user_id);
CREATE INDEX idx_kyc_verifications_type ON kyc_verifications(verification_type);
CREATE INDEX idx_kyc_verifications_status ON kyc_verifications(status);
CREATE INDEX idx_kyc_verifications_provider ON kyc_verifications(provider);
CREATE INDEX idx_kyc_verifications_expires_at ON kyc_verifications(expires_at);

-- Documents table indexes
CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_documents_kyc_verification_id ON documents(kyc_verification_id);
CREATE INDEX idx_documents_type ON documents(document_type);
CREATE INDEX idx_documents_storage_key ON documents(storage_key);
CREATE INDEX idx_documents_processing_status ON documents(processing_status);
CREATE INDEX idx_documents_uploaded_at ON documents(uploaded_at);

-- User sessions table indexes
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_refresh_token ON user_sessions(refresh_token);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_user_sessions_ip_address ON user_sessions(ip_address);
CREATE INDEX idx_user_sessions_created_at ON user_sessions(created_at);

-- User preferences table indexes
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX idx_user_preferences_key ON user_preferences(preference_key);

-- Audit logs table indexes
CREATE INDEX idx_user_audit_logs_user_id ON user_audit_logs(user_id);
CREATE INDEX idx_user_audit_logs_actor_id ON user_audit_logs(actor_id);
CREATE INDEX idx_user_audit_logs_event_type ON user_audit_logs(event_type);
CREATE INDEX idx_user_audit_logs_event_category ON user_audit_logs(event_category);
CREATE INDEX idx_user_audit_logs_created_at ON user_audit_logs(created_at);
CREATE INDEX idx_user_audit_logs_severity ON user_audit_logs(severity);
CREATE INDEX idx_user_audit_logs_session_id ON user_audit_logs(session_id);

-- Create functions for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for automatic timestamp updates
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kyc_verifications_updated_at BEFORE UPDATE ON kyc_verifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON user_preferences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function for session cleanup
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM user_sessions 
    WHERE expires_at < CURRENT_TIMESTAMP 
       OR revoked_at IS NOT NULL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ language 'plpgsql';

-- Create function for KYC verification expiration
CREATE OR REPLACE FUNCTION expire_kyc_verifications()
RETURNS INTEGER AS $$
DECLARE
    expired_count INTEGER;
BEGIN
    UPDATE kyc_verifications 
    SET status = 'expired'
    WHERE status = 'approved' 
      AND expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS expired_count = ROW_COUNT;
    RETURN expired_count;
END;
$$ language 'plpgsql';

-- Grant permissions (adjust as needed for your environment)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO user_service_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO user_service_user;
-- GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO user_service_user;

-- Comments for documentation
COMMENT ON TABLE users IS 'Core user accounts with authentication and basic profile information';
COMMENT ON TABLE user_profiles IS 'Extended user profile information including personal, professional, and financial details';
COMMENT ON TABLE kyc_verifications IS 'Know Your Customer verification records for compliance and risk management';
COMMENT ON TABLE documents IS 'User-uploaded documents with encryption and processing metadata';
COMMENT ON TABLE user_sessions IS 'Active user sessions for security tracking and management';
COMMENT ON TABLE user_preferences IS 'User-specific configuration and preference settings';
COMMENT ON TABLE user_audit_logs IS 'Comprehensive audit trail for all user-related activities';

COMMENT ON COLUMN users.salt IS 'Cryptographic salt used for password hashing';
COMMENT ON COLUMN users.two_factor_secret IS 'TOTP secret for two-factor authentication';
COMMENT ON COLUMN user_profiles.primary_address_encrypted IS 'AES encrypted primary residential address';
COMMENT ON COLUMN user_profiles.annual_income_encrypted IS 'AES encrypted annual income information';
COMMENT ON COLUMN documents.encryption_key_id IS 'Reference to the encryption key used for file encryption';
COMMENT ON COLUMN documents.file_hash IS 'SHA-256 hash for file integrity verification';

-- Insert sample data for testing (uncomment if needed)
/*
INSERT INTO users (email, username, password_hash, salt, first_name, last_name, role, status) VALUES
('admin@example.com', 'admin', 'hashed_password_here', 'salt_here', 'System', 'Administrator', 'admin', 'active'),
('test@example.com', 'testuser', 'hashed_password_here', 'salt_here', 'Test', 'User', 'customer', 'active');
*/
