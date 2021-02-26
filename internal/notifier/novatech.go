package notifier

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	novatechSleep  = 2
	novatechSearch = "https://www.novatech.co.uk/search.html?search=%s&pg=%d&i=200"
)

// CheckNovatech will check Novatech for the specified filter
func (c *Context) CheckNovatech(filter Filter, matches *[]Product, cPage, fPage int) (Response, error) {
	response := Response{}

	// Get the page contents and our goquery document
	page, err := c.getPage(fmt.Sprintf(novatechSearch, url.QueryEscape(filter.Term), cPage))

	if err != nil {
		return response, err
	}

	// Get the pagination HTML and determine
	// how many pages we need to parse
	pages := regexp.MustCompile("([0-9]+) Pages")
	count := pages.FindString(page.Find("div.results").Text())
	count = strings.ReplaceAll(count, " Pages", "")

	// Convert to int, failure will set
	// to 0, which is fine
	fPage, _ = strconv.Atoi(count)

	// Get products on the current page and extract
	// the fields we want to filter on
	products := page.Find("div.search-box-results")
	products.Each(func(i int, data *goquery.Selection) {
		title := data.Find("div.search-box-title").Text()
		title = strings.ReplaceAll(title, "\n", "")

		// Increment parsed count
		response.Parsed++

		// Build our product
		product := Product{
			Name: title,
		}

		// Get the product price
		// we need to use regex to extract the price
		re := regexp.MustCompile("[0-9].+[0-9]")
		price := re.FindString(data.Find("p.newspec-price").Text())

		// Convert price to float
		f, err := strconv.ParseFloat(price, 64)

		if err == nil {
			product.Price = f
		}

		// Ensure the product is in-stock
		// and matches our filter and then append to our slice
		if strings.Contains(data.Find("a.basket-button").Text(), "View Product") {
			*matches = append(*matches, product)
		}
	})

	cPage++

	// If there are further pages we need to recurse
	if cPage <= fPage {
		// Sleep between pages for 2 seconds
		time.Sleep(time.Duration(novatechSleep) * time.Second)

		// Call this function recursively
		_, err := c.CheckNovatech(filter, matches, cPage, fPage)

		if err != nil {
			return response, err
		}
	}

	response.Matches = *matches

	return response, nil
}
