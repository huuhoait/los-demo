#!/bin/bash

# User Service Deployment Script
# This script sets up and deploys the User Service microservice

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="user-service"
SERVICE_PORT=8082
DB_NAME="user_service_db"
DB_USER="user_service_user"
DB_PASSWORD="secure_password_123"
REDIS_DB=1

echo -e "${BLUE}🚀 Starting User Service Deployment${NC}"

# Function to print status
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if running in the correct directory
if [ ! -f "cmd/main.go" ]; then
    print_error "This script must be run from the user service root directory"
    exit 1
fi

# Check dependencies
echo -e "${BLUE}📋 Checking dependencies...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi
print_status "Go is installed"

# Check if PostgreSQL is available
if ! command -v psql &> /dev/null; then
    print_warning "PostgreSQL client not found. Make sure PostgreSQL is accessible."
fi

# Check if Redis is available
if ! command -v redis-cli &> /dev/null; then
    print_warning "Redis client not found. Make sure Redis is accessible."
fi

# Check if Docker is available (optional)
if command -v docker &> /dev/null; then
    print_status "Docker is available"
    DOCKER_AVAILABLE=true
else
    print_warning "Docker not found. Skipping containerized deployment options."
    DOCKER_AVAILABLE=false
fi

# Create necessary directories
echo -e "${BLUE}📁 Creating directories...${NC}"
mkdir -p logs configs tmp uploads

# Create configuration files
echo -e "${BLUE}⚙️  Creating configuration files...${NC}"

# Development configuration
cat > configs/development.yaml << EOF
environment: development
service:
  name: $SERVICE_NAME
  port: $SERVICE_PORT
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: $SERVICE_PORT
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

database:
  host: "localhost"
  port: 5432
  name: "$DB_NAME"
  user: "$DB_USER"
  password: "$DB_PASSWORD"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: $REDIS_DB
  pool_size: 10
  min_idle_conns: 3
  max_retries: 3

storage:
  provider: "s3"
  bucket: "user-service-documents-dev"
  region: "us-west-2"
  access_key: "your-access-key"
  secret_key: "your-secret-key"
  endpoint: ""

encryption:
  master_key: "your-32-character-master-key-here"
  key_rotation_days: 90

logging:
  level: "debug"
  format: "json"
  output: "stdout"
  file_path: "logs/user-service.log"
  max_size: 100
  max_backups: 5
  max_age: 30

security:
  jwt_secret: "your-jwt-secret-key-here"
  bcrypt_cost: 12
  rate_limit_requests: 100
  rate_limit_window: 60

external_services:
  kyc_provider:
    url: "https://api.kyc-provider.com"
    api_key: "your-kyc-api-key"
    timeout: 30
  
  notification_service:
    url: "http://localhost:8084"
    timeout: 10
  
  audit_service:
    url: "http://localhost:8085"
    timeout: 5

features:
  enable_2fa: true
  enable_document_ocr: false
  enable_real_kyc: false
  max_file_size: 52428800
  allowed_file_types: ["image/jpeg", "image/png", "application/pdf"]
EOF

# Production configuration template
cat > configs/production.yaml.template << EOF
# Production Configuration Template
# Copy this to production.yaml and update with actual values

environment: production
service:
  name: $SERVICE_NAME
  port: $SERVICE_PORT
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: $SERVICE_PORT
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

database:
  host: "your-production-db-host"
  port: 5432
  name: "$DB_NAME"
  user: "$DB_USER"
  password: "secure-production-password"
  ssl_mode: "require"
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 600

redis:
  host: "your-production-redis-host"
  port: 6379
  password: "secure-redis-password"
  db: $REDIS_DB
  pool_size: 20
  min_idle_conns: 5
  max_retries: 3

storage:
  provider: "s3"
  bucket: "user-service-documents-prod"
  region: "us-west-2"
  access_key: "production-access-key"
  secret_key: "production-secret-key"

encryption:
  master_key: "secure-32-character-production-key"
  key_rotation_days: 30

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/user-service/app.log"

security:
  jwt_secret: "production-jwt-secret"
  bcrypt_cost: 14
  rate_limit_requests: 50
  rate_limit_window: 60

# ... rest of configuration
EOF

print_status "Configuration files created"

# Initialize Go modules
echo -e "${BLUE}📦 Initializing Go modules...${NC}"
if [ ! -f "go.mod" ]; then
    go mod init our-los/services/user
fi

# Add necessary dependencies
go mod tidy
print_status "Go modules initialized"

# Database setup
echo -e "${BLUE}🗄️  Setting up database...${NC}"

