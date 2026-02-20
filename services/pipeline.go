package services

import (
	"context"
	"fmt"

	"github.com/emon51/rental-scraper/config"
	"github.com/emon51/rental-scraper/models"
	"github.com/emon51/rental-scraper/storage"
)

type Pipeline struct {
	cfg *config.Config
}

func NewPipeline(cfg *config.Config) *Pipeline {
	return &Pipeline{cfg: cfg}
}

// Execute runs the complete scraping pipeline
func (p *Pipeline) Execute(ctx context.Context) error {
	// Step 1: Scrape data
	scraperService := NewScraperService(p.cfg)
	cleanedListings, err := scraperService.ScrapeAll(ctx)
	if err != nil {
		return fmt.Errorf("scraping failed: %w", err)
	}

	// Step 2: Save to CSV
	if err := p.saveToCSV(cleanedListings); err != nil {
		return fmt.Errorf("CSV save failed: %w", err)
	}

	// Step 3: Save to PostgreSQL
	if err := p.saveToDatabase(cleanedListings); err != nil {
		return fmt.Errorf("database save failed: %w", err)
	}

	// Step 4: Generate insights
	p.generateInsights(cleanedListings)

	return nil
}

func (p *Pipeline) saveToCSV(listings []models.Listing) error {
	fmt.Println("\n=== STEP 3: SAVING TO CSV ===")

	csvWriter := storage.NewCSVWriter("listings.csv")
	if err := csvWriter.WriteListings(listings); err != nil {
		return err
	}

	fmt.Println("✓ Data saved to listings.csv")
	return nil
}

func (p *Pipeline) saveToDatabase(listings []models.Listing) error {
	fmt.Println("\n=== STEP 4: SAVING TO POSTGRESQL ===")

	pgWriter, err := storage.NewPostgresWriter(
		p.cfg.DBConfig.Host,
		p.cfg.DBConfig.Port,
		p.cfg.DBConfig.User,
		p.cfg.DBConfig.Password,
		p.cfg.DBConfig.DBName,
	)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer pgWriter.Close()

	if err := pgWriter.CreateTable(); err != nil {
		return fmt.Errorf("table creation failed: %w", err)
	}
	fmt.Println("✓ Database table created")

	if err := pgWriter.InsertListings(listings); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	fmt.Printf("✓ %d listings saved to PostgreSQL\n", len(listings))

	return nil
}

func (p *Pipeline) generateInsights(listings []models.Listing) {
	fmt.Println("\n=== STEP 5: GENERATING INSIGHTS ===")

	insightGen := NewInsightGenerator()
	insights := insightGen.Generate(listings)
	insightGen.PrintReport(insights)
}