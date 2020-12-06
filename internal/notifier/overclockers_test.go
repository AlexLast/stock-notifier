package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckOverclockers ensures the function is parsing
// products correctly, the retailer is not mocked
func TestCheckOverclockers(t *testing.T) {
	c := GetTestContext()

	// Test filter that should return products
	// we will actually scrape the retailer live
	filter := Filter{
		Term:     "RTX 3070",
		MinPrice: 500,
		MaxPrice: 700,
	}

	// Check the retailer
	response, err := c.CheckOverclockers(filter, &[]Product{}, 1, 1)

	assert.Nil(t, err)
	assert.NotEqual(t, 0, response.Parsed)
}
