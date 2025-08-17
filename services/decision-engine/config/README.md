# Decision Engine Configuration

This directory contains configuration files for the Decision Engine service.

## Configuration Files

### config.yaml
Main configuration file containing default settings for:
- Server configuration (port, timeouts, CORS)
- Database settings
- Redis configuration
- Logging preferences
- Decision engine specific settings (risk assessment, interest rates, loan limits)
- External service endpoints
- Business rules

### development.yaml
Development environment overrides that provide:
- Debug logging
- Development database settings
- Mock external services
- Lenient risk assessment criteria
- Test endpoints enablement

## Environment-Specific Loading

The application loads configuration in the following order:
1. Base configuration from `config.yaml`
2. Environment-specific overrides from `{environment}.yaml`
3. Environment variables (highest priority)

## Configuration Structure

```yaml
server:
  port: "8082"
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"

database:
  url: "postgres://..."
  max_open_conns: 25
  max_idle_conns: 10

decision_engine:
  risk_assessment:
    min_credit_score: 600
    max_dti_ratio: 0.45
  interest_rates:
    base_rate: 6.5
  loan_limits:
    min_amount: 1000
    max_amount: 100000

business_rules:
  auto_approval:
    min_credit_score: 750
  auto_denial:
    min_credit_score: 550
```

## Environment Variables

Key environment variables that override configuration:

- `ENVIRONMENT`: Set to "development", "staging", or "production"
- `PORT`: Server port (default: 8082)
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_HOST`, `REDIS_PORT`: Redis connection details
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## Adding New Configuration

1. Add the setting to `config.yaml` with default values
2. Add environment-specific overrides if needed
3. Update the configuration struct in `pkg/config/config.go`
4. Use the setting in your service code via the config object
