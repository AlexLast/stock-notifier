# stock-notifier
It's currently near impossible to purchase any next generation graphics cards or consoles in the UK.

This tool continually polls supported retailers for the provided filters and will alert when a product becomes in stock.

Supported retailers:
- Ebuyer.com
- Overclockers.co.uk
- Novatech.co.uk
- Scan.co.uk
- Argos.co.uk
- Very.co.uk
- Currys.co.uk

Supported notification channels:
- SMS via AWS SNS
- Email via AWS SES

Example filters configuration:

```json
[
    {
        "term": "RTX 3070", 
        "interval": 60, 
        "minPrice": 500, 
        "maxPrice": 650
    }, 
    {
        "term": "RTX 3080", 
        "interval": 60, 
        "minPrice": 500, 
        "maxPrice": 800
    }, 
    {
        "term": "Playstation 5",
        "interval": 60, 
        "minPrice": 400,
        "maxPrice": 600
    }
]
```

The `stock-notifier` tool is distributed via a docker image, you can use the latest build at `public.ecr.aws/alexlast/stock-notifier:latest` or pick a specific tag from the releases tab of this repository.

## Testing
Unit tests for retailers hit live endpoints to ensure tests only pass when the actual page spec is correct. Mocking these retailer endpoints would prove more reliable, but less accurate.

```bash
$ make test
```

## Future improvements
- Include direct link to product in notifications
