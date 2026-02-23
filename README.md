# merchant-cli

Command-line tool for Google Merchant Center. Single binary, JSON output, service account auth.

Uses the [Content API v2.1](https://developers.google.com/shopping-content/reference/rest/v2.1) (products, accounts) and [Merchant API Reports](https://developers.google.com/merchant/api/reference/rest) (MCQL queries).

## Install

### From source

```bash
go install github.com/nicolasacchi/merchant-cli/cmd/merchant-cli@latest
```

### From release

Download the binary for your platform from [Releases](https://github.com/nicolasacchi/merchant-cli/releases).

```bash
curl -L https://github.com/nicolasacchi/merchant-cli/releases/latest/download/merchant-cli_Linux_x86_64.tar.gz | tar xz
mv merchant-cli ~/.local/bin/
```

### From source (local)

```bash
git clone https://github.com/nicolasacchi/merchant-cli.git
cd merchant-cli
make install
```

## Authentication

Uses the standard Google Cloud credential chain:

1. `GOOGLE_APPLICATION_CREDENTIALS` environment variable (service account JSON key)
2. Application Default Credentials (`gcloud auth application-default login`)

The service account needs **Content API** scope access on the Merchant Center account.

## Merchant ID

Provide the Merchant Center account ID via:

- `--id` flag on any command
- `MERCHANT_ID` environment variable

## Usage

### Accounts

```bash
# List accessible accounts
merchant-cli accounts list

# Get account details
merchant-cli accounts get --id 196138351

# Account status (issues, warnings)
merchant-cli accounts status --id 196138351
```

### Products

```bash
# List products
merchant-cli products list --max-results 10

# Get a specific product
merchant-cli products get "online:it:IT:12345"

# Search products by title/description/brand
merchant-cli products search "aspirina" --max-results 20

# Product approval statuses
merchant-cli products status --max-results 10
```

### Reports (MCQL)

```bash
# Raw MCQL query
merchant-cli report run "SELECT product_view.id, product_view.title FROM ProductView LIMIT 5"

# Product performance (clicks, impressions, CTR)
merchant-cli report performance --start-date 2026-02-16 --end-date 2026-02-22 --limit 20

# Product issues (disapproved products)
merchant-cli report issues --limit 50

# Competitive visibility
merchant-cli report competitive --start-date 2026-02-16 --end-date 2026-02-22 --limit 20
```

## Output

All commands output JSON to stdout. Example:

```json
{
  "row_count": 3,
  "rows": [
    {"offerId": "12345", "title": "Aspirina 500mg", "clicks": "150", "impressions": "2000"}
  ]
}
```

## License

MIT
