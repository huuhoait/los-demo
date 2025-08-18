# Configuration Management for Underwriting Worker

This directory contains environment-specific configuration files for the Underwriting Worker service.

## Configuration Structure

The configuration is split into multiple files for better organization and environment-specific customization:

- **`base.yaml`** - Common configuration shared across all environments
- **`development.yaml`** - Development environment specific settings
- **`uat.yaml`** - User Acceptance Testing environment specific settings  
- **`production.yaml`** - Production environment specific settings

## Environment Selection

The service automatically determines which configuration to load based on the `ENVIRONMENT` environment variable:

```bash
# Development
export ENVIRONMENT=development

# UAT
export ENVIRONMENT=uat

# Production
export ENVIRONMENT=production
```

If no environment is specified, it defaults to `development`.

## Configuration Loading

The configuration loader follows this hierarchy:

1. **Base Configuration** - Loads common settings from `base.yaml`
2. **Environment Configuration** - Overrides base settings with environment-specific values from `{environment}.yaml`
3. **Environment Variables** - Substitutes placeholder values like `${API_KEY}` with actual environment variables

## Environment-Specific Features

### Development Environment
- **Mock Services**: Enabled for all external services
- **Debug Mode**: Enabled with detailed logging
- **Local Resources**: Uses localhost for databases and services
- **Console Logging**: Human-readable console output
- **Relaxed Security**: Development JWT secrets and CORS

### UAT Environment
- **Real Services**: All external services are real (no mocks)
- **Staging URLs**: Uses UAT-specific service endpoints
- **File Logging**: JSON logs written to files
- **Enhanced Security**: Proper JWT secrets and restricted CORS
- **Performance Monitoring**: Basic metrics and health checks

### Production Environment
- **Production Services**: All services use production endpoints
- **High Performance**: Optimized connection pools and worker counts
- **Advanced Security**: Strict CORS, rate limiting, and encryption
- **Comprehensive Monitoring**: Full observability stack
- **Backup & Recovery**: Automated backup systems
- **Circuit Breakers**: Fault tolerance and resilience patterns

## Key Configuration Sections

### Service Configuration
```yaml
service:
  name: "underwriting-worker"
  version: "1.0.0"
  port: "8083"
```

### Server Configuration
```yaml
server:
  port: 8083
  host: "0.0.0.0"
  read_timeout: 30
  write_timeout: 30
  graceful_shutdown_timeout: 30
  cors:
    enabled: true
    origins: ["https://company.com"]
    methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
```

### Database Configuration
```yaml
database:
  host: "localhost"
  port: 5433
  user: "postgres"
  password: "${DB_PASSWORD}"
  name: "underwriting_db_dev"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
```

### Conductor Configuration
```yaml
conductor:
  base_url: "http://localhost:8082"
  timeout: 30
  retry_attempts: 3
  worker_pool_size: 10
  polling_interval_ms: 1000
```

### External Services
```yaml
services:
  credit_bureau:
    provider: "experian"
    base_url: "https://api.experian.com"
    api_key: "${CREDIT_BUREAU_API_KEY}"
    timeout_seconds: 30
    mock_enabled: true  # Only in development
```

### Logging Configuration
```yaml
logging:
  level: "debug"  # debug, info, warn, error
  format: "console"  # console, json
  output: "stdout"  # stdout, file
  enable_console: true
  enable_file: false
  file_path: "./logs/underwriting-worker-dev.log"
```

### Security Configuration
```yaml
security:
  jwt_secret: "${JWT_SECRET}"
  jwt_expiration: 3600
  bcrypt_cost: 10
  cors_allowed_origins: ["http://localhost:3000"]
  rate_limiting_enabled: false  # Only in production
```

## Environment Variables

The following environment variables are supported for configuration:

### Database
- `DB_PASSWORD` - Database password
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_NAME` - Database name

### External Services
- `CREDIT_BUREAU_API_KEY` - Experian API key
- `RISK_SCORING_API_KEY` - Risk scoring service API key
- `INCOME_VERIFICATION_API_KEY` - Plaid API key
- `DECISION_ENGINE_API_KEY` - Decision engine API key
- `NOTIFICATION_API_KEY` - SendGrid API key

### Security
- `JWT_SECRET` - JWT signing secret
- `REDIS_PASSWORD` - Redis password

### Environment
- `ENVIRONMENT` - Current environment (development, uat, production)
- `APP_ENV` - Alternative environment variable name

## Usage Examples

### Running in Development
```bash
export ENVIRONMENT=development
export CREDIT_BUREAU_API_KEY=your_dev_key
export JWT_SECRET=dev-secret-key

go run cmd/main.go
```

### Running in UAT
```bash
export ENVIRONMENT=uat
export UAT_DB_PASSWORD=uat_password
export UAT_CREDIT_BUREAU_API_KEY=uat_key
export UAT_JWT_SECRET=uat-secret-key

go run cmd/main.go
```

### Running in Production
```bash
export ENVIRONMENT=production
export PROD_DB_PASSWORD=prod_password
export PROD_CREDIT_BUREAU_API_KEY=prod_key
export PROD_JWT_SECRET=prod-secret-key

go run cmd/main.go
```

## Docker Environment

For Docker deployments, you can override the environment:

```bash
docker run -e ENVIRONMENT=production \
  -e PROD_DB_PASSWORD=secret \
  -e PROD_CREDIT_BUREAU_API_KEY=key \
  underwriting-worker:latest
```

## Configuration Validation

The configuration loader validates required fields:

- Service name and port
- Database host
- Conductor base URL
- Required API keys

## Best Practices

1. **Never commit sensitive data** - Use environment variables for secrets
2. **Environment-specific files** - Keep configurations separate and focused
3. **Base configuration** - Share common settings across environments
4. **Validation** - Always validate configuration on startup
5. **Documentation** - Keep this README updated with new configuration options

## Troubleshooting

### Common Issues

1. **Configuration not found**: Ensure the environment variable is set correctly
2. **Missing environment file**: Check that `{environment}.yaml` exists
3. **Environment variable substitution**: Verify placeholder format `${VARIABLE_NAME}`
4. **YAML syntax errors**: Validate YAML syntax with a YAML validator

### Debug Configuration

Enable debug logging to see configuration loading details:

```bash
export LOG_LEVEL=debug
go run cmd/main.go
```

## Adding New Configuration

To add new configuration options:

1. **Update the Config struct** in `pkg/config/config.go`
2. **Add to base.yaml** if it's common across environments
3. **Add to environment-specific files** if it varies by environment
4. **Update this README** with documentation
5. **Add validation** if the field is required
