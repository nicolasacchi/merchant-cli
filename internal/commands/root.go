package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	merchantID string
)

var rootCmd = &cobra.Command{
	Use:   "merchant-cli",
	Short: "Google Merchant Center CLI",
	Long:  "Command-line tool for Google Merchant Center.\nUses Content API v2.1 (products, accounts) and Merchant API (reports).\nAuthentication via service account or Application Default Credentials.",
}

func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func getMerchantID() string {
	if merchantID != "" {
		return merchantID
	}
	return os.Getenv("MERCHANT_ID")
}

func init() {
	rootCmd.PersistentFlags().StringVar(&merchantID, "id", "", "Merchant Center account ID (default: MERCHANT_ID env var)")
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(productsCmd)
	rootCmd.AddCommand(reportCmd)
}
