package commands

import (
	"fmt"
	"strings"

	"github.com/nicolasacchi/merchant-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	productsMaxResults int32
	productsPageToken  string
)

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage Merchant Center products",
}

var productsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List processed products",
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		items, next, err := newService().ListProducts(ctx(), "accounts/"+mid, productsMaxResults, productsPageToken)
		if err != nil {
			return WrapAPIError(err)
		}
		return output.PrintProtos("products", items, next)
	},
}

var productsGetCmd = &cobra.Command{
	Use:   "get <product-id>",
	Short: "Get product details",
	Long:  "Get a processed product. <product-id> is the Merchant API product id (contentLanguage~feedLabel~offerId, e.g. it~IT~12345) or a full name accounts/<id>/products/<id>.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		name := args[0]
		if !strings.HasPrefix(name, "accounts/") {
			name = fmt.Sprintf("accounts/%s/products/%s", mid, name)
		}
		resp, err := newService().GetProduct(ctx(), name)
		if err != nil {
			return WrapAPIError(err)
		}
		return output.PrintProto(resp)
	},
}

// mcqlString escapes a value for use inside a single-quoted MCQL literal.
func mcqlString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, `'`, `\'`)
}

func buildSearchQuery(q string, limit int32) string {
	pat := "%" + mcqlString(q) + "%"
	return fmt.Sprintf(`SELECT product_view.id, product_view.offer_id, product_view.title, product_view.brand, `+
		`product_view.aggregated_reporting_context_status FROM product_view `+
		`WHERE product_view.title LIKE '%s' OR product_view.offer_id LIKE '%s' OR product_view.brand LIKE '%s' LIMIT %d`,
		pat, pat, pat, limit)
}

var productsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search products by title, brand, or offer ID (MCQL product_view)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		return runReportQuery(buildSearchQuery(args[0], productsMaxResults), mid, int(productsMaxResults), map[string]any{"query": args[0]})
	},
}

var productsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "List product statuses (productStatus on processed products)",
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		items, next, err := newService().ListProducts(ctx(), "accounts/"+mid, productsMaxResults, productsPageToken)
		if err != nil {
			return WrapAPIError(err)
		}
		type row struct {
			Name   string `json:"name"`
			Offer  string `json:"offerId"`
			Title  string `json:"title"`
			Status any    `json:"productStatus"`
		}
		rows := make([]row, 0, len(items))
		for _, p := range items {
			raw, err := output.ProtoToRaw(p.GetProductStatus())
			if err != nil {
				return err
			}
			rows = append(rows, row{Name: p.GetName(), Offer: p.GetOfferId(), Title: p.GetProductAttributes().GetTitle(), Status: raw})
		}
		return output.PrintJSON(map[string]any{"statuses": rows, "totalResults": len(rows), "nextPageToken": next})
	},
}

func init() {
	for _, c := range []*cobra.Command{productsListCmd, productsSearchCmd, productsStatusCmd} {
		c.Flags().Int32Var(&productsMaxResults, "max-results", 50, "Maximum results to return")
	}
	productsListCmd.Flags().StringVar(&productsPageToken, "page-token", "", "Pagination token")
	productsStatusCmd.Flags().StringVar(&productsPageToken, "page-token", "", "Pagination token")
	productsCmd.AddCommand(productsListCmd, productsGetCmd, productsSearchCmd, productsStatusCmd)
}
