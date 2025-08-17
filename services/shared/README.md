# Shared Library for Microservices

This shared library provides common functionality for all microservices in the LOS (Loan Origination System) project.

## Packages

### 1. Config (`pkg/config`)

Provides configuration management with environment variable support and YAML configuration files.

```go
import "github.com/huuhoait/los-demo/services/shared/pkg/config"

// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Access configuration
fmt.Println(cfg.Service.Name)
fmt.Println(cfg.Database.Host)
```

### 2. Logger (`pkg/logger`)

Provides structured logging with Zap logger, request logging middleware, and business event logging.

```go
import "github.com/huuhoait/los-demo/services/shared/pkg/logger"

// Create logger
logger, err := logger.New(logger.Config{
    Level:       "info",
    Format:      "json",
    Environment: "production",
})

// Use logger middleware
r.Use(logger.NewLoggerMiddleware(logger).Handler())

// Log business events
logger.LogUserAction("user123", "login", "auth", map[string]interface{}{
    "ip": "192.168.1.1",
})
```

### 3. I18n (`pkg/i18n`)

Provides internationalization support with Go-i18n, middleware for language detection, and helper functions.

```go
import "github.com/huuhoait/los-demo/services/shared/pkg/i18n"

// Create i18n manager
manager, err := i18n.NewManager(i18n.Config{
    DefaultLanguage: "en",
    Languages:       []string{"en", "vi"},
    BundlePath:      "./i18n",
})

// Use i18n middleware
r.Use(i18n.NewMiddleware(manager).Handler())

// Translate messages in handlers
message := i18n.T(c, "welcome.message", map[string]interface{}{
    "Name": "John",
})
```

### 4. Database (`pkg/database`)

Provides database connectivity with GORM, base repository pattern, and query builders.

```go
import "github.com/huuhoait/los-demo/services/shared/pkg/database"

// Connect to database
db, err := database.NewConnection(database.Config{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    Database: "los_db",
})

// Use base repository
type UserRepository struct {
    *database.Repository
}

func NewUserRepository(db *database.Connection) *UserRepository {
    return &UserRepository{
        Repository: database.NewRepository(db),
    }
}
```

### 5. Cache (`pkg/cache`)

Provides Redis caching with session storage, rate limiting, and distributed locking.

```go
import "github.com/huuhoait/los-demo/services/shared/pkg/cache"

// Connect to Redis
client, err := cache.NewClient(cache.Config{
    Host:     "localhost",
    Port:     6379,
    Database: 0,
})

// Use cache repository
repo := cache.NewRepository(client, "users")
err = repo.Set(ctx, "user:123", userData, time.Hour)

// Use rate limiter
limiter := cache.NewRateLimiter(client)
allowed, err := limiter.Allow(ctx, "user:123", 100, time.Minute)
```

### 6. Middleware (`pkg/middleware`)

Provides common HTTP middleware for Gin framework.

```go
import "github.com/huuhoait/los-demo/services/shared/pkg/middleware"

// Use common middleware
r.Use(middleware.RequestIDMiddleware())
r.Use(middleware.CORSMiddleware())
r.Use(middleware.SecurityHeadersMiddleware())
r.Use(middleware.RecoveryMiddleware())

// Use pagination middleware
r.GET("/users", middleware.PaginationMiddleware(), func(c *gin.Context) {
    pagination := middleware.GetPagination(c)
    // Use pagination.Page, pagination.PerPage, pagination.Offset
})
```

## Usage in Services

### 1. Add Dependency

Add the shared library as a dependency in your service's `go.mod`:

```go
require (
    github.com/huuhoait/los-demo/services/shared v0.0.0
)
```

### 2. Replace Local Packages

Replace local `pkg/` directories with imports from the shared library:

```go
// Before
import "github.com/your-org/your-service/pkg/config"

// After
import "github.com/huuhoait/los-demo/services/shared/pkg/config"
```

### 3. Update Main Application

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/huuhoait/los-demo/services/shared/pkg/config"
    "github.com/huuhoait/los-demo/services/shared/pkg/database"
    "github.com/huuhoait/los-demo/services/shared/pkg/logger"
    "github.com/huuhoait/los-demo/services/shared/pkg/middleware"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        panic(err)
    }

    // Initialize logger
    log, err := logger.New(cfg.Logging)
    if err != nil {
        panic(err)
    }

    // Connect to database
    db, err := database.NewConnection(cfg.Database)
    if err != nil {
        panic(err)
    }

    // Setup Gin router
    r := gin.New()
    
    // Add shared middleware
    r.Use(middleware.RequestIDMiddleware())
    r.Use(middleware.CORSMiddleware())
    r.Use(logger.NewLoggerMiddleware(log).Handler())
    
    // Your routes here
    setupRoutes(r, db, log)
    
    // Start server
    r.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}
```

## Configuration Structure

All services should follow this configuration structure:

```yaml
service:
  name: "auth-service"
  version: "1.0.0"
  environment: "development"

server:
  port: 8080
  read_timeout: 30
  write_timeout: 30

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  database: "auth_db"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 60

redis:
  host: "localhost"
  port: 6379
  password: ""
  database: 0
  pool_size: 10

logging:
  level: "info"
  format: "json"
  output: "stdout"
  environment: "development"

i18n:
  default_language: "en"
  languages: ["en", "vi"]
  bundle_path: "./i18n"
```

## Migration Guide

### Step 1: Update go.mod

```bash
go mod edit -require=github.com/huuhoait/los-demo/services/shared@latest
```

### Step 2: Remove Local Packages

Remove the following directories from each service:
- `pkg/config/`
- `pkg/logger/`
- `pkg/i18n/`
- Common middleware files

### Step 3: Update Imports

Find and replace imports across your codebase:

```bash
# Update config imports
find . -name "*.go" -exec sed -i 's|".*pkg/config"|"github.com/huuhoait/los-demo/services/shared/pkg/config"|g' {} \;

# Update logger imports
find . -name "*.go" -exec sed -i 's|".*pkg/logger"|"github.com/huuhoait/los-demo/services/shared/pkg/logger"|g' {} \;

# Update i18n imports
find . -name "*.go" -exec sed -i 's|".*pkg/i18n"|"github.com/huuhoait/los-demo/services/shared/pkg/i18n"|g' {} \;
```

### Step 4: Update Configuration

Ensure your service configuration follows the standard structure defined above.

### Step 5: Test and Validate

```bash
go mod tidy
go build
go test ./...
```

## Benefits

1. **Code Reuse**: Eliminates duplication across services
2. **Consistency**: Ensures all services use the same patterns
3. **Maintenance**: Centralized updates and bug fixes
4. **Standards**: Enforces coding standards and best practices
5. **Efficiency**: Faster development of new services

## Development

To contribute to the shared library:

1. Make changes to packages in `pkg/`
2. Update version in `go.mod`
3. Tag a new release
4. Update services to use the new version

```bash
# Tag a new version
git tag v1.0.1
git push origin v1.0.1

# Update services
go get github.com/huuhoait/los-demo/services/shared@v1.0.1
```
