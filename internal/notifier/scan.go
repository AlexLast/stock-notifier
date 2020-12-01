package notifier

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	scanSearch = "https://www.scan.co.uk/search?q=%s"
)

// CheckScan will check Scan for the specified filter
func (c *Context) CheckScan(filter Filter) ([]Product, error) {
	// Get the page contents and our goquery document
	page, err := c.getPage(fmt.Sprintf(scanSearch, url.QueryEscape(filter.Term)))

	if err != nil {
		return nil, err
	}

	// Slice of matches
	var matches []Product

	// Get products on the current page and extract
	// the fields we want to filter on
	products := page.Find("ul.productColumns")
	products.Each(func(i int, column *goquery.Selection) {
		column.Find("li.product").Each(func(i int, data *goquery.Selection) {
			product := Product{
				Name: data.Find("span.description").Text(),
			}

			// Get the product price
			// we need to use regex to extract the price
			re := regexp.MustCompile("[0-9].+[0-9]")
			price := re.FindString(data.Find("span.price").Text())
			price = strings.ReplaceAll(price, ",", "")

			// Convert price to float
			f, err := strconv.ParseFloat(price, 64)

			if err == nil {
				product.Price = f
			}

			// Ensure the product is in-stock
			// and matches our filter and then append to our slice
			if strings.Contains(data.Find("div.buyButton").Text(), "Add To Basket") && product.PriceMatch(filter) {
				matches = append(matches, product)
			}
		})
	})

	return matches, nil
}
