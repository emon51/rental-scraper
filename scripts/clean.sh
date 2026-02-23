#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Cleaning up scraper data...${NC}"
echo ""

# Remove CSV files
if [ -f "listings.csv" ]; then
    rm listings.csv
    echo -e "${GREEN}✓${NC} Removed listings.csv"
fi

# Remove log files
if [ -f "scraper.log" ]; then
    rm scraper.log
    echo -e "${GREEN}✓${NC} Removed scraper.log"
fi

# Clear PostgreSQL data
echo ""
read -p "Do you want to clear PostgreSQL data? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker exec -it rental_scraper_db psql -U postgres -d rental_scraper -c "TRUNCATE TABLE listings;" 2>/dev/null
    echo -e "${GREEN}✓${NC} PostgreSQL data cleared"
fi

echo ""
echo -e "${GREEN}Cleanup complete!${NC}"