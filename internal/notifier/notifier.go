package notifier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
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

// FilterDecoder is a type
// used for an envconfig custom decoder
type FilterDecoder []Filter

// Notify defines the configuration
// for who should be notified
type Notify struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// Config defines the configuration
// for notifier
type Config struct {
	Notify    Notify        `required:"true"`
	LogLevel  string        `split_words:"true"`
	AWSRegion string        `required:"true" envconfig:"AWS_REGION"`
	Filters   FilterDecoder `required:"true"`
}

// Context defines the notifier
// context
type Context struct {
	SNS    snsiface.SNSAPI
	HTTP   *http.Client
	Config *Config
}

const (
	smsFromName     = "Stock"
	smsFormat       = "The following products were found on %s: \n\n%s"
	cacheKeyFormat  = "%s:%s:%f"
	cacheTimeToLive = int64(86400)
)

// notificationCache is a simple cache
// of sent notifications with a TTL
var notificationCache = map[string]time.Time{}

// Start will start all polling jobs
// for retailers
func (c *Context) Start() {
	log.Infoln("Polling retailers")

	// Dummy channel to block on
	finished := make(chan bool)

	// Start polling for all filters
	for _, filter := range c.Config.Filters {
		go c.PollRetailer("Ebuyer", filter)
		go c.PollRetailer("Overclockers", filter)
	}

	// Block main thread waiting for our channel
	// that will never get a message
	<-finished
}

// PollRetailer is the wrapper for polling a retailer
// including the sleep interval and notification trigger
func (c *Context) PollRetailer(retailer string, filter Filter) {
	for {
		log.Debugf("Polling %s for %s", retailer, filter.Term)

		// Slice of matched products
		var err error
		var matched []Product

		// Check the retailers for stock
		switch retailer {
		case "Ebuyer":
			matched, err = c.CheckEbuyer(filter)
		case "Overclockers":
			matched, err = c.CheckOverclockers(filter)
		}

		if err != nil {
			log.Errorln(err)
		}

		// If we matched some products, log them
		for _, product := range matched {
			log.Infof("%s has stock for %s, product: %s", retailer, filter.Term, product.Name)
		}

		// Send notifications
		err = c.SendNotification(retailer, matched)

		if err != nil {
			log.Errorf("Unable to send notification, error: %v", err)
		}

		// Sleep for the filters interval
		time.Sleep((time.Duration(filter.Interval) * time.Second))
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

// Decode is a custom decoder for filters
// required by envconfig
func (n *Notify) Decode(value string) error {
	notify := new(Notify)

	// Unmarshal into notify
	err := json.Unmarshal([]byte(value), notify)

	if err != nil {
		return fmt.Errorf("Invalid notify JSON, error: %v", err)
	}

	// Set the notify value
	*n = *notify

	return nil
}

// getPage returns the decoded HTML ready for parsing
func (c *Context) getPage(url string) (*goquery.Document, error) {
	response, err := c.HTTP.Get(url)

	// We couldn't make the HTTP request
	if err != nil {
		return nil, fmt.Errorf("Unable to load page, error: %v", err)
	}

	defer response.Body.Close()

	// We got a non 200 response
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to load page, got status code %d", response.StatusCode)
	}

	// Decode the response body
	body, err := goquery.NewDocumentFromReader(response.Body)

	if err != nil {
		return nil, fmt.Errorf("Unable to decode body, error: %v", err)
	}

	return body, err
}

// SendNotification will send notifications
// for the supplied matches if the notification isnt in cache
func (c *Context) SendNotification(retailer string, matches []Product) error {
	var notifications []string

	// Iterate our matches and build the message
	for _, match := range matches {
		// Build our cache key
		key := fmt.Sprintf(cacheKeyFormat, retailer, match.Name, match.Price)

		// Check whether we've already
		// sent a notification
		ttl, exists := notificationCache[key]

		if !exists {
			notifications = append(notifications, match.Name)
			notificationCache[key] = time.Now()
		} else {
			// If the key is older than the default TTL remove it
			if time.Since(ttl) > (time.Second * time.Duration(cacheTimeToLive)) {
				delete(notificationCache, key)
			}
		}
	}

	// Send notifications if we have some
	// products to alert on
	if len(notifications) > 0 {
		// Build the message
		message := fmt.Sprintf(smsFormat, retailer, strings.Join(notifications, "\n\n"))

		// Send the SMS message
		_, err := c.SNS.Publish(&sns.PublishInput{
			Message:     aws.String(message),
			PhoneNumber: aws.String(c.Config.Notify.Phone),
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
		})

		return err
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
