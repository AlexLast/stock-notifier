package notifier

const (
	overclockersSearch = "https://www.overclockers.co.uk/search/index/sSearch/%s/sPerPage/48/sPage/%d"
)

// CheckOverclockers will check Overclockers for the specified filter
func (c *Context) CheckOverclockers(filter Filter) ([]Product, error) {
	return []Product{}, nil
}
