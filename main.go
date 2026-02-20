package main

import (
	"fmt"
	"log"
	"time"

	"github.com/emon51/rental-scraper/config"
	"github.com/emon51/rental-scraper/services"
	"github.com/emon51/rental-scraper/utils"
)

func main() {
	startTime := time.Now()

	fmt.Println("Airbnb Rental Scraper Starting...")

	// Load configuration
	cfg := config.NewConfig()

	// Create browser context
	ctx := utils.CreateBrowserContext(cfg)

	// Create and execute pipeline
	pipeline := services.NewPipeline(cfg)
	if err := pipeline.Execute(ctx); err != nil {
		log.Fatalf("Pipeline failed: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n✓ Scraping Complete! (Duration: %v)\n", duration)
}