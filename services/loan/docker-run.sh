#!/bin/bash

# Loan Service Docker Management Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build     Build the Docker image"
    echo "  run       Run the container"
    echo "  stop      Stop the container"
    echo "  restart   Restart the container"
    echo "  logs      Show container logs"
    echo "  shell     Open shell in running container"
    echo "  clean     Remove container and image"
    echo "  compose   Use docker-compose (build and run all services)"
    echo "  compose-stop  Stop docker-compose services"
    echo "  compose-logs   Show docker-compose logs"
    echo "  db-shell      Open shell in PostgreSQL container"
    echo "  conductor-ui      Open Conductor UI in browser"
    echo "  deploy-workflows  Deploy workflows to Conductor server"
    echo "  check-conductor   Check Conductor server and UI status"
    echo "  health            Check service health"
    echo "  help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build"
    echo "  $0 run"
    echo "  $0 compose"
    echo "  $0 health"
    echo "  $0 conductor-ui"
}

# Function to build Docker image
build_image() {
    print_status "Building Docker image..."
    docker build -t loan-service:latest .
    print_success "Docker image built successfully!"
}

# Function to run container
run_container() {
    print_status "Running loan service container..."
    docker run -d \
        --name loan-service \
        -p 8080:8080 \
        -e PORT=8080 \
        -e HOST=0.0.0.0 \
        -e LOG_LEVEL=info \
        -e LOG_FORMAT=json \
        --restart unless-stopped \
        loan-service:latest
    
    print_success "Container started successfully!"
    print_status "Service will be available at http://localhost:8080"
    print_status "Health check: http://localhost:8080/v1/health"
}

# Function to stop container
stop_container() {
    print_status "Stopping loan service container..."
    docker stop loan-service 2>/dev/null || print_warning "Container not running"
    docker rm loan-service 2>/dev/null || print_warning "Container not found"
    print_success "Container stopped and removed!"
}

# Function to restart container
restart_container() {
    stop_container
    run_container
}

# Function to show logs
show_logs() {
    print_status "Showing container logs..."
    docker logs -f loan-service
}

# Function to open shell
open_shell() {
    print_status "Opening shell in container..."
    docker exec -it loan-service /bin/sh
}

# Function to clean up
clean_up() {
    print_status "Cleaning up Docker resources..."
    stop_container
    docker rmi loan-service:latest 2>/dev/null || print_warning "Image not found"
    print_success "Cleanup completed!"
}

# Function to use docker-compose
use_compose() {
    print_status "Using docker compose to build and run..."
    docker compose up -d --build
    print_success "Services started with docker compose!"
    print_status "Service will be available at http://localhost:8080"
}

# Function to stop docker-compose
stop_compose() {
    print_status "Stopping docker compose services..."
    docker compose down
    print_success "Docker compose services stopped!"
}

# Function to show docker-compose logs
show_compose_logs() {
    print_status "Showing docker compose logs..."
    docker compose logs -f
}

# Function to deploy workflows to Conductor
deploy_workflows() {
    print_status "Deploying workflows to Conductor..."
    
    # Wait for Conductor to be ready
    print_status "Waiting for Conductor server to be ready..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s http://localhost:8082/health > /dev/null; then
            print_success "Conductor server is ready!"
            break
        fi
        
        print_status "Attempt $attempt/$max_attempts - Waiting 10 seconds..."
        sleep 10
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        print_error "Conductor server did not become ready!"
        return 1
    fi
    
    # Deploy workflows
    if [ -f "./workflows/deploy-conductor.sh" ]; then
        print_status "Running workflow deployment script..."
        ./workflows/deploy-conductor.sh -s http://localhost:8080
    else
        print_error "Workflow deployment script not found!"
        return 1
    fi
}

# Function to check Conductor status
check_conductor() {
    print_status "Checking Conductor server status..."
    
    if curl -f -s http://localhost:8080/health > /dev/null; then
        print_success "Conductor Server is healthy!"
        echo "Health response:"
        curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health
    else
        print_error "Conductor Server is not responding!"
    fi
    
    print_status "Checking Conductor UI..."
    if curl -f -s http://localhost:3000 > /dev/null; then
        print_success "Conductor UI is accessible!"
    else
        print_warning "Conductor UI is not responding!"
    fi
}

# Function to open database shell
open_db_shell() {
    print_status "Opening PostgreSQL shell..."
    docker compose exec postgres psql -U postgres -d loan_service
}

# Function to open Conductor UI
open_conductor_ui() {
    print_status "Opening Conductor UI in browser..."
    if command -v open >/dev/null 2>&1; then
        open http://localhost:3000
    elif command -v xdg-open >/dev/null 2>&1; then
        xdg-open http://localhost:3000
    else
        print_status "Please open http://localhost:3000 in your browser"
    fi
}

# Function to check health
check_health() {
    print_status "Checking service health..."
    
    # Wait a moment for service to start
    sleep 3
    
    # Check loan service health
    if curl -f -s http://localhost:8081/v1/health > /dev/null; then
        print_success "Loan Service is healthy!"
        echo "Health response:"
        curl -s http://localhost:8081/v1/health | jq . 2>/dev/null || curl -s http://localhost:8081/v1/health
    else
        print_error "Loan Service is not responding!"
    fi
    
    # Check Conductor server health
    if curl -f -s http://localhost:8082/health > /dev/null; then
        print_success "Conductor Server is healthy!"
    else
        print_warning "Conductor Server is not responding!"
    fi
    
    # Check PostgreSQL
    if docker compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
        print_success "PostgreSQL is healthy!"
    else
        print_warning "PostgreSQL is not responding!"
    fi
}

# Main script logic
case "${1:-help}" in
    build)
        build_image
        ;;
    run)
        run_container
        ;;
    stop)
        stop_container
        ;;
    restart)
        restart_container
        ;;
    logs)
        show_logs
        ;;
    shell)
        open_shell
        ;;
    clean)
        clean_up
        ;;
    compose)
        use_compose
        ;;
    compose-stop)
        stop_compose
        ;;
    compose-logs)
        show_compose_logs
        ;;
    db-shell)
        open_db_shell
        ;;
    conductor-ui)
        open_conductor_ui
        ;;
    deploy-workflows)
        deploy_workflows
        ;;
    check-conductor)
        check_conductor
        ;;
    health)
        check_health
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        print_error "Unknown command: $1"
        show_usage
        exit 1
        ;;
esac
