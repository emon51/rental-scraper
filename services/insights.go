package services

import (
	"fmt"
	"sort"
	"strings"

	"github.com/emon51/rental-scraper/models"
)

type Insights struct {
	TotalListings       int
	AveragePrice        float64
	MinPrice            float64
	MaxPrice            float64
	MostExpensive       *models.Listing
	TopRatedListings    []models.Listing
	ListingsByLocation  map[string]int
}

type InsightGenerator struct{}

func NewInsightGenerator() *InsightGenerator {
	return &InsightGenerator{}
}

func (ig *InsightGenerator) Generate(listings []models.Listing) Insights {
	insights := Insights{
		TotalListings:      len(listings),
		ListingsByLocation: make(map[string]int),
		TopRatedListings:   make([]models.Listing, 0),
	}

	if len(listings) == 0 {
		return insights
	}

	var totalPrice float64
	var priceCount int
	var mostExpensive *models.Listing
	maxPrice := 0.0

	// Calculate price statistics
	for i := range listings {
		listing := &listings[i]

		// Parse price
		var price float64
		if _, err := fmt.Sscanf(listing.Price, "%f", &price); err == nil {
			totalPrice += price
			priceCount++

			if insights.MinPrice == 0 || price < insights.MinPrice {
				insights.MinPrice = price
			}

			if price > maxPrice {
				maxPrice = price
				mostExpensive = listing
			}
		}

		// Count by location
		if listing.Location != "" {
			insights.ListingsByLocation[listing.Location]++
		}
	}

	insights.MaxPrice = maxPrice
	insights.MostExpensive = mostExpensive

	if priceCount > 0 {
		insights.AveragePrice = totalPrice / float64(priceCount)
	}

	// Get top 5 rated listings
	insights.TopRatedListings = ig.getTopRated(listings, 5)

	return insights
}

func (ig *InsightGenerator) getTopRated(listings []models.Listing, count int) []models.Listing {
	// Filter listings with ratings
	withRatings := make([]models.Listing, 0)
	for _, listing := range listings {
		if listing.Rating != "" {
			withRatings = append(withRatings, listing)
		}
	}

	// Sort by rating descending
	sort.Slice(withRatings, func(i, j int) bool {
		var ratingI, ratingJ float64
		fmt.Sscanf(withRatings[i].Rating, "%f", &ratingI)
		fmt.Sscanf(withRatings[j].Rating, "%f", &ratingJ)
		return ratingI > ratingJ
	})

	// Return top N
	if len(withRatings) > count {
		return withRatings[:count]
	}
	return withRatings
}

func (ig *InsightGenerator) PrintReport(insights Insights) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("VACATION RENTAL MARKET INSIGHTS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("\nTotal Listings Scraped: %d\n", insights.TotalListings)
	fmt.Printf("Airbnb Listings: %d\n", insights.TotalListings)

	fmt.Printf("\nAverage Price: $%.2f\n", insights.AveragePrice)
	fmt.Printf("Minimum Price: $%.2f\n", insights.MinPrice)
	fmt.Printf("Maximum Price: $%.2f\n", insights.MaxPrice)

	if insights.MostExpensive != nil {
		fmt.Println("\nMost Expensive Property:")
		fmt.Printf("  Title: %s\n", insights.MostExpensive.Title)
		fmt.Printf("  Price: $%s\n", insights.MostExpensive.Price)
		fmt.Printf("  Location: %s\n", insights.MostExpensive.Location)
	}

	fmt.Println("\nListings per Location:")
	for location, count := range insights.ListingsByLocation {
		fmt.Printf("  %s: %d\n", location, count)
	}

	if len(insights.TopRatedListings) > 0 {
		fmt.Println("\nTop 5 Highest Rated Properties:")
		for i, listing := range insights.TopRatedListings {
			fmt.Printf("  %d. %s — %s\n", i+1, listing.Title, listing.Rating)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}