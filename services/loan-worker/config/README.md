# Configuration System

This directory contains configuration files for the Loan Service application.

## Configuration Files

### `config.yaml`
Main configuration file containing all environment configurations:
- `default` - Default configuration used when no environment is specified
- `development` - Development environment configuration
- `docker` - Docker environment configuration
- `production` - Production environment configuration
- `test` - Test environment configuration

### `development.yaml`
Development-specific configuration overrides (optional).

## Environment Variables

The configuration system supports environment variable overrides. The following environment variables can be used:

### Server Configuration
- `PORT` - Server port (default: 8080)
- `HOST` - Server host (default: 0.0.0.0)
- `READ_TIMEOUT` - Read timeout in seconds (default: 30)
- `WRITE_TIMEOUT` - Write timeout in seconds (default: 30)

### Database Configuration
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: password)
- `DB_NAME` - Database name (default: loan_service)
- `DB_SSLMODE` - Database SSL mode (default: disable)

### Conductor Configuration
- `CONDUCTOR_BASE_URL` - Netflix Conductor base URL
- `CONDUCTOR_TIMEOUT` - Conductor timeout in seconds

### Logging Configuration
- `LOG_LEVEL` - Log level (debug, info, warn, error)
- `LOG_FORMAT` - Log format (json, console)

### Security Configuration
- `JWT_SECRET` - JWT secret key
- `JWT_EXPIRATION` - JWT expiration time in seconds

### Application Configuration
- `APP_ENV` - Application environment (development, docker, production, test)
- `MAX_LOAN_AMOUNT` - Maximum loan amount
- `MIN_LOAN_AMOUNT` - Minimum loan amount

## Usage

### Setting Environment
Set the `APP_ENV` environment variable to specify which configuration to use:

```bash
# Development
export APP_ENV=development

# Docker
export APP_ENV=docker

# Production
export APP_ENV=production

# Test
export APP_ENV=test
```

### Running with Different Configurations

#### Development
```bash
export APP_ENV=development
go run cmd/main.go
```

#### Docker
```bash
export APP_ENV=docker
docker-compose up
```

#### Production
```bash
export APP_ENV=production
export DB_HOST=your-db-host
export DB_PASSWORD=your-secure-password
export JWT_SECRET=your-secure-jwt-secret
./loan-service
```

## Configuration Structure

The configuration is structured as follows:

```yaml
environment_name:
  server:
    port: 8080
    host: "0.0.0.0"
    read_timeout: 30
    write_timeout: 30
    graceful_shutdown_timeout: 30
  
  database:
    host: "localhost"
    port: 5432
    user: "postgres"
    password: "password"
    name: "loan_service"
    ssl_mode: "disable"
    max_open_conns: 25
    max_idle_conns: 5
    conn_max_lifetime: 300
  
  conductor:
    base_url: "http://localhost:8080"
    timeout: 30
    retry_attempts: 3
    retry_delay: 1000
  
  logging:
    level: "info"
    format: "json"
    output: "stdout"
    enable_console: true
    enable_file: false
    file_path: "./logs/loan-service.log"
    max_size: 100
    max_age: 30
    max_backups: 10
  
  security:
    jwt_secret: "your-secret-key"
    jwt_expiration: 3600
    bcrypt_cost: 12
    cors_allowed_origins: ["*"]
    cors_allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    cors_allowed_headers: ["Content-Type", "Authorization", "X-Request-ID", "X-Language"]
  
  application:
    name: "loan-service"
    version: "v1.0.0"
    environment: "development"
    max_loan_amount: 50000
    min_loan_amount: 5000
    max_dti_ratio: 0.43
    default_interest_rate: 8.5
    max_interest_rate: 15.0
    min_interest_rate: 5.0
    offer_expiration_hours: 168
  
  i18n:
    default_language: "en"
    supported_languages: ["en", "vi"]
    fallback_language: "en"
```

## Security Considerations

1. **Never commit sensitive information** like passwords, API keys, or JWT secrets to version control
2. **Use environment variables** for sensitive configuration in production
3. **Use different secrets** for different environments
4. **Rotate secrets regularly** in production environments
5. **Use secure SSL/TLS** connections for database and external services in production

## Best Practices

1. **Environment-specific configs**: Use different configurations for different environments
2. **Default values**: Always provide sensible defaults
3. **Validation**: Validate configuration on startup
4. **Documentation**: Document all configuration options
5. **Testing**: Test configuration loading in different environments
6. **Backup**: Keep backup copies of production configurations
