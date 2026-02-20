package scraper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/emon51/rental-scraper/config"
	"github.com/emon51/rental-scraper/models"
)

type Scraper struct {
	baseURL              string
	listingsPerPage      int
	pagesToScrape        int
	requestDelay         int
	descriptionConfig    config.DescriptionFetchConfig
}

func NewScraper(baseURL string, listingsPerPage, pagesToScrape, requestDelay int, descConfig config.DescriptionFetchConfig) *Scraper {
	return &Scraper{
		baseURL:           baseURL,
		listingsPerPage:   listingsPerPage,
		pagesToScrape:     pagesToScrape,
		requestDelay:      requestDelay,
		descriptionConfig: descConfig,
	}
}

// ScrapeLocation scrapes multiple pages from a location
func (s *Scraper) ScrapeLocation(ctx context.Context, locationSlug, displayName string) ([]models.Listing, error) {
	var allListings []models.Listing

	for page := 1; page <= s.pagesToScrape; page++ {
		offset := (page - 1) * AirbnbPageOffset

		url := s.buildURL(locationSlug, offset)

		fmt.Printf("  [%s] Page %d: Fetching %d listings...\n", displayName, page, s.listingsPerPage)

		listings, err := s.fetchListingsFromPage(ctx, url)
		if err != nil {
			fmt.Printf("  WARNING: Failed page %d for %s: %v\n", page, displayName, err)
			continue
		}

		fmt.Printf("  Found %d listings on page %d of %s\n", len(listings), page, displayName)

		// Set metadata
		s.setListingMetadata(listings, displayName)

		allListings = append(allListings, listings...)

		// Pause between pages
		if page < s.pagesToScrape {
			time.Sleep(3 * time.Second)
		}
	}

	// Fetch descriptions concurrently
	fmt.Printf("  Fetching descriptions concurrently for %s...\n", displayName)
	s.fetchDescriptionsConcurrently(ctx, allListings)

	return allListings, nil
}

// buildURL constructs the Airbnb search URL with pagination
func (s *Scraper) buildURL(locationSlug string, offset int) string {
	url := fmt.Sprintf(s.baseURL, locationSlug)
	if offset > 0 {
		url = fmt.Sprintf("%s?items_offset=%d", url, offset)
	}
	return url
}

// fetchListingsFromPage extracts listings from a single page
func (s *Scraper) fetchListingsFromPage(ctx context.Context, url string) ([]models.Listing, error) {
	var listings []models.Listing

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(12*time.Second),
		chromedp.Evaluate(s.getExtractionScript(), &listings),
	)

	return listings, err
}

// setListingMetadata adds platform and location to listings
func (s *Scraper) setListingMetadata(listings []models.Listing, location string) {
	for i := range listings {
		listings[i].Platform = "Airbnb"
		listings[i].Location = location
	}
}

// fetchDescriptionsConcurrently fetches descriptions in parallel with rate limiting
func (s *Scraper) fetchDescriptionsConcurrently(ctx context.Context, listings []models.Listing) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.descriptionConfig.MaxConcurrent)

	for i := range listings {
		if listings[i].URL == "" {
			continue
		}

		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("    [%d/%d] Fetching description...\n", index+1, len(listings))

			listings[index].Description = s.getDescription(ctx, listings[index].URL)

			// Rate limiting
			time.Sleep(time.Duration(s.requestDelay) * time.Second)
		}(i)
	}

	wg.Wait()
}

// getExtractionScript returns JavaScript to extract listings
func (s *Scraper) getExtractionScript() string {
	return fmt.Sprintf(ExtractionScriptTemplate, s.listingsPerPage)
}

// getDescription fetches description from a listing detail page
func (s *Scraper) getDescription(ctx context.Context, url string) string {
	// Create timeout context for description fetch
	descCtx, cancel := context.WithTimeout(ctx, time.Duration(s.descriptionConfig.Timeout)*time.Second)
	defer cancel()

	var description string

	err := chromedp.Run(descCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.Evaluate(fmt.Sprintf(`
			document.querySelector('%s')?.innerText || ''
		`, DescriptionSelector), &description),
	)

	if err != nil {
		return ""
	}

	return strings.TrimSpace(description)
}