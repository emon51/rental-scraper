#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Banner
echo "=================================================="
echo "   Airbnb Rental Scraper"
echo "=================================================="
echo ""

# Check if Go is installed
print_info "Checking Go installation..."
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi
print_success "Go is installed: $(go version)"
echo ""

# Check if Docker is installed
print_info "Checking Docker installation..."
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker."
    exit 1
fi
print_success "Docker is installed: $(docker --version)"
echo ""

# Check if docker-compose is installed
print_info "Checking Docker Compose installation..."
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi
print_success "Docker Compose is installed: $(docker-compose --version)"
echo ""

# Install Go dependencies
print_info "Installing Go dependencies..."
if go mod download; then
    print_success "Dependencies installed successfully"
else
    print_error "Failed to install dependencies"
    exit 1
fi
echo ""

# Check if PostgreSQL container is running
print_info "Checking PostgreSQL container status..."
if docker ps | grep -q rental_scraper_db; then
    print_success "PostgreSQL container is already running"
else
    print_warning "PostgreSQL container is not running"
    print_info "Starting PostgreSQL container..."
    
    if docker-compose up -d; then
        print_success "PostgreSQL container started successfully"
        print_info "Waiting for PostgreSQL to be ready (10 seconds)..."
        sleep 10
    else
        print_error "Failed to start PostgreSQL container"
        exit 1
    fi
fi
echo ""

# Check PostgreSQL connection
print_info "Verifying PostgreSQL connection..."
if docker exec rental_scraper_db pg_isready -U postgres &> /dev/null; then
    print_success "PostgreSQL is ready"
else
    print_warning "PostgreSQL may not be fully ready yet"
fi
echo ""

# Run the scraper
print_info "Starting Airbnb Rental Scraper..."
echo "=================================================="
echo ""

if go run main.go; then
    echo ""
    echo "=================================================="
    print_success "Scraping completed successfully!"
    echo ""
    
    # Show output files
    print_info "Output files:"
    if [ -f "listings.csv" ]; then
        echo "  - listings.csv ($(wc -l < listings.csv) lines)"
    fi
    if [ -f "scraper.log" ]; then
        echo "  - scraper.log ($(wc -l < scraper.log) lines)"
    fi
    echo ""
    
    # Show database stats
    print_info "Database statistics:"
    docker exec -it rental_scraper_db psql -U postgres -d rental_scraper -c "SELECT COUNT(*) as total_listings FROM listings;" 2>/dev/null
    echo ""
    
    print_info "To view data in PostgreSQL, run:"
    echo "  docker exec -it rental_scraper_db psql -U postgres -d rental_scraper"
    echo ""
    
else
    echo ""
    print_error "Scraping failed. Check scraper.log for details."
    exit 1
fi