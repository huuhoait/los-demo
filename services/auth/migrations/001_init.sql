-- Authentication Service Database Schema
-- PostgreSQL DDL for LOS Authentication Service

-- Extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'applicant',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT users_email_check CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT users_role_check CHECK (role IN ('applicant', 'junior_reviewer', 'senior_reviewer', 'manager', 'admin')),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'inactive', 'locked', 'suspended'))
);

-- User sessions table
CREATE TABLE user_sessions (
    id VARCHAR(50) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    
    -- Indexes
    INDEX idx_user_sessions_user_id (user_id),
    INDEX idx_user_sessions_refresh_token (refresh_token),
    INDEX idx_user_sessions_expires_at (expires_at)
);

-- Authentication events audit table
CREATE TABLE auth_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    session_id VARCHAR(50),
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    error_code VARCHAR(20),
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Indexes
    INDEX idx_auth_events_user_id (user_id),
    INDEX idx_auth_events_event_type (event_type),
    INDEX idx_auth_events_created_at (created_at),
    INDEX idx_auth_events_success (success),
    
    -- Constraints
    CONSTRAINT auth_events_event_type_check CHECK (
        event_type IN ('login', 'logout', 'refresh', 'failed_login', 'logout_all')
    )
);

-- Security events audit table
CREATE TABLE security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(50) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    description TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Indexes
    INDEX idx_security_events_event_type (event_type),
    INDEX idx_security_events_user_id (user_id),
    INDEX idx_security_events_severity (severity),
    INDEX idx_security_events_created_at (created_at),
    
    -- Constraints
    CONSTRAINT security_events_severity_check CHECK (
        severity IN ('low', 'medium', 'high', 'critical')
    )
);

-- Failed login attempts tracking table
CREATE TABLE failed_login_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identifier VARCHAR(255) NOT NULL, -- IP address or user_id
    attempts_count INTEGER NOT NULL DEFAULT 1,
    first_attempt_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_attempt_at TIMESTAMP NOT NULL DEFAULT NOW(),
    blocked_until TIMESTAMP,
    
    -- Indexes
    INDEX idx_failed_attempts_identifier (identifier),
    INDEX idx_failed_attempts_blocked_until (blocked_until)
);

-- Rate limiting table
CREATE TABLE rate_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identifier VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    requests_count INTEGER NOT NULL DEFAULT 1,
    window_start TIMESTAMP NOT NULL DEFAULT NOW(),
    window_end TIMESTAMP NOT NULL,
    
    -- Indexes
    UNIQUE INDEX idx_rate_limits_identifier_endpoint (identifier, endpoint),
    INDEX idx_rate_limits_window_end (window_end)
);

-- Create indexes for performance
CREATE INDEX idx_users_email_lower ON users (LOWER(email));
CREATE INDEX idx_users_status ON users (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users (role) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users (created_at);

-- Triggers for updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin user (password: 'admin123!')
-- In production, this should be done securely through a setup script
INSERT INTO users (
    id,
    email,
    password_hash,
    first_name,
    last_name,
    role,
    status
) VALUES (
    uuid_generate_v4(),
    'admin@lendingplatform.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK9Lf6xam', -- bcrypt hash of 'admin123!'
    'System',
    'Administrator',
    'admin',
    'active'
) ON CONFLICT (email) DO NOTHING;

-- Insert test users for development
INSERT INTO users (
    id,
    email,
    password_hash,
    first_name,
    last_name,
    role,
    status
) VALUES 
(
    uuid_generate_v4(),
    'applicant@test.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK9Lf6xam', -- bcrypt hash of 'admin123!'
    'John',
    'Doe',
    'applicant',
    'active'
),
(
    uuid_generate_v4(),
    'reviewer@test.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK9Lf6xam', -- bcrypt hash of 'admin123!'
    'Jane',
    'Smith',
    'junior_reviewer',
    'active'
),
(
    uuid_generate_v4(),
    'manager@test.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RK9Lf6xam', -- bcrypt hash of 'admin123!'
    'Mike',
    'Johnson',
    'manager',
    'active'
)
ON CONFLICT (email) DO NOTHING;

-- Create cleanup job for expired sessions (run daily)
-- This would typically be handled by a cron job or background worker
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    INSERT INTO security_events (
        event_type,
        severity,
        description,
        metadata
    ) VALUES (
        'session_cleanup',
        'low',
        'Automated cleanup of expired sessions',
        jsonb_build_object('deleted_sessions', deleted_count)
    );
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE users IS 'User account information for authentication';
COMMENT ON TABLE user_sessions IS 'Active user sessions with refresh tokens';
COMMENT ON TABLE auth_events IS 'Authentication events audit log';
COMMENT ON TABLE security_events IS 'Security-related events audit log';
COMMENT ON TABLE failed_login_attempts IS 'Failed login attempt tracking for rate limiting';
COMMENT ON TABLE rate_limits IS 'Rate limiting counters per identifier and endpoint';

COMMENT ON COLUMN users.password_hash IS 'bcrypt hashed password';
COMMENT ON COLUMN users.role IS 'User role for RBAC: applicant, junior_reviewer, senior_reviewer, manager, admin';
COMMENT ON COLUMN users.status IS 'Account status: active, inactive, locked, suspended';
COMMENT ON COLUMN user_sessions.refresh_token IS 'Cryptographically secure refresh token';
COMMENT ON COLUMN auth_events.success IS 'Whether the authentication event was successful';
COMMENT ON COLUMN security_events.severity IS 'Event severity: low, medium, high, critical';
