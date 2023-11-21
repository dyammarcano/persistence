package scraper

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"strings"
)

func GetData() []string {
	// Create a new collector
	c := colly.NewCollector()

	var paragraphs []string

	// Set up a callback to be called for each HTML element with the "p" tag
	c.OnHTML("p", func(e *colly.HTMLElement) {
		// Extract and print the text content of the paragraph
		paragraphText := strings.TrimSpace(e.Text)
		if paragraphText != "" {
			paragraphs = append(paragraphs, paragraphText)
		}
	})

	// Set up error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	if err := c.Visit("https://meetime.com.br/"); err != nil {
		fmt.Println(err)
	}

	return paragraphs
}
