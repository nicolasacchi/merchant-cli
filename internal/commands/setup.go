package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/nicolasacchi/merchant-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	registerEmail string
	registerYes   bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-time Merchant API setup (GCP project registration)",
}

var setupStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the GCP developer registration of the Merchant account (read-only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		resp, err := newService().GetDeveloperRegistration(ctx(), mid)
		if err != nil {
			return WrapAPIError(err)
		}
		return output.PrintProto(resp)
	},
}

var setupRegisterCmd = &cobra.Command{
	Use:   "register-gcp",
	Short: "Register the calling GCP project with the Merchant account (developerRegistration.registerGcp)",
	Long: `Registers the GCP project behind the calling credentials with the Merchant
account. Required once per (project, account) before any Merchant API v1 call
works; until then every call fails with 401 GCP_NOT_REGISTERED.

It must be run with the credentials of a Merchant Center ADMIN user, not the
service account. Typical flow:

  gcloud auth application-default login \
    --scopes=https://www.googleapis.com/auth/content,https://www.googleapis.com/auth/cloud-platform
  merchant-cli setup register-gcp --id <account-id> \
    --developer-email <admin-email> --quota-project ` + projectID + ` --yes

--quota-project selects the GCP project being registered. The developer email
must belong to a Google Account (not a service account); it gets the
API_DEVELOPER role / API notifications. Without --yes the command only prints
what it would do.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		if registerEmail == "" {
			return errors.New("--developer-email is required")
		}
		project := quotaProject
		if project == "" {
			project = "<unset: the project of the credentials>"
		}
		if !registerYes {
			fmt.Fprintf(os.Stderr, "Dry run: would register GCP project %s with accounts/%s (developer email %s). Re-run with --yes to apply.\n", project, mid, registerEmail)
			return nil
		}
		resp, err := newService().RegisterGcp(ctx(), mid, registerEmail)
		if err != nil {
			return err
		}
		return output.PrintProto(resp)
	},
}

func init() {
	setupRegisterCmd.Flags().StringVar(&registerEmail, "developer-email", "", "Google Account email to receive the API developer role (required)")
	setupRegisterCmd.Flags().BoolVar(&registerYes, "yes", false, "Actually perform the registration (otherwise dry run)")
	setupCmd.AddCommand(setupStatusCmd, setupRegisterCmd)
}
