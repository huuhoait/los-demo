#!/bin/bash

echo "🚀 Starting Authentication Service..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start dependencies
echo "📦 Starting PostgreSQL and Redis..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are healthy
if ! docker-compose ps | grep -q "healthy"; then
    echo "⚠️  Services may not be fully ready yet. Continuing anyway..."
fi

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=los_auth
export DB_USER=postgres
export DB_PASSWORD=password
export DB_SSL_MODE=disable
export REDIS_HOST=localhost
export REDIS_PORT=6379
export JWT_SIGNING_KEY=your-secret-key-change-in-production
export LOG_LEVEL=info

echo "🔧 Environment configured"
echo "🌐 Starting service on port 8080..."

# Run the service
go run cmd/main.go
