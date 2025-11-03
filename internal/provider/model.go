package provider

type Scraper interface {
	Scrape() error
}
