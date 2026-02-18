package storage

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/emon51/rental-scraper/models"
)

type CSVWriter struct {
	filename string
}

func NewCSVWriter(filename string) *CSVWriter {
	return &CSVWriter{filename: filename}
}

func (w *CSVWriter) WriteListings(listings []models.Listing) error {
	file, err := os.Create(w.filename)
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
			listing.Platform,
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