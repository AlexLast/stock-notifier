package notifier

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/stretchr/testify/assert"
)

// mockSESClient defines a mock SES
// client to be used for testing
type mockSESClient struct {
	sesiface.SESAPI
	SendEmailReturnValue *ses.SendEmailOutput
	SendEmailReturnError error
}

// mockSNSClient defines a mock SES
// client to be used for testing
type mockSNSClient struct {
	snsiface.SNSAPI
	PublishReturnValue *sns.PublishOutput
	PublishReturnError error
}

// SendEmail mocks the AWS SES SendEmail function
func (m *mockSESClient) SendEmail(*ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	return m.SendEmailReturnValue, m.SendEmailReturnError
}

// Publish mocks the AWS SNS Publish function
func (m *mockSNSClient) Publish(*sns.PublishInput) (*sns.PublishOutput, error) {
	return m.PublishReturnValue, m.PublishReturnError
}

// GetTestContext returns a new context
// that can be used for all unit tests
func GetTestContext() *Context {
	return &Context{
		SES: &mockSESClient{
			SendEmailReturnValue: &ses.SendEmailOutput{},
		},
		SNS: &mockSNSClient{
			PublishReturnValue: &sns.PublishOutput{},
		},
		HTTP: http.DefaultClient,
	}
}

// TestPriceMatch ensures filtering based
// on price returns correctly
func TestPriceMatch(t *testing.T) {
	product := &Product{
		Name:  "Some product",
		Price: 100.99,
	}

	assert.Equal(t, true, product.PriceMatch(Filter{MinPrice: 100, MaxPrice: 101}))
	assert.Equal(t, true, product.PriceMatch(Filter{MinPrice: 100, MaxPrice: 100.99}))
	assert.Equal(t, false, product.PriceMatch(Filter{MinPrice: 50, MaxPrice: 100.98}))
}

// TestDecodeFilter tests the custom envconfig decoder
func TestDecodeFilter(t *testing.T) {
	filter := new(FilterDecoder)
	err := filter.Decode(`[{"term": "test", "minPrice": 100.99, "maxPrice": 200, "interval": 60}]`)

	assert.Nil(t, err)
	assert.Len(t, *filter, 1)

	// We can't index a pointer so iterate
	for _, f := range *filter {
		assert.Equal(t, "test", f.Term)
		assert.Equal(t, float64(100.99), f.MinPrice)
		assert.Equal(t, float64(200), f.MaxPrice)
		assert.Equal(t, int64(60), f.Interval)
	}
}

// TestDecodeNotify tests the custom envconfig decoder
func TestDecodeNotify(t *testing.T) {
	notify := new(Notify)
	err := notify.Decode(`{"email": "test@example.org", "phone": "+123456789"}`)

	assert.Nil(t, err)
	assert.Equal(t, "test@example.org", *notify.Email)
	assert.Equal(t, "+123456789", *notify.Phone)
}

// TestGetPage tests the getPage function
func TestGetPage(t *testing.T) {
	c := GetTestContext()

	// Build a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/test")
		assert.Equal(t, req.Method, "GET")

		// Test response
		rw.Write([]byte(`<p class="t">test</p>`))
	}))

	// Close the server when test finishes
	defer server.Close()

	// Set to our test client
	c.HTTP = server.Client()

	// Get the page
	page, err := c.getPage(fmt.Sprintf("%s/test", server.URL))

	assert.Nil(t, err)
	assert.Equal(t, "test", page.Find("p.t").Text())
}

// TestGetPageBadStatus tests the getPage function
// when a bad HTTP status code is returned
func TestGetPageBadStatus(t *testing.T) {
	c := GetTestContext()

	// Build a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/test")
		assert.Equal(t, req.Method, "GET")

		// Test response
		rw.WriteHeader(403)
		rw.Write([]byte(`<p class="t">test</p>`))
	}))

	// Close the server when test finishes
	defer server.Close()

	// Set to our test client
	c.HTTP = server.Client()

	// Get the page
	_, err := c.getPage(fmt.Sprintf("%s/test", server.URL))

	assert.NotNil(t, err)
}

// TestGetPageNoResponse tests the getPage function
// when no response is returned
func TestGetPageNoResponse(t *testing.T) {
	c := GetTestContext()

	// Get the page
	_, err := c.getPage("http://localhost/test")

	assert.NotNil(t, err)
}

// TestSendNotification tests the
// SendNotification function
func TestSendNotification(t *testing.T) {
	c := GetTestContext()

	// Build some test config
	c.Config = &Config{
		Notify: Notify{
			Email: aws.String("test@example.com"),
			Phone: aws.String("+12345678"),
		},
		FromAddress: "test@example.com",
	}

	// Send the notificatiom
	err := c.SendNotification("test", []Product{{Name: "test", Price: 100}})
	assert.Nil(t, err)

	// Test AWS error is surfaced
	c.SNS = &mockSNSClient{
		PublishReturnError: errors.New("Some AWS error"),
	}

	// Send the notification again
	err = c.SendNotification("test", []Product{{Name: "test", Price: 100}})
	assert.NotNil(t, err)

	// With phone not set the error
	// should no longer be surfaced
	c.Config.Notify.Phone = nil

	err = c.SendNotification("test", []Product{{Name: "test", Price: 100}})
	assert.Nil(t, err)
}
