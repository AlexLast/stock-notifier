package notifier

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/jasonlvhit/gocron"
	log "github.com/sirupsen/logrus"
)

// Filter defines the configuration
// for a search filter
type Filter struct {
	Term     string  `json:"term"`
	MinPrice float64 `json:"minPrice"`
	MaxPrice float64 `json:"maxPrice"`
	Interval int64   `json:"interval"`
}

// Product defines the structure
// for any product returned by
// any retailer
type Product struct {
	Name  string
	Price float64
}

// Response defines the structure
// for a response wrapper all retailers
// will return that includes metadata
type Response struct {
	Matches []Product
	Parsed  int
}

// FilterDecoder is a type
// used for an envconfig custom decoder
type FilterDecoder []Filter

// Notify defines the configuration
// for who should be notified
type Notify struct {
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

// NotifyDecoder is a type
// used for an envconfig custom decoder
type NotifyDecoder []Notify

// Config defines the configuration
// for notifier
type Config struct {
	Notify      NotifyDecoder `required:"true"`
	Filters     FilterDecoder `required:"true"`
	CacheTTL    int           `required:"true" split_words:"true"`
	LogLevel    string        `split_words:"true"`
	AWSRegion   string        `required:"true" envconfig:"AWS_REGION"`
	FromAddress string        `required:"true" split_words:"true"`
}

// Context defines the notifier
// context
type Context struct {
	SES    sesiface.SESAPI
	SNS    snsiface.SNSAPI
	HTTP   *http.Client
	Config *Config
}

const (
	smsFromName    = "Stock"
	smsFormat      = "The following products were found on %s: \n\n%s"
	cacheKeyFormat = "%s:%s:%f:%s"
)

// notificationCache is a simple cache
// of sent notifications with a TTL
var notificationCache = map[string]time.Time{}

// Start will start all polling jobs
// for retailers
func (c *Context) Start() {
	log.Infoln("Polling retailers")

	// Start polling for all filters
	// against all retailers
	for _, filter := range c.Config.Filters {
		gocron.Every(uint64(filter.Interval)).Seconds().Do(c.PollRetailer, "Ebuyer", filter)
		gocron.Every(uint64(filter.Interval)).Seconds().Do(c.PollRetailer, "Overclockers", filter)
		gocron.Every(uint64(filter.Interval)).Seconds().Do(c.PollRetailer, "Novatech", filter)
		gocron.Every(uint64(filter.Interval)).Seconds().Do(c.PollRetailer, "Scan", filter)
	}

	<-gocron.Start()
}

// PollRetailer is the wrapper for polling a retailer
// including the sleep interval and notification trigger
func (c *Context) PollRetailer(retailer string, filter Filter) {
	log.Debugf("Polling %s for %s", retailer, filter.Term)

	// Slice of matched products
	var err error
	var response Response

	// Check the retailers for stock
	switch retailer {
	case "Ebuyer":
		response, err = c.CheckEbuyer(filter, &[]Product{}, 1, 1)
	case "Overclockers":
		response, err = c.CheckOverclockers(filter, &[]Product{}, 1, 1)
	case "Novatech":
		response, err = c.CheckNovatech(filter, &[]Product{}, 1, 1)
	case "Scan":
		response, err = c.CheckScan(filter)
	}

	if err != nil {
		log.Errorln(err)
		return
	}

	// Perform generic filtering
	response.Matches = FilterProducts(response.Matches, filter)

	// Log some useful information
	log.Debugf("%s poll for %s parsed %d products, %d matched the filter", retailer, filter.Term, response.Parsed, len(response.Matches))

	// If we matched some products, log them
	for _, product := range response.Matches {
		log.Infof("%s has stock for %s, product: %s", retailer, filter.Term, product.Name)
	}

	// Send notifications
	for _, notify := range c.Config.Notify {
		err = c.SendNotification(retailer, response.Matches, notify)

		if err != nil {
			log.Errorf("Unable to send notification, error: %v", err)
		}
	}
}

// Decode is a custom decoder for filters
// required by envconfig
func (f *FilterDecoder) Decode(value string) error {
	var filters []Filter

	// Unmarshal into filters slice
	err := json.Unmarshal([]byte(value), &filters)

	if err != nil {
		return fmt.Errorf("Invalid filters JSON, error: %v", err)
	}

	// Set the filter value
	*f = filters

	return nil
}

// Decode is a custom decoder for notifies
// required by envconfig
func (n *NotifyDecoder) Decode(value string) error {
	var notifies []Notify

	// Unmarshal into notifies
	err := json.Unmarshal([]byte(value), &notifies)

	if err != nil {
		return fmt.Errorf("Invalid notification JSON, error: %v", err)
	}

	// Set the notifies value
	*n = notifies

	return nil
}

// FilterProducts will return a slice of filtered products
func FilterProducts(p []Product, f Filter) []Product {
	var filtered []Product

	// Ensure all product names directly contain the search filter
	// for better match accuracy, also ensure prices match
	for _, product := range p {
		if strings.Contains(strings.ToLower(product.Name), strings.ToLower(f.Term)) && product.PriceMatch(f) {
			filtered = append(filtered, product)
		}
	}

	return filtered
}

// GetHash returns the hash of a notification
// so it can be cached
func (n Notify) GetHash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v", n))))
}

