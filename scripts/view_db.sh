#!/bin/bash

# Colors
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Connecting to PostgreSQL...${NC}"
echo ""
docker exec -it rental_scraper_db psql -U postgres -d rental_scraper