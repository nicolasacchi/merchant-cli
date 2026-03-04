# CLAUDE.md — merchant-cli

Go CLI for Google Merchant Center. Single binary, JSON output, service account auth.

**APIs**: Content API v2.1 (accounts, products) + Merchant Reports API v1beta (MCQL queries).

## Authentication

1. `GOOGLE_APPLICATION_CREDENTIALS` env var — path to service account JSON key file
2. Falls back to Application Default Credentials (`gcloud auth application-default login`)

Service account needs Content API scope access on the Merchant Center account.

## Merchant ID

Set via `--id` flag or `MERCHANT_ID` env var. Required by most commands.

## Commands

### accounts

```bash
merchant-cli accounts list                    # List all accessible accounts
merchant-cli accounts get --id 196138351      # Get account details
merchant-cli accounts status --id 196138351   # Account status (issues, warnings)
```

### products

```bash
merchant-cli products list --max-results 10                          # List products (paginated)
merchant-cli products list --max-results 50 --page-token TOKEN       # Next page
merchant-cli products get "online:it:IT:12345"                       # Get specific product
merchant-cli products search "aspirina" --max-results 20             # Search by title/description/brand/offer ID
merchant-cli products status --max-results 50                        # Product approval statuses
```

| Flag | Default | Description |
|------|---------|-------------|
| `--max-results` | 50 | Max results per page |
| `--page-token` | — | Pagination token |

Note: `products search` is client-side filtering (case-insensitive substring match on title, description, offer ID, brand). Limited to first 250 products.

### report

```bash
# Raw MCQL query
merchant-cli report run "SELECT product_view.id, product_view.title FROM ProductView LIMIT 5"

# Performance report (clicks, impressions, CTR)
merchant-cli report performance --start-date 2026-02-16 --end-date 2026-02-22 --limit 20

# Product disapprovals and issues
merchant-cli report issues --limit 50

# Competitive visibility (requires category ID)
merchant-cli report competitive --start-date 2026-02-16 --end-date 2026-02-22 --category-id 12345 --country IT --limit 20
```

| Flag | Default | Commands | Description |
|------|---------|----------|-------------|
| `--start-date` | — | performance, competitive | YYYY-MM-DD (required) |
| `--end-date` | — | performance, competitive | YYYY-MM-DD (required) |
| `--limit` | 50/100 | performance (50), issues (100), competitive (50) | Max rows |
| `--category-id` | — | competitive | Google product category ID (required) |
| `--country` | IT | competitive | Country code |

## MCQL Reference

Use snake_case for table and field names. All fields are prefixed with their view name. String values use double quotes.

### Key Tables & Fields

**product_performance_view**: `offer_id`, `title`, `clicks`, `impressions`, `click_through_rate`, `date`

**product_view**: `id`, `offer_id`, `title`, `aggregated_reporting_context_status`, `item_issues`

**competitive_visibility_competitor_view**: `report_category_id`, `report_country_code`, `traffic_source`, `domain`, `is_your_domain`, `relative_visibility`, `page_overlap_rate`, `higher_position_rate`, `date`

### Example Queries

```sql
SELECT product_performance_view.offer_id, product_performance_view.title, product_performance_view.clicks
FROM ProductPerformanceView
WHERE product_performance_view.date BETWEEN "2026-02-16" AND "2026-02-22"
ORDER BY product_performance_view.clicks DESC
LIMIT 10

SELECT product_view.offer_id, product_view.title, product_view.aggregated_reporting_context_status
FROM ProductView
WHERE product_view.aggregated_reporting_context_status = "NOT_ELIGIBLE_OR_DISAPPROVED"
LIMIT 20
```

## Output

All commands output indented JSON to stdout. Errors go to stderr with exit code 1.

## Build

```bash
make install                          # Install to $GOPATH/bin/merchant-cli
make build                            # Build to ./bin/merchant-cli
go install ./cmd/merchant-cli         # Direct Go install
make test                             # Run tests
```

Requires Go 1.25+.

## Project Structure

```
cmd/merchant-cli/main.go              # Entry point
internal/auth/credentials.go          # Google Cloud credential handling
internal/commands/root.go             # Root cobra command, global flags
internal/commands/accounts.go         # accounts subcommands
internal/commands/products.go         # products subcommands
internal/commands/report.go           # report/MCQL subcommands
internal/output/json.go               # JSON output formatter
```
