package commands

import (
	"fmt"
	"os"
	"strings"

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

The caller needs ADMIN access on the Merchant account: either the service
account itself (if it was added as an admin user in Merchant Center >
Settings > Account access) or an admin's own Google credentials:

  merchant-cli setup register-gcp --id <account-id> --yes          # service account with ADMIN access (simplest)

  # admin user: gcloud's built-in OAuth client is blocked by Google for the
  # content scope, so pass your own Desktop OAuth client (--client-id-file)
  gcloud auth application-default login --client-id-file=<oauth-client.json> \
    --scopes=https://www.googleapis.com/auth/content,https://www.googleapis.com/auth/cloud-platform
  merchant-cli setup register-gcp --id <account-id> \
    --credentials ~/.config/gcloud/application_default_credentials.json \
    --developer-email <admin-email> --quota-project ` + projectID + ` --yes

--quota-project selects the GCP project being registered when the credentials
are a user's. --developer-email is optional: it must be a Google Account (not a
service account) and gets the API_DEVELOPER role (existing users are updated,
others are invited). Registration is complete only once at least one Merchant
Center user has the API developer role. Without --yes the command only prints
what it would do.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mid, err := requireMerchantID()
		if err != nil {
			return err
		}
		project := quotaProject
		if project == "" {
			project = "<unset: the project of the credentials>"
		}
		email := registerEmail
		if email == "" {
			email = "<none>"
		}
		if !registerYes {
			fmt.Fprintf(os.Stderr, "Dry run: would register GCP project %s with accounts/%s (developer email %s). Re-run with --yes to apply.\n", project, mid, email)
			return nil
		}
		resp, err := newService().RegisterGcp(ctx(), mid, registerEmail)
		if err != nil {
			if strings.Contains(err.Error(), "PERMISSION_DENIED_INSUFFICIENT_PERMISSIONS") {
				return fmt.Errorf("%w\n\nThese credentials do not have ADMIN access on accounts/%s. Either give the service account\n"+
					"Admin access in Merchant Center (Settings > People and access) and re-run, or run this command with an\n"+
					"admin's own credentials (--credentials <authorized_user.json>, created with\n"+
					"'gcloud auth application-default login --client-id-file=<oauth-client.json>': gcloud's built-in\n"+
					"OAuth client is blocked for the content scope) and --quota-project %s", err, mid, projectID)
			}
			return err
		}
		return output.PrintProto(resp)
	},
}

func init() {
	setupRegisterCmd.Flags().StringVar(&registerEmail, "developer-email", "", "Optional Google Account email (not a service account) to receive the API developer role")
	setupRegisterCmd.Flags().BoolVar(&registerYes, "yes", false, "Actually perform the registration (otherwise dry run)")
	setupCmd.AddCommand(setupStatusCmd, setupRegisterCmd)
}
