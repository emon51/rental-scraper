package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/emon51/rental-scraper/models"
)

type PostgresWriter struct {
	db *sql.DB
}

func NewPostgresWriter(host string, port int, user, password, dbname string) (*PostgresWriter, error) {
	// Use parameterized connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresWriter{db: db}, nil
}

func (w *PostgresWriter) Close() error {
	return w.db.Close()
}

// CreateTable creates table with proper indexes - SQL injection safe
func (w *PostgresWriter) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS listings (
		id SERIAL PRIMARY KEY,
		platform VARCHAR(50) NOT NULL,
		title TEXT NOT NULL,
		price NUMERIC(10, 2),
		location VARCHAR(255),
		rating NUMERIC(3, 2),
		url TEXT UNIQUE NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Indexes on important fields for query performance
	CREATE INDEX IF NOT EXISTS idx_listings_price ON listings(price);
	CREATE INDEX IF NOT EXISTS idx_listings_location ON listings(location);
	CREATE INDEX IF NOT EXISTS idx_listings_rating ON listings(rating);
	CREATE INDEX IF NOT EXISTS idx_listings_platform ON listings(platform);
	`

	_, err := w.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// InsertListings uses parameterized queries to prevent SQL injection
func (w *PostgresWriter) InsertListings(listings []models.Listing) error {
	if len(listings) == 0 {
		return nil
	}

	// Use transaction for batch insert
	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Parameterized query - prevents SQL injection
	stmt, err := tx.Prepare(`
		INSERT INTO listings (platform, title, price, location, rating, url, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (url) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Batch insert with parameterized values
	for _, listing := range listings {
		var price *float64
		if listing.Price != "" {
			p := parsePrice(listing.Price)
			if p > 0 {
				price = &p
			}
		}

		var rating *float64
		if listing.Rating != "" {
			r := parseRating(listing.Rating)
			if r > 0 {
				rating = &r
			}
		}

		// Execute with bound parameters - SQL injection safe
		_, err := stmt.Exec(
			listing.Platform,
			listing.Title,
			price,
			listing.Location,
			rating,
			listing.URL,
			listing.Description,
		)
		if err != nil {
			return fmt.Errorf("failed to insert listing: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAllListings retrieves all listings - uses parameterized query
func (w *PostgresWriter) GetAllListings() ([]models.Listing, error) {
	query := `
		SELECT platform, title, price, location, rating, url, description
		FROM listings
		ORDER BY created_at DESC
	`

	rows, err := w.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query listings: %w", err)
	}
	defer rows.Close()

	listings := make([]models.Listing, 0)

	for rows.Next() {
		var listing models.Listing
		var price, rating sql.NullFloat64

		err := rows.Scan(
			&listing.Platform,
			&listing.Title,
			&price,
			&listing.Location,
			&rating,
			&listing.URL,
			&listing.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if price.Valid {
			listing.Price = fmt.Sprintf("%.2f", price.Float64)
		}

		if rating.Valid {
			listing.Rating = fmt.Sprintf("%.2f", rating.Float64)
		}

		listings = append(listings, listing)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return listings, nil
}

func parsePrice(price string) float64 {
	var p float64
	fmt.Sscanf(price, "%f", &p)
	return p
}

func parseRating(rating string) float64 {
	var r float64
	fmt.Sscanf(rating, "%f", &r)
	return r
}