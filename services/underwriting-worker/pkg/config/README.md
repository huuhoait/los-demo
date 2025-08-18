# Underwriting Worker Configuration

This package provides configuration management for the underwriting-worker service, specifically designed to handle external services configuration that was moved from the shared package.

## Overview

The underwriting-worker configuration includes all the standard service configurations (server, database, logging, etc.) plus specialized configurations for external services used in underwriting operations:

- **Credit Bureau Services** (Experian, TransUnion, Equifax)
- **Risk Scoring Services** (FICO, VantageScore)
- **Income Verification Services** (Plaid, Yodlee)
- **Decision Engine Services** (Internal or external)
- **Notification Services** (SendGrid, Twilio, etc.)

## Configuration Structure

### Main Configuration
```go
type UnderwritingConfig struct {
    Server      ServerConfig    `yaml:"server"`
    Database    DatabaseConfig  `yaml:"database"`
    Conductor   ConductorConfig `yaml:"conductor"`
    Logging     LoggingConfig   `yaml:"logging"`
    Security    SecurityConfig  `yaml:"security"`
    Application AppConfig       `yaml:"application"`
    I18n        I18nConfig      `yaml:"i18n"`
    Services    ServicesConfig  `yaml:"services"`  // Underwriting-specific
}
```

### External Services Configuration
```go
type ServicesConfig struct {
    CreditBureau       CreditBureauConfig       `yaml:"credit_bureau"`
    RiskScoring        RiskScoringConfig        `yaml:"risk_scoring"`
    IncomeVerification IncomeVerificationConfig `yaml:"income_verification"`
    DecisionEngine     DecisionEngineConfig     `yaml:"decision_engine"`
    Notification       NotificationConfig       `yaml:"notification"`
}
```

## Usage

### 1. Load Configuration
```go
import "underwriting_worker/pkg/config"

cfg, err := config.LoadConfig("config/underwriting.yaml")
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

### 2. Access External Services Configuration
```go
// Credit Bureau
creditBureauURL := cfg.Services.CreditBureau.BaseURL
creditBureauAPIKey := cfg.Services.CreditBureau.APIKey
creditBureauTimeout := cfg.Services.CreditBureau.Timeout

// Risk Scoring
riskScoringURL := cfg.Services.RiskScoring.BaseURL
riskScoringModel := cfg.Services.RiskScoring.ModelVersion

// Income Verification
incomeVerificationURL := cfg.Services.IncomeVerification.BaseURL

// Decision Engine
decisionEngineURL := cfg.Services.DecisionEngine.BaseURL

// Notification
notificationURL := cfg.Services.Notification.BaseURL
```

### 3. Environment Variable Overrides
The configuration automatically supports environment variable overrides:

```bash
# Credit Bureau
export CREDIT_BUREAU_API_KEY="your-api-key"
export CREDIT_BUREAU_BASE_URL="https://api.experian.com"

# Risk Scoring
export RISK_SCORING_API_KEY="your-api-key"
export RISK_SCORING_BASE_URL="https://api.fico.com"

# Income Verification
export INCOME_VERIFICATION_API_KEY="your-api-key"
export INCOME_VERIFICATION_BASE_URL="https://api.plaid.com"

# Decision Engine
export DECISION_ENGINE_API_KEY="your-api-key"
export DECISION_ENGINE_BASE_URL="http://localhost:8083"

# Notification
export NOTIFICATION_API_KEY="your-api-key"
export NOTIFICATION_BASE_URL="https://api.sendgrid.com"
```

## Configuration File Example

See `config/underwriting.yaml` for a complete example with:
- Default configuration
- Development environment
- Production environment
- Environment variable placeholders

## Default Values

The configuration provides sensible defaults for all services:

- **Credit Bureau**: Experian provider, 30s timeout
- **Risk Scoring**: FICO provider, 30s timeout
- **Income Verification**: Plaid provider, 30s timeout
- **Decision Engine**: Internal provider, 30s timeout
- **Notification**: SendGrid provider, 30s timeout

## Validation

The configuration includes validation to ensure:
- Server port is valid
- Database port is valid
- All required fields are present

## Migration from Shared Package

This configuration was moved from the shared package to keep underwriting-specific configurations separate and focused. The shared package now contains only common configurations that are truly shared across all services.

## Benefits

1. **Separation of Concerns**: Underwriting-specific configs are now in the underwriting service
2. **Focused Configuration**: Each service has only the configs it needs
3. **Easier Maintenance**: Changes to underwriting configs don't affect other services
4. **Better Security**: Sensitive API keys are isolated to the underwriting service
5. **Flexibility**: Each service can evolve its configuration independently
