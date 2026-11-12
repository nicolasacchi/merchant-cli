# merchant-cli

Command-line tool for Google Merchant Center. Single binary, JSON output, service account auth.

Uses the [Merchant API v1](https://developers.google.com/merchant/api/reference/rest) throughout (accounts, products, reports/MCQL). The Content API for Shopping v2.1 was sunset on 2026-08-18 and is no longer used.

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
make install   # injects the git tag as the version
```

## Authentication

Uses the standard Google Cloud credential chain:

1. `GOOGLE_APPLICATION_CREDENTIALS` environment variable (service account JSON key)
2. Application Default Credentials (`gcloud auth application-default login`)

Flags `--credentials <file>` (service account or authorized_user JSON) and `--quota-project <gcp-project>` override the environment.

### One-time GCP registration (required)

The Merchant API v1 rejects every call with `401 GCP_NOT_REGISTERED` until the GCP project behind the credentials is registered with the Merchant account. A Merchant Center **admin** (not the service account) runs once:

```bash
gcloud auth application-default login \
  --scopes=https://www.googleapis.com/auth/content,https://www.googleapis.com/auth/cloud-platform
merchant-cli setup register-gcp --id 196138351 \
  --developer-email <admin-email> --quota-project flawless-mason-482810-p4 --yes
merchant-cli setup status --id 196138351   # read-only check
```

Without `--yes` the command is a dry run. The service account must additionally be added as a user of the Merchant account (Merchant Center > Settings > Account access).

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

# Account issues (Merchant API accountIssues; replaces the old accountstatuses)
merchant-cli accounts status --id 196138351
```

### Products

```bash
# List products
merchant-cli products list --max-results 10

# Get a specific product
merchant-cli products get "it~IT~12345"   # contentLanguage~feedLabel~offerId

# Search products by title/offer id/brand (MCQL product_view, server-side)
merchant-cli products search "aspirina" --max-results 20

# Product approval statuses
merchant-cli products status --max-results 10
```

### Reports (MCQL)

```bash
# Raw MCQL query
merchant-cli report run "SELECT product_view.id, product_view.title FROM product_view LIMIT 5"

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
