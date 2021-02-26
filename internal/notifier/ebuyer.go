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
	ebuyerSleep  = 2
	ebuyerSearch = "https://www.ebuyer.com/search?q=%s&page=%d"
)

// CheckEbuyer will check Ebuyer for the specified filter
func (c *Context) CheckEbuyer(filter Filter, matches *[]Product, cPage, fPage int) (Response, error) {
	response := Response{}

	// Get the page contents and our goquery document
	page, err := c.getPage(fmt.Sprintf(ebuyerSearch, url.QueryEscape(filter.Term), cPage))

	if err != nil {
		return response, err
	}

	// Get the pagination HTML and determine
	// how many pages we need to parse
	pagination := page.Find("ul.pagination")
	pagination.Find("li.pagination__item").Each(func(i int, data *goquery.Selection) {
		f, err := strconv.Atoi(data.Text())

		if err == nil {
			fPage = f
		}
	})

	// Get products on the current page and extract
	// the fields we want to filter on
	products := page.Find("div.listing-product")
	products.Each(func(i int, data *goquery.Selection) {
		// Increment parsed count
		response.Parsed++

		// Build our product
		product := Product{
			Name: data.Find("h3.listing-product-title").Text(),
		}

		// Get the product price
		// we need to use regex to extract the price
		re := regexp.MustCompile("[0-9].+[0-9]")
		price := re.FindString(data.Find("div.inc-vat").Text())
		price = strings.ReplaceAll(price, ",", "")

		// Convert price to float
		f, err := strconv.ParseFloat(price, 64)

		if err == nil {
			product.Price = f
		}

		// Ensure the product is in-stock
		// and matches our filter and then append to our slice
		if data.Find("button").Text() == "Add to Basket" {
			*matches = append(*matches, product)
		}
	})

	cPage++

	// If there are further pages we need to recurse
	if cPage <= fPage {
		// Sleep between pages for 2 seconds
		time.Sleep(time.Duration(ebuyerSleep) * time.Second)

		// Call this function recursively
		_, err := c.CheckEbuyer(filter, matches, cPage, fPage)

		if err != nil {
			return response, err
		}
	}

	response.Matches = *matches

	return response, nil
}
