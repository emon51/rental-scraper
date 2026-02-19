package scraper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/emon51/rental-scraper/models"
)

type Scraper struct {
	baseURL      string
	maxListings  int
	requestDelay int
}

func NewScraper(baseURL string, maxListings, requestDelay int) *Scraper {
	return &Scraper{
		baseURL:      baseURL,
		maxListings:  maxListings,
		requestDelay: requestDelay,
	}
}

// ScrapeLocation scrapes 1 listing from page 1 and 1 listing from page 2.
func (s *Scraper) ScrapeLocation(ctx context.Context, locationSlug, displayName string) ([]models.Listing, error) {
	var allListings []models.Listing

	// Airbnb pagination: page 1 = offset 0, page 2 = offset 20
	pages := []struct {
		pageNum int
		offset  int
	}{
		{1, 0},
		{2, 20},
	}

	for _, page := range pages {
		url := fmt.Sprintf(s.baseURL, locationSlug)
		if page.offset > 0 {
			url = fmt.Sprintf("%s?items_offset=%d", url, page.offset)
		}

		fmt.Printf("  [%s] Page %d: %s\n", displayName, page.pageNum, url)

		var listings []models.Listing

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(12*time.Second),
			chromedp.Evaluate(s.getExtractionScript(1), &listings), // fetch only 1 per page
		)
		if err != nil {
			fmt.Printf("  WARNING: Failed page %d for %s: %v\n", page.pageNum, displayName, err)
			continue
		}

		fmt.Printf("  Found %d listing on page %d of %s\n", len(listings), page.pageNum, displayName)

		// Set metadata and fetch description
		for i := range listings {
			if listings[i].URL != "" {
				fmt.Printf("    Fetching description (page %d, item %d)...\n", page.pageNum, i+1)
				listings[i].Platform = "Airbnb"
				listings[i].Location = displayName
				listings[i].Description = s.getDescription(ctx, listings[i].URL)
				time.Sleep(time.Duration(s.requestDelay) * time.Second)
			}
		}

		allListings = append(allListings, listings...)

		// Pause between pages
		time.Sleep(3 * time.Second)
	}

	return allListings, nil
}

// getExtractionScript returns JS to extract up to `limit` listings from the current page.
func (s *Scraper) getExtractionScript(limit int) string {
	return fmt.Sprintf(`
		(() => {
			const cards = Array.from(document.querySelectorAll('[itemprop="itemListElement"]')).slice(0, %d);

			return cards.map(card => {
				const link = card.querySelector('a[href*="/rooms/"]');
				const url = link ? link.href : '';

				const titleEl = card.querySelector('[data-testid="listing-card-name"]');
				const title = titleEl ? titleEl.innerText : '';

				let price = '';
				const allSpans = card.querySelectorAll('span');
				for (let span of allSpans) {
					const text = span.innerText.trim();
					if (text.match(/^\$\d+/) || text.match(/^[A-Z]{1,3}\$?\d+/)) {
						price = text.split('\n')[0];
						break;
					}
				}

				let rating = '';
				for (let span of allSpans) {
					const text = span.innerText.trim();
					const match = text.match(/^(\d+\.\d+)/);
					if (match && parseFloat(match[1]) >= 1 && parseFloat(match[1]) <= 5) {
						rating = match[1];
						break;
					}
				}

				return {
					title: title,
					price: price,
					rating: rating,
					url: url
				};
			});
		})()
	`, limit)
}

func (s *Scraper) getDescription(ctx context.Context, url string) string {
	var description string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(6*time.Second),
		chromedp.Evaluate(`
			document.querySelector('[data-section-id="DESCRIPTION_DEFAULT"]')?.innerText || ''
		`, &description),
	)

	if err != nil {
		return ""
	}

	return strings.TrimSpace(description)
}