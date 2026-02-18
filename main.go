package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Listing struct {
	Title       string
	Price       string
	Location    string
	Rating      string
	URL         string
	Description string
}

func main() {
	fmt.Println("Airbnb Rental Scraper Starting...")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	baseURL := "https://www.airbnb.com/s/Kuala-Lumpur/homes"

	fmt.Println("Navigating to Airbnb search page...")

	var rawListings []map[string]string
	var listings []Listing

	// Navigate and extract 10 listings
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL),
		chromedp.Sleep(10*time.Second),
		chromedp.Evaluate(`
			(() => {
				const cards = Array.from(document.querySelectorAll('[itemprop="itemListElement"]')).slice(0, 10);
				
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
						if (match) {
							const value = parseFloat(match[1]);
							if (value >= 1 && value <= 5) {
								rating = match[1];
								break;
							}
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
		`, &rawListings),
	)

	if err != nil {
		log.Fatal("Error scraping listings:", err)
	}

	// Convert to []Listing
	for _, item := range rawListings {
		listings = append(listings, Listing{
			Title:    item["title"],
			Price:    item["price"],
			Rating:   item["rating"],
			URL:      item["url"],
			Location: "Kuala Lumpur, Malaysia",
		})
	}

	fmt.Printf("Found %d listings\n", len(listings))

	// Get descriptions
	fmt.Println("\nFetching descriptions...")
	for i := range listings {
		if listings[i].URL != "" {
			fmt.Printf("  Description %d/%d...\n", i+1, len(listings))
			listings[i].Description = getDescription(ctx, listings[i].URL)
			time.Sleep(2 * time.Second)
		}
	}


	fmt.Println("\n========== RESULTS ==========")
	for i, l := range listings {
		fmt.Printf("\nListing %d\n", i+1)
		fmt.Println("Title:", l.Title)
		fmt.Println("Price:", l.Price)
		fmt.Println("Rating:", l.Rating)
		fmt.Println("Location:", l.Location)
		fmt.Println("URL:", l.URL)
		fmt.Println("Description:", l.Description)
	}
}

func getDescription(ctx context.Context, url string) string {
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
