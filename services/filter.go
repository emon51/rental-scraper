package services

import (
	"fmt"
	"strings"

	"github.com/emon51/rental-scraper/models"
)

type Filter struct{}

func NewFilter() *Filter {
	return &Filter{}
}

// CleanListings removes invalid and duplicate listings
func (f *Filter) CleanListings(listings []models.Listing) []models.Listing {
	cleaned := make([]models.Listing, 0)
	seen := make(map[string]bool)

	for _, listing := range listings {
		// Skip if URL is empty or duplicate
		if listing.URL == "" || seen[listing.URL] {
			continue
		}

		// Skip if title is empty
		if strings.TrimSpace(listing.Title) == "" {
			continue
		}

		// Clean and validate
		listing = f.cleanListing(listing)

		cleaned = append(cleaned, listing)
		seen[listing.URL] = true
	}

	return cleaned
}

func (f *Filter) cleanListing(listing models.Listing) models.Listing {
	// Trim whitespace
	listing.Title = strings.TrimSpace(listing.Title)
	listing.Price = strings.TrimSpace(listing.Price)
	listing.Location = strings.TrimSpace(listing.Location)
	listing.Rating = strings.TrimSpace(listing.Rating)
	listing.Description = strings.TrimSpace(listing.Description)

	// Normalize price (remove any remaining non-numeric characters except decimal)
	listing.Price = normalizePriceString(listing.Price)

	return listing
}

func normalizePriceString(price string) string {
	// Keep only digits and decimal point
	result := ""
	for _, ch := range price {
		if (ch >= '0' && ch <= '9') || ch == '.' {
			result += string(ch)
		}
	}
	return result
}

// FilterByPrice returns listings within price range
func (f *Filter) FilterByPrice(listings []models.Listing, minPrice, maxPrice float64) []models.Listing {
	filtered := make([]models.Listing, 0)

	for _, listing := range listings {
		var price float64
		if _, err := fmt.Sscanf(listing.Price, "%f", &price); err == nil {
			if price >= minPrice && price <= maxPrice {
				filtered = append(filtered, listing)
			}
		}
	}

	return filtered
}

// FilterByRating returns listings with minimum rating
func (f *Filter) FilterByRating(listings []models.Listing, minRating float64) []models.Listing {
	filtered := make([]models.Listing, 0)

	for _, listing := range listings {
		var rating float64
		if _, err := fmt.Sscanf(listing.Rating, "%f", &rating); err == nil {
			if rating >= minRating {
				filtered = append(filtered, listing)
			}
		}
	}

	return filtered
}