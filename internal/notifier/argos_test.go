package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckArgos ensures the function is parsing
// products correctly, the retailer is not mocked
func TestCheckArgos(t *testing.T) {
	c := GetTestContext()

	// Test filter that should return products
	// we will actually scrape the retailer live
	filter := Filter{
		Term:     "Playstation 5",
		MinPrice: 300,
		MaxPrice: 600,
	}

	// Check the retailer
	response, err := c.CheckArgos(filter, &[]Product{}, 1, 1)

	assert.Nil(t, err)
	assert.NotEqual(t, 0, response.Parsed)
}
