package commands

import (
	"github.com/nicolasacchi/merchant-cli/internal/output"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage Merchant Center accounts",
}

var accountsPageSize int32
var accountsPageToken string

var accountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List accessible Merchant Center accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, next, err := newService().ListAccounts(ctx(), accountsPageSize, accountsPageToken)
		if err != nil {
			return WrapAPIError(err)
		}
		return output.PrintProtos("accounts", items, next)
	},
}

var accountsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get account details",
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		resp, err := newService().GetAccount(ctx(), "accounts/"+mid)
		if err != nil {
			return WrapAPIError(err)
		}
		return output.PrintProto(resp)
	},
}

var accountsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get account issues (Merchant API accountIssues)",
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		items, next, err := newService().ListAccountIssues(ctx(), "accounts/"+mid, accountsPageSize, accountsPageToken)
		if err != nil {
			return WrapAPIError(err)
		}
		return output.PrintProtos("issues", items, next)
	},
}

func init() {
	for _, c := range []*cobra.Command{accountsListCmd, accountsStatusCmd} {
		c.Flags().Int32Var(&accountsPageSize, "max-results", 50, "Maximum results to return")
		c.Flags().StringVar(&accountsPageToken, "page-token", "", "Pagination token")
	}
	accountsCmd.AddCommand(accountsListCmd, accountsGetCmd, accountsStatusCmd)
}
