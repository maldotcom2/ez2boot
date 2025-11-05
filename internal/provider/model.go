package provider

type Scraper interface {
	Scrape() error
}

type Manager interface {
	Start() error
	Stop() error
}
