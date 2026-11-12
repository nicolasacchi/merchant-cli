package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nicolasacchi/merchant-cli/internal/merchant"
	"github.com/spf13/cobra"
)

const projectID = "flawless-mason-482810-p4"

var (
	version         = "dev"
	merchantID      string
	credentialsFile string
	quotaProject    string
)

// newService builds the Merchant API client; tests replace it with a fake.
type service = merchant.Service

var newService = func() service {
	return merchant.New(merchant.Options{CredentialsFile: credentialsFile, QuotaProject: quotaProject})
}

var rootCmd = &cobra.Command{
	Use:           "merchant-cli",
	Short:         "Google Merchant Center CLI",
	Long:          "Command-line tool for Google Merchant Center.\nUses the Merchant API v1 (accounts, products, reports/MCQL).\nAuthentication via service account or Application Default Credentials.\nThe GCP project must be registered once per Merchant account: see 'merchant-cli setup register-gcp'.",
	SilenceUsage:  true,
	SilenceErrors: true,
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

func requireMerchantID() (string, error) {
	mid := getMerchantID()
	if mid == "" {
		return "", errors.New("merchant ID required: use --id flag or MERCHANT_ID env var")
	}
	for _, r := range mid {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("invalid merchant ID %q: must be numeric", mid)
		}
	}
	return mid, nil
}

// WrapAPIError turns a GCP_NOT_REGISTERED failure into an actionable message.
func WrapAPIError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "GCP_NOT_REGISTERED") {
		return fmt.Errorf("%w\n\nThe GCP project used by these credentials is not registered with Merchant account %s.\n"+
			"Register it once with credentials that have ADMIN access on the account:\n"+
			"  merchant-cli setup register-gcp --id %s --yes   (the service account, if it is a Merchant Center admin)\n"+
			"or with an admin's own credentials (--credentials <authorized_user.json>; gcloud's default OAuth client is blocked for this scope, use --client-id-file):\n"+
			"  merchant-cli setup register-gcp --id %s --developer-email <admin@example.com> --quota-project %s --yes\n"+
			"See 'merchant-cli setup register-gcp --help'.", err, orUnknown(getMerchantID()), orUnknown(getMerchantID()), orUnknown(getMerchantID()), projectID)
	}
	return err
}

func orUnknown(s string) string {
	if s == "" {
		return "<account-id>"
	}
	return s
}

func ctx() context.Context { return context.Background() }

func init() {
	rootCmd.PersistentFlags().StringVar(&merchantID, "id", "", "Merchant Center account ID (default: MERCHANT_ID env var)")
	rootCmd.PersistentFlags().StringVar(&credentialsFile, "credentials", "", "Credentials JSON (service account or authorized_user); default GOOGLE_APPLICATION_CREDENTIALS, then ADC")
	rootCmd.PersistentFlags().StringVar(&quotaProject, "quota-project", "", "GCP project to bill/register calls to (needed with user ADC)")
	rootCmd.AddCommand(accountsCmd)
	rootCmd.AddCommand(productsCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(setupCmd)
}
