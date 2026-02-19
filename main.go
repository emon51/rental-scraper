package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/emon51/rental-scraper/config"
	"github.com/emon51/rental-scraper/models"
	scraper "github.com/emon51/rental-scraper/scraper"
	"github.com/emon51/rental-scraper/services"
	"github.com/emon51/rental-scraper/storage"
)

func main() {
	fmt.Println("Airbnb Rental Scraper Starting...")

	// Load configuration
	cfg := config.NewConfig()

	// Setup browser context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cfg.Headless),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(cfg.PageTimeout)*time.Second)
	defer cancel()

	// Create scraper
	s := scraper.NewScraper(cfg.BaseURL, cfg.MaxListings, cfg.RequestDelay)

	// Scrape all locations
	fmt.Println("\n=== STEP 1: SCRAPING ===")
	var allListings []models.Listing

	for _, loc := range cfg.Locations {
		fmt.Printf("\nScraping: %s\n", loc.DisplayName)
		listings, err := s.ScrapeLocation(ctx, loc.Slug, loc.DisplayName)
		if err != nil {
			fmt.Printf("  WARNING: Failed to scrape %s: %v\n", loc.DisplayName, err)
			continue
		}
		allListings = append(allListings, listings...)

		// Small pause between locations to avoid rate limiting
		time.Sleep(3 * time.Second)
	}

	fmt.Printf("\nRaw listings scraped: %d\n", len(allListings))

	// Filter and clean data
	fmt.Println("\n=== STEP 2: FILTERING & CLEANING ===")
	filter := services.NewFilter()
	cleanedListings := filter.CleanListings(allListings)
	fmt.Printf("Cleaned listings: %d\n", len(cleanedListings))

	// Save to CSV
	fmt.Println("\n=== STEP 3: SAVING TO CSV ===")
	csvWriter := storage.NewCSVWriter("listings.csv")
	if err := csvWriter.WriteListings(cleanedListings); err != nil {
		log.Fatalf("Failed to save CSV: %v", err)
	}
	fmt.Println("✓ Data saved to listings.csv")

	// Save to PostgreSQL
	fmt.Println("\n=== STEP 4: SAVING TO POSTGRESQL ===")
	pgWriter, err := storage.NewPostgresWriter(
		cfg.DBConfig.Host,
		cfg.DBConfig.Port,
		cfg.DBConfig.User,
		cfg.DBConfig.Password,
		cfg.DBConfig.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgWriter.Close()

	if err := pgWriter.CreateTable(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	fmt.Println("✓ Database table created")

	if err := pgWriter.InsertListings(cleanedListings); err != nil {
		log.Fatalf("Failed to insert listings: %v", err)
	}
	fmt.Printf("✓ %d listings saved to PostgreSQL\n", len(cleanedListings))

	// Generate insights
	fmt.Println("\n=== STEP 5: GENERATING INSIGHTS ===")
	insightGen := services.NewInsightGenerator()
	insights := insightGen.Generate(cleanedListings)
	insightGen.PrintReport(insights)
}