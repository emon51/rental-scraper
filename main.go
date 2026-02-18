package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
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
	fmt.Println("Airbnb Scraper Starting...")

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
		`, &listings),
	)

	if err != nil {
		log.Fatal("Error scraping listings:", err)
	}

	fmt.Printf("Found %d listings\n", len(listings))

	// Get descriptions
	fmt.Println("\nFetching descriptions...")
	for i := range listings {
		if listings[i].URL != "" {
			fmt.Printf("  Description %d/%d...\n", i+1, len(listings))
			description := getDescription(ctx, listings[i].URL)
			listings[i].Description = description
			listings[i].Location = "Kuala Lumpur, Malaysia"
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("\nTotal listings scraped: %d\n", len(listings))

	// Save to CSV
	if err := saveToCSV(listings, "listings.csv"); err != nil {
		log.Fatal("Error saving to CSV:", err)
	}

	fmt.Println("Data saved to listings.csv")

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

func saveToCSV(listings []Listing, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Platform", "Title", "Price", "Location", "Rating", "URL", "Description"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, listing := range listings {
		record := []string{
			"Airbnb",
			listing.Title,
			cleanPrice(listing.Price),
			listing.Location,
			listing.Rating,
			listing.URL,
			listing.Description,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func cleanPrice(price string) string {
	price = strings.TrimSpace(price)
	price = strings.ReplaceAll(price, "$", "")
	price = strings.ReplaceAll(price, "RM", "")
	price = strings.ReplaceAll(price, ",", "")

	fields := strings.Fields(price)
	if len(fields) > 0 {
		return fields[0]
	}

	return price
}
