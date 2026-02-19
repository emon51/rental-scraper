package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/emon51/rental-scraper/config"
	"github.com/emon51/rental-scraper/models"
	"github.com/emon51/rental-scraper/scraper"
	"github.com/emon51/rental-scraper/services"
	"github.com/emon51/rental-scraper/storage"
)

func main() {
	startTime := time.Now()
	
	fmt.Println("Airbnb Rental Scraper Starting...")

	cfg := config.NewConfig()

	// Setup browser context
	ctx := setupBrowserContext(cfg)

	// Execute scraping pipeline
	totalListings := 0
	if err := runScrapingPipeline(ctx, cfg, &totalListings); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✓ Scraping Complete! (Duration: %v)\n", duration)
}

// setupBrowserContext creates and configures the browser context
func setupBrowserContext(cfg *config.Config) context.Context {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cfg.Headless),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)
	ctx, _ = context.WithTimeout(ctx, time.Duration(cfg.PageTimeout)*time.Second)

	return ctx
}

// runScrapingPipeline executes the complete data collection and processing pipeline
func runScrapingPipeline(ctx context.Context, cfg *config.Config, totalListings *int) error {
	// Step 1: Scrape data
	rawListings, err := scrapeListingsConcurrently(ctx, cfg)
	if err != nil {
		return fmt.Errorf("scraping failed: %w", err)
	}
	*totalListings = len(rawListings)

	// Step 2: Clean and filter
	cleanedListings := cleanData(rawListings)

	// Step 3: Save to CSV
	if err := saveToCSV(cleanedListings); err != nil {
		return fmt.Errorf("CSV save failed: %w", err)
	}

	// Step 4: Save to PostgreSQL
	if err := saveToDatabase(cleanedListings, cfg); err != nil {
		return fmt.Errorf("database save failed: %w", err)
	}

	// Step 5: Generate insights
	generateInsights(cleanedListings)

	return nil
}

// scrapeListingsConcurrently collects listings from all configured locations using goroutines
func scrapeListingsConcurrently(ctx context.Context, cfg *config.Config) ([]models.Listing, error) {
	fmt.Println("\n=== STEP 1: SCRAPING (CONCURRENT) ===")

	s := scraper.NewScraper(cfg.BaseURL, cfg.ListingsPerPage, cfg.PagesToScrape, cfg.RequestDelay)
	
	// Channel to collect listings
	listingsChan := make(chan []models.Listing, len(cfg.Locations))
	
	// WaitGroup to wait for all goroutines
	var wg sync.WaitGroup
	
	// Semaphore to limit concurrent scrapers
	maxConcurrent := cfg.MaxConcurrent
	semaphore := make(chan struct{}, maxConcurrent)

	// Launch goroutines for each location
	for i, loc := range cfg.Locations {
		wg.Add(1)
		
		go func(index int, location config.LocationConfig) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("\n[%d/%d] Scraping: %s\n", index+1, len(cfg.Locations), location.DisplayName)

			listings, err := s.ScrapeLocation(ctx, location.Slug, location.DisplayName)
			if err != nil {
				fmt.Printf("  WARNING: Failed to scrape %s: %v\n", location.DisplayName, err)
				listingsChan <- []models.Listing{}
				return
			}

			fmt.Printf("✓ Collected %d listings from %s\n", len(listings), location.DisplayName)
			listingsChan <- listings
		}(i, loc)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(listingsChan)
	}()

	// Collect all listings
	allListings := make([]models.Listing, 0)
	for listings := range listingsChan {
		allListings = append(allListings, listings...)
	}

	fmt.Printf("\nRaw listings scraped: %d\n", len(allListings))
	return allListings, nil
}

// cleanData filters and validates scraped data
func cleanData(listings []models.Listing) []models.Listing {
	fmt.Println("\n=== STEP 2: FILTERING & CLEANING ===")

	filter := services.NewFilter()
	cleaned := filter.CleanListings(listings)

	fmt.Printf("Cleaned listings: %d\n", len(cleaned))
	return cleaned
}

// saveToCSV exports listings to CSV file
func saveToCSV(listings []models.Listing) error {
	fmt.Println("\n=== STEP 3: SAVING TO CSV ===")

	csvWriter := storage.NewCSVWriter("listings.csv")
	if err := csvWriter.WriteListings(listings); err != nil {
		return err
	}

	fmt.Println("✓ Data saved to listings.csv")
	return nil
}

// saveToDatabase persists listings to PostgreSQL
func saveToDatabase(listings []models.Listing, cfg *config.Config) error {
	fmt.Println("\n=== STEP 4: SAVING TO POSTGRESQL ===")

	pgWriter, err := storage.NewPostgresWriter(
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.DBName,
	)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer pgWriter.Close()

	if err := pgWriter.CreateTable(); err != nil {
		return fmt.Errorf("table creation failed: %w", err)
	}
	fmt.Println("✓ Database table created")

	if err := pgWriter.InsertListings(listings); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	fmt.Printf("✓ %d listings saved to PostgreSQL\n", len(listings))

	return nil
}

// generateInsights calculates and displays statistics
func generateInsights(listings []models.Listing) {
	fmt.Println("\n=== STEP 5: GENERATING INSIGHTS ===")

	insightGen := services.NewInsightGenerator()
	insights := insightGen.Generate(listings)
	insightGen.PrintReport(insights)
}