# Check if we can connect to PostgreSQL
if command -v psql &> /dev/null; then
    echo "Attempting to connect to PostgreSQL..."
    
    # Try to connect as postgres user (adjust as needed)
    if psql -h localhost -U postgres -c "\q" 2>/dev/null; then
        print_status "Connected to PostgreSQL"
        
        # Create database and user
        echo "Creating database and user..."
        psql -h localhost -U postgres << EOF
CREATE DATABASE $DB_NAME;
CREATE USER $DB_USER WITH ENCRYPTED PASSWORD '$DB_PASSWORD';
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
\q
EOF
        
        # Run migrations
        echo "Running database migrations..."
        psql -h localhost -U $DB_USER -d $DB_NAME -f migrations/001_create_user_tables.sql
        print_status "Database setup complete"
    else
        print_warning "Could not connect to PostgreSQL. Please ensure it's running and create the database manually:"
        echo "  createdb $DB_NAME"
        echo "  psql -d $DB_NAME -f migrations/001_create_user_tables.sql"
    fi
else
    print_warning "PostgreSQL client not available. Please run migrations manually."
fi

# Build the service
echo -e "${BLUE}🔨 Building the service...${NC}"
go build -o bin/$SERVICE_NAME ./cmd/main.go
print_status "Service built successfully"

# Create systemd service file (for Linux)
if [ "$(uname)" = "Linux" ]; then
    echo -e "${BLUE}📋 Creating systemd service file...${NC}"
    
    cat > $SERVICE_NAME.service << EOF
[Unit]
Description=User Service for LOS System
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=los
Group=los
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/bin/$SERVICE_NAME
Restart=always
RestartSec=5
Environment=CONFIG_PATH=./configs
Environment=ENVIRONMENT=production

[Install]
WantedBy=multi-user.target
EOF
    
    print_status "Systemd service file created: $SERVICE_NAME.service"
    echo "To install: sudo cp $SERVICE_NAME.service /etc/systemd/system/"
    echo "To enable: sudo systemctl enable $SERVICE_NAME"
    echo "To start: sudo systemctl start $SERVICE_NAME"
fi

# Create Docker files if Docker is available
if [ "$DOCKER_AVAILABLE" = true ]; then
    echo -e "${BLUE}🐳 Creating Docker files...${NC}"
    
    # Dockerfile
    cat > Dockerfile << EOF
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

EXPOSE $SERVICE_PORT
CMD ["./main"]
EOF

    # Docker Compose
    cat > docker-compose.yml << EOF
version: '3.8'

services:
  $SERVICE_NAME:
    build: .
    ports:
      - "$SERVICE_PORT:$SERVICE_PORT"
    environment:
      - ENVIRONMENT=development
      - CONFIG_PATH=./configs
    depends_on:
      - postgres
      - redis
    volumes:
      - ./logs:/root/logs
      - ./uploads:/root/uploads

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: $DB_NAME
      POSTGRES_USER: $DB_USER
      POSTGRES_PASSWORD: $DB_PASSWORD
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
EOF
    
    print_status "Docker files created"
    echo "To run with Docker: docker-compose up --build"
fi

# Create development scripts
echo -e "${BLUE}📝 Creating development scripts...${NC}"

# Development runner script
cat > scripts/dev.sh << 'EOF'
#!/bin/bash
export ENVIRONMENT=development
export CONFIG_PATH=./configs
go run ./cmd/main.go
EOF

# Test runner script
cat > scripts/test.sh << 'EOF'
#!/bin/bash
echo "Running tests..."
go test -v ./...

echo "Running tests with coverage..."
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo "Coverage report generated: coverage.html"
EOF

# Migration runner script
cat > scripts/migrate.sh << 'EOF'
#!/bin/bash
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-user_service_user}
DB_NAME=${DB_NAME:-user_service_db}

