-- Decision Engine Database Schema

-- Create decision_requests table
CREATE TABLE IF NOT EXISTS decision_requests (
    id SERIAL PRIMARY KEY,
    application_id VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    customer_id VARCHAR(255),
    loan_amount DECIMAL(15,2) NOT NULL,
    annual_income DECIMAL(15,2) NOT NULL,
    monthly_income DECIMAL(15,2) NOT NULL,
    monthly_debt DECIMAL(15,2) DEFAULT 0,
    credit_score INTEGER NOT NULL,
    employment_type VARCHAR(50) NOT NULL,
    requested_term INTEGER NOT NULL,
    loan_term_months INTEGER,
    loan_purpose VARCHAR(100) NOT NULL,
    additional_data JSONB,
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create decisions table
CREATE TABLE IF NOT EXISTS decisions (
    id SERIAL PRIMARY KEY,
    application_id VARCHAR(255) UNIQUE NOT NULL,
    decision VARCHAR(50) NOT NULL,
    confidence_score DECIMAL(5,2),
    interest_rate DECIMAL(5,2),
    max_amount DECIMAL(15,2),
    reason TEXT,
    risk_assessment JSONB,
    applied_rules JSONB,
    recommendations JSONB,
    decision_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_decision_requests_application_id ON decision_requests(application_id);
CREATE INDEX IF NOT EXISTS idx_decision_requests_user_id ON decision_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_decision_requests_customer_id ON decision_requests(customer_id);
CREATE INDEX IF NOT EXISTS idx_decision_requests_requested_at ON decision_requests(requested_at);

CREATE INDEX IF NOT EXISTS idx_decisions_application_id ON decisions(application_id);
CREATE INDEX IF NOT EXISTS idx_decisions_decision ON decisions(decision);
CREATE INDEX IF NOT EXISTS idx_decisions_decision_date ON decisions(decision_date);
CREATE INDEX IF NOT EXISTS idx_decisions_created_at ON decisions(created_at);

-- Add foreign key constraint
ALTER TABLE decisions 
ADD CONSTRAINT fk_decisions_application_id 
FOREIGN KEY (application_id) REFERENCES decision_requests(application_id);
