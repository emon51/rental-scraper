package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/emon51/rental-scraper/config"
	"github.com/emon51/rental-scraper/models"
	"github.com/emon51/rental-scraper/scraper"
)

type ScraperService struct {
	cfg *config.Config
}

func NewScraperService(cfg *config.Config) *ScraperService {
	return &ScraperService{cfg: cfg}
}

// ScrapeAll collects listings from all configured locations concurrently
func (ss *ScraperService) ScrapeAll(ctx context.Context) ([]models.Listing, error) {
	fmt.Println("\n=== STEP 1: SCRAPING (CONCURRENT) ===")

	s := scraper.NewScraper(
		ss.cfg.BaseURL,
		ss.cfg.ListingsPerPage,
		ss.cfg.PagesToScrape,
		ss.cfg.RequestDelay,
	)

	// Channel to collect listings
	listingsChan := make(chan []models.Listing, len(ss.cfg.Locations))

	// WaitGroup to wait for all goroutines
	var wg sync.WaitGroup

	// Semaphore to limit concurrent scrapers
	semaphore := make(chan struct{}, ss.cfg.MaxConcurrent)

	// Launch goroutines for each location
	for i, loc := range ss.cfg.Locations {
		wg.Add(1)

		go func(index int, location config.LocationConfig) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("\n[%d/%d] Scraping: %s\n", index+1, len(ss.cfg.Locations), location.DisplayName)

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

	// Step 2: Clean data
	fmt.Println("\n=== STEP 2: FILTERING & CLEANING ===")
	filter := NewFilter()
	cleaned := filter.CleanListings(allListings)
	fmt.Printf("Cleaned listings: %d\n", len(cleaned))

	return cleaned, nil
}