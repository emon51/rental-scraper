package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/emon51/rental-scraper/config"
	"github.com/emon51/rental-scraper/scraper"
	"github.com/emon51/rental-scraper/storage"
	"github.com/emon51/rental-scraper/services"

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
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(cfg.PageTimeout)*time.Second)
	defer cancel()

	// Create scraper
	scraper := airbnb.NewScraper(cfg.BaseURL, cfg.MaxListings, cfg.RequestDelay)

	// Scrape listings
	fmt.Println("\n=== STEP 1: SCRAPING ===")
	listings, err := scraper.Scrape(ctx)
	if err != nil {
		log.Fatalf("Scraping failed: %v", err)
	}

	fmt.Printf("Raw listings scraped: %d\n", len(listings))

	// Filter and clean data
	fmt.Println("\n=== STEP 2: FILTERING & CLEANING ===")
	filter := services.NewFilter()
	cleanedListings := filter.CleanListings(listings)
	fmt.Printf("Cleaned listings: %d\n", len(cleanedListings))

	// Save to CSV
	fmt.Println("\n=== STEP 3: SAVING TO CSV ===")
	csvWriter := storage.NewCSVWriter("listings.csv")
	if err := csvWriter.WriteListings(cleanedListings); err != nil {
		log.Fatalf("Failed to save CSV: %v", err)
	}
	fmt.Println("Data saved to listings.csv")

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

	// Create table
	if err := pgWriter.CreateTable(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	fmt.Println("Database table created")

	// Insert listings
	if err := pgWriter.InsertListings(cleanedListings); err != nil {
		log.Fatalf("Failed to insert listings: %v", err)
	}
	fmt.Printf("%d listings saved to PostgreSQL\n", len(cleanedListings))

	// Generate insights
	fmt.Println("\n=== STEP 5: GENERATING INSIGHTS ===")
	insightGen := services.NewInsightGenerator()
	insights := insightGen.Generate(cleanedListings)
	insightGen.PrintReport(insights)

}