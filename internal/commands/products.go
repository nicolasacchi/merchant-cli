package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/nicolasacchi/merchant-cli/internal/auth"
	"github.com/nicolasacchi/merchant-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	productsMaxResults int64
	productsPageToken  string
)

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage Merchant Center products",
}

var productsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List products",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		svc, err := auth.NewContentService(ctx)
		if err != nil {
			return err
		}

		mid := getMerchantID()
		if mid == "" {
			return fmt.Errorf("merchant ID required: use --id flag or MERCHANT_ID env var")
		}
		id, err := strconv.ParseUint(mid, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid merchant ID: %w", err)
		}

		call := svc.Products.List(id).MaxResults(productsMaxResults)
		if productsPageToken != "" {
			call = call.PageToken(productsPageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return err
		}

		return output.PrintJSON(map[string]any{
			"products":      resp.Resources,
			"nextPageToken": resp.NextPageToken,
			"totalResults":  len(resp.Resources),
		})
	},
}

var productsGetCmd = &cobra.Command{
	Use:   "get <product-id>",
	Short: "Get product details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		svc, err := auth.NewContentService(ctx)
		if err != nil {
			return err
		}

		mid := getMerchantID()
		if mid == "" {
			return fmt.Errorf("merchant ID required: use --id flag or MERCHANT_ID env var")
		}
		id, err := strconv.ParseUint(mid, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid merchant ID: %w", err)
		}

		resp, err := svc.Products.Get(id, args[0]).Do()
		if err != nil {
			return err
		}

		return output.PrintJSON(resp)
	},
}

var productsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search products by title, description, or offer ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		svc, err := auth.NewContentService(ctx)
		if err != nil {
			return err
		}

		mid := getMerchantID()
		if mid == "" {
			return fmt.Errorf("merchant ID required: use --id flag or MERCHANT_ID env var")
		}
		id, err := strconv.ParseUint(mid, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid merchant ID: %w", err)
		}

		// Content API has no native search — list and filter client-side
		resp, err := svc.Products.List(id).MaxResults(250).Do()
		if err != nil {
			return err
		}

		query := strings.ToLower(args[0])
		var filtered []any
		for _, p := range resp.Resources {
			title := strings.ToLower(p.Title)
			desc := strings.ToLower(p.Description)
			offerID := strings.ToLower(p.OfferId)
			brand := strings.ToLower(p.Brand)

			if strings.Contains(title, query) || strings.Contains(desc, query) ||
				strings.Contains(offerID, query) || strings.Contains(brand, query) {
				filtered = append(filtered, p)
				if int64(len(filtered)) >= productsMaxResults {
					break
				}
			}
		}

		return output.PrintJSON(map[string]any{
			"query":        args[0],
			"products":     filtered,
			"totalResults": len(filtered),
		})
	},
}

var productsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "List product statuses (approval, disapproval reasons)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		svc, err := auth.NewContentService(ctx)
		if err != nil {
			return err
		}

		mid := getMerchantID()
		if mid == "" {
			return fmt.Errorf("merchant ID required: use --id flag or MERCHANT_ID env var")
		}
		id, err := strconv.ParseUint(mid, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid merchant ID: %w", err)
		}

		call := svc.Productstatuses.List(id).MaxResults(productsMaxResults)
		if productsPageToken != "" {
			call = call.PageToken(productsPageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return err
		}

		return output.PrintJSON(map[string]any{
			"statuses":      resp.Resources,
			"nextPageToken": resp.NextPageToken,
			"totalResults":  len(resp.Resources),
		})
	},
}

func init() {
	productsListCmd.Flags().Int64Var(&productsMaxResults, "max-results", 50, "Maximum results to return")
	productsListCmd.Flags().StringVar(&productsPageToken, "page-token", "", "Pagination token")
	productsSearchCmd.Flags().Int64Var(&productsMaxResults, "max-results", 50, "Maximum results to return")
	productsStatusCmd.Flags().Int64Var(&productsMaxResults, "max-results", 50, "Maximum results to return")
	productsStatusCmd.Flags().StringVar(&productsPageToken, "page-token", "", "Pagination token")

	productsCmd.AddCommand(productsListCmd)
	productsCmd.AddCommand(productsGetCmd)
	productsCmd.AddCommand(productsSearchCmd)
	productsCmd.AddCommand(productsStatusCmd)
}
