# stock-notifier
It's currently near impossible to purchase any next generation graphics cards in the UK, whether Nvidia or AMD.

This tool continually polls supported retailers for the provided filters and will alert when a product becomes in stock.

Supported retailers:
- Ebuyer.com
- Overclockers.co.uk
- Novatech.co.uk
- Scan.co.uk

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
        "term": "RX 6800", 
        "interval": 60, 
        "minPrice": 500, 
        "maxPrice": 800
    }
]
```

To do:
- Add more UK retailers
- Include direct link to product in notifications