// getPage returns the decoded HTML ready for parsing
func (c *Context) getPage(url string) (*goquery.Document, error) {
	response, err := c.HTTP.Get(url)

	// We couldn't make the HTTP request
	if err != nil {
		return nil, fmt.Errorf("Unable to load %s, error: %v", url, err)
	}

	defer response.Body.Close()

	// We got a non 200 response
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to load %s, got status code %d", url, response.StatusCode)
	}

	// Decode the response body
	body, err := goquery.NewDocumentFromReader(response.Body)

	if err != nil {
		return nil, fmt.Errorf("Unable to decode body for %s, error: %v", url, err)
	}

	return body, err
}

// BuildSNS returns the SNS publish input
func BuildSNS(message string, phone *string) *sns.PublishInput {
	return &sns.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: phone,
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"AWS.SNS.SMS.SenderID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(smsFromName),
			},
			"AWS.SNS.SMS.SMSType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("Transactional"),
			},
		},
	}
}

// BuildSES returns the SES send email input
func BuildSES(from, message string, email *string) *ses.SendEmailInput {
	return &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{email},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(message),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String("New alert from stock-notifier"),
			},
		},
		Source: aws.String(from),
	}
}

// SendNotification will send notifications
// for the supplied matches if the notification isnt in cache
func (c *Context) SendNotification(retailer string, matches []Product, notify Notify) error {
	var notifications []string

	// Iterate our matches and build the message
	for _, match := range matches {
		// Build our cache key
		key := fmt.Sprintf(cacheKeyFormat, retailer, match.Name, match.Price, notify.GetHash())

		// Check whether we've already
		// sent a notification
		ttl, exists := notificationCache[key]

		// Update the TTL if expired or create new cache entry
		if (exists && time.Since(ttl) > (time.Second*time.Duration(c.Config.CacheTTL))) || !exists {
			notifications = append(notifications, match.Name)
			notificationCache[key] = time.Now()
		}
	}

	// Send notifications if we have some
	// products to alert on
	if len(notifications) > 0 {
		// Build the message
		message := fmt.Sprintf(smsFormat, retailer, strings.Join(notifications, "\n\n"))

		var smsErr error
		var emailErr error

		if notify.Phone != nil {
			// Send the SMS
			_, smsErr = c.SNS.Publish(BuildSNS(message, notify.Phone))
		}

		if notify.Email != nil {
			// Send the email
			_, emailErr = c.SES.SendEmail(BuildSES(c.Config.FromAddress, message, notify.Email))
		}

		// Ensure neither of these channels errored
		for _, err := range []error{smsErr, emailErr} {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// PriceMatch checks whether a products price
// matches a supplied filter
func (p *Product) PriceMatch(filter Filter) bool {
	if p.Price >= filter.MinPrice && p.Price <= filter.MaxPrice {
		return true
	}

	return false
}
