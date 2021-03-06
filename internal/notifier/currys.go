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
	currysSleep  = 2
	currysSearch = "https://www.currys.co.uk/gbuk/search-keywords/xx_xx_xx_xx_xx/%s/%d_50/relevance-desc/xx-criteria.html"
)

// FetchCurrys will fetch results from Currys.co.uk for the specified filter
func (c *Context) FetchCurrys(filter Filter, matches *[]Product, cPage, fPage int) (Response, error) {
	response := Response{}

	// Get the page contents and our goquery document
	page, err := c.getPage(fmt.Sprintf(currysSearch, url.QueryEscape(filter.Term), cPage))

	if err != nil {
		return response, err
	}

	// Get the pagination HTML and determine
	// how many pages we need to parse
	pagination := page.Find("ul.pagination")
	pagination.Find("li").Each(func(i int, data *goquery.Selection) {
		f, err := strconv.Atoi(strings.TrimSpace(data.Text()))

		if err == nil {
			fPage = f
		}
	})

	// Get products on the current page and extract
	// the fields we want to filter on
	products := page.Find("article.product")
	products.Each(func(i int, data *goquery.Selection) {
		// Build our product
		product := Product{
			Name: strings.TrimSpace(data.Find(`[data-product="name"]`).Text()),
		}

		// Get the product price
		// we need to use regex to extract the price
		re := regexp.MustCompile("[0-9].+[0-9]")
		price := re.FindString(data.Find("strong.price").Text())

		// Convert price to float
		f, err := strconv.ParseFloat(price, 64)

		if err == nil {
			product.Price = f
		}

		// Ensure the product is in-stock
		// and matches our filter and then append to our slice
		if strings.TrimSpace(data.Find(`[data-availability="homeDeliveryAvailable"]`).Text()) == "FREE delivery available" {
			product.InStock = true
		}

		*matches = append(*matches, product)
	})

	cPage++

	// If there are further pages we need to recurse
	if cPage <= fPage {
		// Sleep between pages for 2 seconds
		time.Sleep(time.Duration(currysSleep) * time.Second)

		// Call this function recursively
		_, err := c.FetchCurrys(filter, matches, cPage, fPage)

		if err != nil {
			return response, err
		}
	}

	// Update the response
	response.Matches = *matches

	return response, nil
}
