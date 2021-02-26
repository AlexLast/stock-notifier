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
	overclockersSleep  = 2
	overclockersSearch = "https://www.overclockers.co.uk/search/index/sSearch/%s/sPerPage/48/sPage/%d"
)

// CheckOverclockers will check Overclockers for the specified filter
func (c *Context) CheckOverclockers(filter Filter, matches *[]Product, cPage, fPage int) (Response, error) {
	response := Response{}

	// Get the page contents and our goquery document
	page, err := c.getPage(fmt.Sprintf(overclockersSearch, url.QueryEscape(filter.Term), cPage))

	if err != nil {
		return response, err
	}

	// Get the pagination HTML and determine
	// how many pages we need to parse
	pagination := page.Find("div.display_sites")
	pagination.Find("strong").Each(func(i int, data *goquery.Selection) {
		f, err := strconv.Atoi(data.Text())

		if err == nil {
			fPage = f
		}
	})

	// Get products on the current page and extract
	// the fields we want to filter on
	products := page.Find("div.artbox")
	products.Each(func(i int, data *goquery.Selection) {
		title := data.Find("span.ProductTitle").Text()
		title = strings.ReplaceAll(title, "\n", "")
		title = strings.ReplaceAll(title, `"`, "")

		// Increment parsed count
		response.Parsed++

		// Build our product
		product := Product{
			Name: title,
		}

		// Get the product price
		// we need to use regex to extract the price
		re := regexp.MustCompile("[0-9].+[0-9]")
		price := re.FindString(data.Find("span.price").Text())

		// Convert price to float
		f, err := strconv.ParseFloat(price, 64)

		if err == nil {
			product.Price = f
		}

		// Ensure the product is in-stock
		// and matches our filter and then append to our slice
		if strings.Contains(data.Find("p.deliverable1").Text(), "In stock") {
			*matches = append(*matches, product)
		}
	})

	cPage++

	// If there are further pages we need to recurse
	if cPage <= fPage {
		// Sleep between pages for 2 seconds
		time.Sleep(time.Duration(overclockersSleep) * time.Second)

		// Call this function recursively
		_, err := c.CheckOverclockers(filter, matches, cPage, fPage)

		if err != nil {
			return response, err
		}
	}

	response.Matches = *matches

	return response, nil
}
