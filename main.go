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
	listings, err := scraper.Scrape(ctx)
	if err != nil {
		log.Fatalf("Scraping failed: %v", err)
	}

	fmt.Printf("\nTotal listings scraped: %d\n", len(listings))

	// Save to CSV
	csvWriter := storage.NewCSVWriter("listings.csv")
	if err := csvWriter.WriteListings(listings); err != nil {
		log.Fatalf("Failed to save CSV: %v", err)
	}

	fmt.Println("Data saved to listings.csv")
}