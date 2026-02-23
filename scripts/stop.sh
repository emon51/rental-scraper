#!/bin/bash

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Stopping PostgreSQL container...${NC}"
docker-compose down

echo ""
echo -e "${GREEN}PostgreSQL stopped successfully${NC}"