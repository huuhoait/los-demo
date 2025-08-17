#!/bin/bash

# Decision Engine Service Runner

set -e

# Load environment variables
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(cat .env | grep -v '#' | awk '/=/ {print $1}')
fi

# Default values
PORT=${PORT:-8082}
ENVIRONMENT=${ENVIRONMENT:-development}
LOG_LEVEL=${LOG_LEVEL:-info}

echo "Starting Decision Engine Service..."
echo "Environment: $ENVIRONMENT"
echo "Port: $PORT"
echo "Log Level: $LOG_LEVEL"

# Build and run the service
if [ "$ENVIRONMENT" = "development" ]; then
    echo "Running in development mode with hot reload..."
    go run cmd/main.go
else
    echo "Building and running in $ENVIRONMENT mode..."
    go build -o decision-engine-service cmd/main.go
    ./decision-engine-service
fi
