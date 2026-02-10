#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
GITHUB_RAW_URL="https://raw.githubusercontent.com/nus25/gyoka/main/packages/editor/schema/openapi.json"
SCHEMA_FILE="schema/openapi.json"
BACKUP_FILE="schema/openapi.json.backup"

# Get the script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Ensure required tooling exists
if ! command -v jq >/dev/null 2>&1; then
    echo -e "${RED}✗ Missing dependency: jq${NC}"
    echo "Please install jq to validate the downloaded schema."
    exit 1
fi

# Temporary file for atomic updates
TEMP_SCHEMA_FILE="$(mktemp 2>/dev/null)"
if [ -z "$TEMP_SCHEMA_FILE" ]; then
    echo -e "${RED}✗ Failed to create temporary file${NC}"
    exit 1
fi
trap 'rm -f "$TEMP_SCHEMA_FILE"' EXIT

echo -e "${YELLOW}Updating OpenAPI schema from GitHub...${NC}"
echo "Source: $GITHUB_RAW_URL"
echo "Target: $SCHEMA_FILE"
echo ""

# Create backup of current schema
if [ -f "$SCHEMA_FILE" ]; then
    echo -e "${YELLOW}Creating backup...${NC}"
    cp "$SCHEMA_FILE" "$BACKUP_FILE"
    echo -e "${GREEN}✓ Backup created: $BACKUP_FILE${NC}"
fi

# Download the latest schema
echo -e "${YELLOW}Downloading latest schema...${NC}"
if curl -f -s -o "$TEMP_SCHEMA_FILE" "$GITHUB_RAW_URL"; then
    echo -e "${GREEN}✓ Schema downloaded successfully${NC}"
    
    # Validate JSON format
    echo -e "${YELLOW}Validating JSON format...${NC}"
    if jq empty "$TEMP_SCHEMA_FILE" 2>/dev/null; then
        echo -e "${GREEN}✓ JSON validation passed${NC}"
        mv "$TEMP_SCHEMA_FILE" "$SCHEMA_FILE"
        
        # Remove backup if download was successful
        if [ -f "$BACKUP_FILE" ]; then
            rm "$BACKUP_FILE"
            echo -e "${GREEN}✓ Backup removed${NC}"
        fi
        
        echo ""
        echo -e "${GREEN}Schema update completed successfully!${NC}"
    else
        echo -e "${RED}✗ Invalid JSON format${NC}"
        
        # Restore from backup
        if [ -f "$BACKUP_FILE" ]; then
            mv "$BACKUP_FILE" "$SCHEMA_FILE"
            echo -e "${YELLOW}Restored from backup${NC}"
        fi
        
        exit 1
    fi
else
    echo -e "${RED}✗ Failed to download schema${NC}"
    
    # Restore from backup if it exists
    if [ -f "$BACKUP_FILE" ]; then
        mv "$BACKUP_FILE" "$SCHEMA_FILE"
        echo -e "${YELLOW}Restored from backup${NC}"
    fi
    
    exit 1
fi
