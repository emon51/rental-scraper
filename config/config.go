package config

type Config struct {
	BaseURL           string
	Locations         []LocationConfig
	ListingsPerPage   int
	PagesToScrape     int
	PageTimeout       int
	RequestDelay      int
	Headless          bool
	MaxConcurrent     int
	DBConfig          DatabaseConfig
}

type LocationConfig struct {
	Slug        string 
	DisplayName string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func NewConfig() *Config {
	return &Config{
		BaseURL:         "https://www.airbnb.com/s/%s/homes",
		ListingsPerPage: 5,
		PagesToScrape:   2, // Page 1 and Page 2
		PageTimeout:     600,
		RequestDelay:    2,
		Headless:        true,
		MaxConcurrent:   3,
		Locations: []LocationConfig{
			{Slug: "Kuala-Lumpur", DisplayName: "Kuala Lumpur, Malaysia"},
			{Slug: "Bangkok", DisplayName: "Bangkok, Thailand"},
			{Slug: "Seoul", DisplayName: "Seoul, South Korea"},
			{Slug: "Tokyo", DisplayName: "Tokyo, Japan"},
			{Slug: "Melbourne", DisplayName: "Melbourne, Australia"},
			{Slug: "Sydney", DisplayName: "Sydney, Australia"},
			{Slug: "Osaka", DisplayName: "Osaka, Japan"},
			{Slug: "Johor-Bahru-District", DisplayName: "Johor Bahru, Malaysia"},
			{Slug: "Busan", DisplayName: "Busan, South Korea"},
		},
		DBConfig: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "rental_scraper",
		},
	}
}