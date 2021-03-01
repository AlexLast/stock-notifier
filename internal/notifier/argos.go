package notifier

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

const (
	argosSleep  = 2
	argosSearch = `https://www.argos.co.uk/finder-api/product;isSearch=true;queryParams={"page":"%d"};searchTerm=%s?returnMeta=true`
)

// argosPageMeta defines the structure for
// the pagination metadata
type argosPageMeta struct {
	PageSize    int `json:"pageSize"`
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
}

// argosProduct defines the structure for a product
// returned by argos
type argosProductAttributes struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Reservable  bool    `json:"reservable"`
	Deliverable bool    `json:"deliverable"`
}

// argosProductWrapper defines the structure of the
// product wrapper
type argosProductWrapper struct {
	Attributes argosProductAttributes `json:"attributes"`
}

// argosResponse defines the structure for the response
// from argos
type argosResponse struct {
	Meta argosPageMeta         `json:"meta"`
	Data []argosProductWrapper `json:"data"`
}

// argosWrapper defines the structure of the wrapper
// around the response from argos
type argosWrapper struct {
	Data struct {
		Response argosResponse `json:"response"`
	}
}

// FetchArgos will fetch results from Argos.co.uk for the specified filter
func (c *Context) FetchArgos(filter Filter, matches *[]Product, cPage, fPage int) (Response, error) {
	response := Response{}
	argosResponse := new(argosWrapper)

	// Get the API response
	url := fmt.Sprintf(argosSearch, cPage, url.QueryEscape(filter.Term))
	raw, err := c.getRaw(url)

	if err != nil {
		return response, err
	}

	// Unmarshal the response
	err = json.Unmarshal(raw, argosResponse)

	if err != nil {
		return response, fmt.Errorf("Unable to unmarshal response for %s, error: %v", url, err)
	}

	// Set the final page
	fPage = argosResponse.Data.Response.Meta.TotalPages

	// Iterate products and append to matches
	for _, product := range argosResponse.Data.Response.Data {
		// Convert to product type
		p := Product{
			Name:    product.Attributes.Name,
			Price:   product.Attributes.Price,
			InStock: product.Attributes.Deliverable,
		}

		// Append to our matches
		*matches = append(*matches, p)
	}

	cPage++

	// If there are further pages we need to recurse
	if cPage <= fPage {
		// Sleep between pages for 2 seconds
		time.Sleep(time.Duration(argosSleep) * time.Second)

		// Call this function recursively
		_, err := c.FetchArgos(filter, matches, cPage, fPage)

		if err != nil {
			return response, err
		}
	}

	response.Matches = *matches

	return response, nil
}
