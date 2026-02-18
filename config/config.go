package config

type Config struct {
	BaseURL      string
	MaxListings  int
	PageTimeout  int
	RequestDelay int
	Headless     bool
	DBConfig     DatabaseConfig
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
		BaseURL:      "https://www.airbnb.com/s/Kuala-Lumpur/homes",
		MaxListings:  10,
		PageTimeout:  300,
		RequestDelay: 2,
		Headless:     true,
		DBConfig: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "rental_scraper",
		},
	}
}