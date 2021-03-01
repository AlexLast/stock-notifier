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
	verySleep  = 2
	verySearch = "https://www.very.co.uk/e/q/%s.end?pageNumber=%d&numProducts=99"
)

// FetchVery will fetch results from Very.co.uk for the specified filter
func (c *Context) FetchVery(filter Filter, matches *[]Product, cPage, fPage int) (Response, error) {
	response := Response{}

	// Get the page contents and our goquery document
	page, err := c.getPage(fmt.Sprintf(verySearch, url.QueryEscape(filter.Term), cPage))

	if err != nil {
		return response, err
	}

	// Get the pagination HTML and determine
	// how many pages we need to parse
	pagination := page.Find("div.pagination")
	pagination.Find("li").Each(func(i int, data *goquery.Selection) {
		f, err := strconv.Atoi(strings.TrimSpace(data.Text()))

		if err == nil {
			fPage = f
		}
	})

	// Get products on the current page and extract
	// the fields we want to filter on
	products := page.Find("li.product")
	products.Each(func(i int, data *goquery.Selection) {
		// Build our product
		product := Product{
			Name: strings.TrimSpace(data.Find("span.productBrandDesc").Text()),
		}

		// Get the product price
		// we need to use regex to extract the price
		re := regexp.MustCompile("[0-9].+[0-9]")
		price := re.FindString(data.Find("dd.productPrice").Text())

		// Convert price to float
		f, err := strconv.ParseFloat(price, 64)

		if err == nil {
			product.Price = f
		}

		// Ensure the product is in-stock
		// and matches our filter and then append to our slice
		if data.Find("dd.available").Text() == "In Stock" {
			product.InStock = true
		}

		*matches = append(*matches, product)
	})

	cPage++

	// If there are further pages we need to recurse
	if cPage <= fPage {
		// Sleep between pages for 2 seconds
		time.Sleep(time.Duration(verySleep) * time.Second)

		// Call this function recursively
		_, err := c.FetchVery(filter, matches, cPage, fPage)

		if err != nil {
			return response, err
		}
	}

	// Update the response
	response.Matches = *matches

	return response, nil
}
