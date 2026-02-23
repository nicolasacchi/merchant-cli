package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nicolasacchi/merchant-cli/internal/auth"
	"github.com/nicolasacchi/merchant-cli/internal/output"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage Merchant Center accounts",
}

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accessible Merchant Center accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		svc, err := auth.NewContentService(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Accounts.Authinfo().Do()
		if err != nil {
			return err
		}

		return output.PrintJSON(resp)
	},
}

var accountsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get account details",
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

		resp, err := svc.Accounts.Get(id, id).Do()
		if err != nil {
			return err
		}

		return output.PrintJSON(resp)
	},
}

var accountsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get account status (issues, warnings)",
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

		resp, err := svc.Accountstatuses.Get(id, id).Do()
		if err != nil {
			return err
		}

		return output.PrintJSON(resp)
	},
}

func init() {
	accountsCmd.AddCommand(accountsListCmd)
	accountsCmd.AddCommand(accountsGetCmd)
	accountsCmd.AddCommand(accountsStatusCmd)
}
