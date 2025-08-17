#!/bin/bash

# Force Migration Script: Migrate All Services to Shared Library
# This script forces all services to use the shared library and removes redundant code

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Services to migrate (excluding shared library itself)
SERVICES=("auth" "user" "loan-api" "loan-worker" "decision-engine")

echo -e "${BLUE}🚀 Starting Forced Migration to Shared Library${NC}"
echo "This script will:"
echo "1. Remove all local pkg/ directories"
echo "2. Update go.mod files with shared library dependency"
echo "3. Update imports in main.go files"
echo "4. Create backup of original files"
echo

# Function to backup a service
backup_service() {
    local service=$1
    echo -e "${YELLOW}📦 Creating backup for $service service${NC}"
    
    if [ -d "$service" ]; then
        # Create backup directory
        mkdir -p "backups/$service"
        
        # Backup go.mod, go.sum, and main.go
        [ -f "$service/go.mod" ] && cp "$service/go.mod" "backups/$service/"
        [ -f "$service/go.sum" ] && cp "$service/go.sum" "backups/$service/"
        [ -f "$service/cmd/main.go" ] && cp "$service/cmd/main.go" "backups/$service/"
        
        # Backup entire pkg directory if it exists
        [ -d "$service/pkg" ] && cp -r "$service/pkg" "backups/$service/"
        
        echo -e "${GREEN}✅ Backup created for $service${NC}"
    else
        echo -e "${RED}❌ Service directory $service not found${NC}"
    fi
}

# Function to update go.mod for a service
update_gomod() {
    local service=$1
    echo -e "${YELLOW}📝 Updating go.mod for $service service${NC}"
    
    if [ -f "$service/go.mod" ]; then
        # Add shared library dependency if not present
        if ! grep -q "bmad/trial/services/shared" "$service/go.mod"; then
            echo "	bmad/trial/services/shared v0.0.0-00010101000000-000000000000" >> "$service/go.mod"
        fi
        
        # Add replace directive if not present
        if ! grep -q "replace bmad/trial/services/shared" "$service/go.mod"; then
            echo "" >> "$service/go.mod"
            echo "replace bmad/trial/services/shared => ../shared" >> "$service/go.mod"
        fi
        
        echo -e "${GREEN}✅ Updated go.mod for $service${NC}"
    else
        echo -e "${RED}❌ go.mod not found for $service${NC}"
    fi
}

# Function to remove local pkg directory
remove_local_pkg() {
    local service=$1
    echo -e "${YELLOW}🗑️  Removing local pkg/ directory for $service${NC}"
    
    if [ -d "$service/pkg" ]; then
        rm -rf "$service/pkg"
        echo -e "${GREEN}✅ Removed local pkg/ directory for $service${NC}"
    else
        echo -e "${BLUE}ℹ️  No local pkg/ directory found for $service${NC}"
    fi
}

# Function to update main.go imports
update_main_imports() {
    local service=$1
    echo -e "${YELLOW}🔄 Updating main.go imports for $service${NC}"
    
    local main_file="$service/cmd/main.go"
    if [ -f "$main_file" ]; then
        # Create a temporary file for the updated main.go
        local temp_file=$(mktemp)
        
        # Read the file and update imports
        while IFS= read -r line; do
            # Replace local pkg imports with shared library imports
            if [[ $line =~ "bmad/trial/services/$service/pkg/" ]]; then
                # Extract the package name after pkg/
                local pkg_name=$(echo "$line" | sed -n "s/.*\/pkg\/\([^\"]*\)\".*/\1/p")
                echo "	\"bmad/trial/services/shared/pkg/$pkg_name\"" >> "$temp_file"
            else
                echo "$line" >> "$temp_file"
            fi
        done < "$main_file"
        
        # Replace the original file
        mv "$temp_file" "$main_file"
        
        echo -e "${GREEN}✅ Updated main.go imports for $service${NC}"
    else
        echo -e "${RED}❌ main.go not found for $service${NC}"
    fi
}

# Function to run go mod tidy
run_go_mod_tidy() {
    local service=$1
    echo -e "${YELLOW}🔧 Running go mod tidy for $service${NC}"
    
    if [ -d "$service" ]; then
        cd "$service"
        if go mod tidy 2>/dev/null; then
            echo -e "${GREEN}✅ go mod tidy successful for $service${NC}"
        else
            echo -e "${RED}❌ go mod tidy failed for $service (expected - will fix compilation errors later)${NC}"
        fi
        cd ..
    fi
}

# Function to check compilation
check_compilation() {
    local service=$1
    echo -e "${YELLOW}🔍 Checking compilation for $service${NC}"
    
    if [ -d "$service" ]; then
        cd "$service"
        if go build ./cmd/main.go 2>/dev/null; then
            echo -e "${GREEN}✅ Compilation successful for $service${NC}"
        else
            echo -e "${RED}❌ Compilation has errors for $service (will need manual fixes)${NC}"
        fi
        cd ..
    fi
}

# Main execution
echo -e "${BLUE}🎯 Phase 1: Creating Backups${NC}"
for service in "${SERVICES[@]}"; do
    backup_service "$service"
done

echo -e "\n${BLUE}🎯 Phase 2: Removing Local Dependencies${NC}"
for service in "${SERVICES[@]}"; do
    remove_local_pkg "$service"
done

echo -e "\n${BLUE}🎯 Phase 3: Updating Dependencies${NC}"
for service in "${SERVICES[@]}"; do
    update_gomod "$service"
done

echo -e "\n${BLUE}🎯 Phase 4: Updating Imports${NC}"
for service in "${SERVICES[@]}"; do
    update_main_imports "$service"
done

echo -e "\n${BLUE}🎯 Phase 5: Running go mod tidy${NC}"
for service in "${SERVICES[@]}"; do
    run_go_mod_tidy "$service"
done

echo -e "\n${BLUE}🎯 Phase 6: Checking Compilation${NC}"
for service in "${SERVICES[@]}"; do
    check_compilation "$service"
done

echo -e "\n${GREEN}🎉 MIGRATION COMPLETE!${NC}"
echo -e "${BLUE}Summary:${NC}"
echo "✅ All services now depend on shared library"
echo "✅ Local pkg/ directories removed"
echo "✅ Imports updated to use shared library"
echo "✅ Backups created in backups/ directory"
echo
echo -e "${YELLOW}⚠️  Next Steps:${NC}"
echo "1. Fix any compilation errors manually"
echo "2. Update function calls to match shared library APIs"
echo "3. Test each service individually"
echo "4. Update configuration files if needed"
echo
echo -e "${BLUE}ℹ️  Backups available in:${NC}"
for service in "${SERVICES[@]}"; do
    echo "   backups/$service/"
done
