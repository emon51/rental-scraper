package airbnb

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

func (s *Scraper) Scrape(ctx context.Context) ([]models.Listing, error) {
	var listings []models.Listing

	fmt.Println("Navigating to Airbnb search page...")

	// Navigate and extract listings
	err := chromedp.Run(ctx,
		chromedp.Navigate(s.baseURL),
		chromedp.Sleep(10*time.Second),
		chromedp.Evaluate(s.getExtractionScript(), &listings),
	)

	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d listings\n", len(listings))

	// Fetch descriptions
	fmt.Println("\nFetching descriptions...")
	for i := range listings {
		if listings[i].URL != "" {
			fmt.Printf("  Description %d/%d...\n", i+1, len(listings))
			description := s.getDescription(ctx, listings[i].URL)
			listings[i].Platform = "Airbnb"
			listings[i].Description = description
			listings[i].Location = "Kuala Lumpur, Malaysia"
			time.Sleep(time.Duration(s.requestDelay) * time.Second)
		}
	}

	return listings, nil
}

func (s *Scraper) getExtractionScript() string {
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
					if (text.match(/^\$\d+/)) {
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
	`, s.maxListings)
}

func (s *Scraper) getDescription(ctx context.Context, url string) string {
	var description string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.Evaluate(`
			document.querySelector('[data-section-id="DESCRIPTION_DEFAULT"]')?.innerText || ''
		`, &description),
	)

	if err != nil {
		return ""
	}

	return strings.TrimSpace(description)
}