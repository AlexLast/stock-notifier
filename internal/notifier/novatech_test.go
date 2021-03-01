package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckNovatech ensures the function is parsing
// products correctly, the retailer is not mocked
func TestFetchNovatech(t *testing.T) {
	c := GetTestContext()

	// Test filter that should return products
	// we will actually scrape the retailer live
	filter := Filter{
		Term:     "AMD Ryzen",
		MinPrice: 100,
		MaxPrice: 200,
	}

	// Check the retailer
	response, err := c.FetchNovatech(filter, &[]Product{}, 1, 1)
	response.Parsed = len(response.Matches)
	response.Matches = FilterProducts(response.Matches, filter)

	assert.Nil(t, err)
	assert.NotEqual(t, 0, response.Parsed)
	assert.NotEmpty(t, response.Matches)
}
