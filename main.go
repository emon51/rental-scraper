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

	// Initialize logger
	logger, err := utils.NewLogger("scraper.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Scraper application started")

	// Load configuration
	cfg := config.NewConfig()

	// Create browser context
	ctx := utils.CreateBrowserContext(cfg)

	// Create and execute pipeline
	pipeline := services.NewPipeline(cfg, logger)
	if err := pipeline.Execute(ctx); err != nil {
		logger.Error("Pipeline execution failed", err)
		log.Fatalf("Pipeline failed: %v", err)
	}

	duration := time.Since(startTime)
	logger.LogScrapingSession(0, duration)
	fmt.Printf("\n✓ Scraping Complete! (Duration: %v)\n", duration)
}