echo "Running migrations..."
for migration in migrations/*.sql; do
    echo "Applying $(basename $migration)..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration"
done
echo "Migrations complete"
EOF

chmod +x scripts/*.sh
print_status "Development scripts created"

# Create health check script
cat > scripts/health-check.sh << EOF
#!/bin/bash
SERVICE_URL=\${SERVICE_URL:-http://localhost:$SERVICE_PORT}
HEALTH_ENDPOINT="\$SERVICE_URL/health"

echo "Checking service health at \$HEALTH_ENDPOINT"

response=\$(curl -s -o /dev/null -w "%{http_code}" "\$HEALTH_ENDPOINT")

if [ "\$response" = "200" ]; then
    echo "✓ Service is healthy"
    exit 0
else
    echo "✗ Service is unhealthy (HTTP \$response)"
    exit 1
fi
EOF

chmod +x scripts/health-check.sh

# Create README
echo -e "${BLUE}📖 Creating documentation...${NC}"

cat > README.md << EOF
# User Service

Enterprise-grade user management microservice for the LOS (Loan Origination System).

## Features

- **User Management**: Complete CRUD operations for user accounts
- **Profile Management**: Extended user profiles with personal, professional, and financial information
- **KYC Verification**: Know Your Customer verification workflows
- **Document Management**: Secure document upload, storage, and processing
- **Security**: AES encryption, secure password hashing, two-factor authentication
- **Audit Logging**: Comprehensive audit trail for compliance
- **Caching**: Redis-based caching for improved performance

## Architecture

This service follows Clean Architecture principles:

- **Domain Layer**: Core business entities and interfaces
- **Application Layer**: Business logic and use cases
- **Infrastructure Layer**: Data access and external service integration
- **Interfaces Layer**: HTTP handlers and API endpoints

## Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Redis 6 or later
- AWS S3 (for document storage)

## Quick Start

1. **Database Setup**:
   \`\`\`bash
   createdb $DB_NAME
   psql -d $DB_NAME -f migrations/001_create_user_tables.sql
   \`\`\`

2. **Configuration**:
   - Copy \`configs/development.yaml\` and update database credentials
   - Set up AWS S3 credentials for document storage

3. **Run the Service**:
   \`\`\`bash
   ./scripts/dev.sh
   \`\`\`

4. **Health Check**:
   \`\`\`bash
   curl http://localhost:$SERVICE_PORT/health
   \`\`\`

## API Endpoints

### Users
- \`POST /api/v1/users\` - Create user
- \`GET /api/v1/users/{id}\` - Get user by ID
- \`PUT /api/v1/users/{id}\` - Update user
- \`DELETE /api/v1/users/{id}\` - Delete user
- \`GET /api/v1/users\` - List users (with pagination)

### Profiles
- \`GET /api/v1/users/{id}/profile\` - Get user profile
- \`PUT /api/v1/users/{id}/profile\` - Update user profile

### Documents
- \`POST /api/v1/users/{id}/documents\` - Upload document
- \`GET /api/v1/users/{id}/documents\` - List user documents
- \`GET /api/v1/documents/{id}\` - Get document details
- \`GET /api/v1/documents/{id}/download\` - Download document

### KYC
- \`POST /api/v1/users/{id}/kyc/{type}/initiate\` - Initiate KYC verification
- \`GET /api/v1/users/{id}/kyc\` - Get KYC status
- \`POST /api/v1/users/{id}/kyc/{id}/approve\` - Approve KYC (admin)
- \`POST /api/v1/users/{id}/kyc/{id}/reject\` - Reject KYC (admin)

## Development

### Running Tests
\`\`\`bash
./scripts/test.sh
\`\`\`

### Database Migrations
\`\`\`bash
./scripts/migrate.sh
\`\`\`

### Health Check
\`\`\`bash
./scripts/health-check.sh
\`\`\`

## Deployment

### Using Systemd (Linux)
1. Copy service file: \`sudo cp $SERVICE_NAME.service /etc/systemd/system/\`
2. Enable service: \`sudo systemctl enable $SERVICE_NAME\`
3. Start service: \`sudo systemctl start $SERVICE_NAME\`

### Using Docker
\`\`\`bash
docker-compose up --build
\`\`\`

## Configuration

The service uses YAML configuration files. Key settings:

- **Database**: PostgreSQL connection settings
- **Redis**: Cache configuration
- **Storage**: S3 bucket and credentials
- **Security**: Encryption keys and JWT settings
- **Features**: Feature flags for optional functionality

See \`configs/development.yaml\` for full configuration options.

## Security

- All sensitive data is encrypted using AES-256
- Passwords are hashed using bcrypt
- JWT tokens for authentication
- Rate limiting on API endpoints
- Comprehensive audit logging
- Document virus scanning
- Input validation and sanitization

## Monitoring

- Health check endpoint: \`/health\`
- Metrics endpoint: \`/metrics\` (if enabled)
- Structured logging with correlation IDs
- Error tracking and alerting

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Copyright (c) 2024 LOS System. All rights reserved.
EOF

print_status "Documentation created"

# Final status
echo -e "${GREEN}🎉 User Service deployment setup complete!${NC}"
echo
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Update configuration in configs/development.yaml"
echo "2. Set up PostgreSQL database (if not done automatically)"
echo "3. Configure Redis connection"
echo "4. Set up AWS S3 bucket for document storage"
echo "5. Run the service: ./scripts/dev.sh"
echo "6. Test the API: curl http://localhost:$SERVICE_PORT/health"
echo
echo -e "${BLUE}Service Information:${NC}"
echo "- Service Port: $SERVICE_PORT"
echo "- Database: $DB_NAME"
echo "- Health Check: http://localhost:$SERVICE_PORT/health"
echo "- API Base URL: http://localhost:$SERVICE_PORT/api/v1"
echo
echo -e "${YELLOW}Note:${NC} Remember to update production configuration before deploying to production!"
