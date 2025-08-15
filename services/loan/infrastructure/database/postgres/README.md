# PostgreSQL Database Repositories

This directory contains the PostgreSQL implementation of the database repositories for the loan service.

## Overview

The database layer provides:
- **Connection Management**: PostgreSQL connection with connection pooling
- **User Repository**: CRUD operations for user data
- **Loan Repository**: CRUD operations for loan applications, offers, and workflow executions
- **Factory Pattern**: Easy repository instantiation and management

## Structure

```
infrastructure/database/postgres/
├── connection.go          # Database connection and configuration
├── user_repository.go     # User CRUD operations
├── loan_repository.go     # Loan application CRUD operations
├── factory.go            # Repository factory
├── migrations/           # Database schema migrations
│   └── 001_create_users_table.sql
└── README.md             # This file
```

## Database Schema

### Users Table
- **id**: UUID primary key
- **first_name, last_name**: User's full name
- **email**: Unique email address
- **phone_number**: Contact phone
- **date_of_birth**: Birth date
- **ssn**: Social Security Number
- **Address fields**: Street, city, state, zip, country, residence type, time at address
- **Employment fields**: Employer, job title, time employed, work contact info
- **Banking fields**: Bank name, account type, account/routing numbers
- **Timestamps**: created_at, updated_at

### Loan Applications Table
- **id**: UUID primary key
- **user_id**: Foreign key to users table
- **application_number**: Unique application identifier
- **loan_amount**: Requested loan amount
- **loan_purpose**: Purpose of the loan
- **requested_term_months**: Loan term in months
- **income fields**: Annual and monthly income
- **employment_status**: Current employment status
- **monthly_debt_payments**: Current debt obligations
- **current_state**: Application workflow state
- **status**: Application status
- **risk_score**: Calculated risk score
- **workflow_id**: Netflix Conductor workflow ID
- **Timestamps**: created_at, updated_at

### Loan Offers Table
- **id**: UUID primary key
- **application_id**: Foreign key to loan_applications
- **offer_amount**: Offered loan amount
- **interest_rate**: Annual interest rate
- **term_months**: Loan term in months
- **monthly_payment**: Calculated monthly payment
- **total_interest**: Total interest over loan term
- **apr**: Annual Percentage Rate
- **expires_at**: Offer expiration date
- **status**: Offer status
- **created_at**: Creation timestamp

### State Transitions Table
- **id**: UUID primary key
- **application_id**: Foreign key to loan_applications
- **from_state**: Previous application state
- **to_state**: New application state
- **transition_reason**: Reason for state change
- **triggered_by**: Who/what triggered the transition
- **created_at**: Transition timestamp

### Workflow Executions Table
- **id**: UUID primary key
- **workflow_id**: Netflix Conductor workflow ID
- **application_id**: Foreign key to loan_applications
- **status**: Workflow execution status
- **start_time**: Workflow start timestamp
- **end_time**: Workflow completion timestamp
- **created_at**: Record creation timestamp

## Usage

### 1. Initialize Database Connection

```go
import "loan-service/infrastructure/database/postgres"

// Create database configuration
config := &postgres.Config{
    Host:     "localhost",
    Port:     "5432",
    User:     "postgres",
    Password: "password",
    Database: "loan_service",
    SSLMode:  "disable",
}

// Create connection
connection, err := postgres.NewConnection(config, logger)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}
defer connection.Close()
```

### 2. Use Repository Factory

```go
// Create factory
factory := postgres.NewFactory(connection, logger)

// Get repositories
userRepo := factory.GetUserRepository()
loanRepo := factory.GetLoanRepository()

// Use repositories
user, err := userRepo.GetUserByID(ctx, "user-123")
if err != nil {
    log.Printf("Failed to get user: %v", err)
}
```

### 3. Direct Repository Usage

```go
// Create repositories directly
userRepo := postgres.NewUserRepository(connection, logger)
loanRepo := postgres.NewLoanRepository(connection, logger)

// Use repositories
app, err := loanRepo.GetApplicationByID(ctx, "app-123")
if err != nil {
    log.Printf("Failed to get application: %v", err)
}
```

## Features

### Connection Pooling
- Configurable connection pool size
- Automatic connection health checks
- Graceful connection cleanup

### Transaction Support
```go
tx, err := connection.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

// Use transaction for multiple operations
// ...

err = tx.Commit()
```

### Error Handling
- Structured error logging with context
- Database-specific error types
- Graceful fallbacks for common scenarios

### Performance
- Prepared statement support
- Connection pooling
- Optimized queries with proper indexing

## Migration

### Running Migrations

1. **Manual SQL Execution**:
   ```bash
   psql -h localhost -U postgres -d loan_service -f infrastructure/database/postgres/migrations/001_create_users_table.sql
   ```

2. **Docker Compose**:
   ```bash
   docker-compose exec postgres psql -U postgres -d loan_service -f /app/migrations/001_create_users_table.sql
   ```

### Creating New Migrations

1. Create a new SQL file in the `migrations/` directory
2. Use sequential numbering (002_, 003_, etc.)
3. Include both `CREATE` and `DROP` statements
4. Test migrations in development first

## Configuration

### Environment Variables

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=loan_service
DB_SSLMODE=disable
```

### Connection Pool Settings

```go
config := &postgres.Config{
    MaxOpenConns:    25,        // Maximum open connections
    MaxIdleConns:    5,         // Maximum idle connections
    ConnMaxLifetime: 5 * time.Minute, // Connection lifetime
}
```

## Testing

### Unit Tests
```bash
cd infrastructure/database/postgres
go test -v ./...
```

### Integration Tests
```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Run tests
go test -v -tags=integration ./...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

## Security Considerations

1. **Connection Security**: Use SSL in production
2. **Credential Management**: Use environment variables or secret management
3. **SQL Injection**: All queries use parameterized statements
4. **Access Control**: Implement proper database user permissions
5. **Audit Logging**: All operations are logged with context

## Monitoring

### Health Checks
```go
err := connection.HealthCheck(ctx)
if err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

### Metrics
- Connection pool utilization
- Query execution times
- Error rates
- Transaction success rates

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check database host/port
2. **Authentication Failed**: Verify username/password
3. **Database Not Found**: Ensure database exists
4. **Permission Denied**: Check user permissions
5. **Connection Timeout**: Increase timeout settings

### Debug Mode
```go
// Enable debug logging
logger.SetLevel(zap.DebugLevel)

// Check connection details
log.Printf("Database: %s:%s/%s", config.Host, config.Port, config.Database)
```

## Future Enhancements

1. **Read Replicas**: Support for read-only replicas
2. **Sharding**: Horizontal database scaling
3. **Caching**: Redis integration for frequently accessed data
4. **Backup/Restore**: Automated backup procedures
5. **Schema Evolution**: Automated migration tools
