package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFetchVery ensures the function is parsing
// products correctly, the retailer is not mocked
func TestFetchVery(t *testing.T) {
	c := GetTestContext()

	// Test filter that should return products
	// we will actually scrape the retailer live
	filter := Filter{
		Term:     "Playstation 5",
		MinPrice: 300,
		MaxPrice: 700,
	}

	// Check the retailer
	response, err := c.FetchVery(filter, &[]Product{}, 1, 1)
	response.Parsed = len(response.Matches)
	response.Matches = FilterProducts(response.Matches, filter)

	assert.Nil(t, err)
	assert.NotEqual(t, 0, response.Parsed)
}
