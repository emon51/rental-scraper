package config

type Config struct {
	BaseURL      string
	MaxListings  int
	PageTimeout  int
	RequestDelay int
	Headless     bool
}

func NewConfig() *Config {
	return &Config{
		BaseURL:      "https://www.airbnb.com/s/Kuala-Lumpur/homes",
		MaxListings:  10,
		PageTimeout:  300,
		RequestDelay: 2,
		Headless:     true,
	}
}