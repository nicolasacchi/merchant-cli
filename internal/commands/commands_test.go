package commands

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/shopping/merchant/accounts/apiv1/accountspb"
	"cloud.google.com/go/shopping/merchant/products/apiv1/productspb"
	"cloud.google.com/go/shopping/merchant/reports/apiv1/reportspb"
)

type fakeService struct {
	err        error
	gotParent  string
	gotName    string
	gotQuery   string
	gotLimit   int
	gotAccount string
	gotEmail   string
	registered int
}

func (f *fakeService) ListAccounts(context.Context, int32, string) ([]*accountspb.Account, string, error) {
	return []*accountspb.Account{{Name: "accounts/1", AccountName: "Shop"}}, "tok", f.err
}
func (f *fakeService) GetAccount(_ context.Context, n string) (*accountspb.Account, error) {
	f.gotName = n
	return &accountspb.Account{Name: n}, f.err
}
func (f *fakeService) ListAccountIssues(_ context.Context, p string, _ int32, _ string) ([]*accountspb.AccountIssue, string, error) {
	f.gotParent = p
	return []*accountspb.AccountIssue{{Title: "Missing policy"}}, "", f.err
}
func (f *fakeService) ListProducts(_ context.Context, p string, _ int32, _ string) ([]*productspb.Product, string, error) {
	f.gotParent = p
	return []*productspb.Product{{Name: p + "/products/it~IT~1", OfferId: "1"}}, "", f.err
}
func (f *fakeService) GetProduct(_ context.Context, n string) (*productspb.Product, error) {
	f.gotName = n
	return &productspb.Product{Name: n}, f.err
}
func (f *fakeService) SearchReport(_ context.Context, p, q string, l int) ([]*reportspb.ReportRow, error) {
	f.gotParent, f.gotQuery, f.gotLimit = p, q, l
	return []*reportspb.ReportRow{{}}, f.err
}
func (f *fakeService) RegisterGcp(_ context.Context, a, e string) (*accountspb.DeveloperRegistration, error) {
	f.gotAccount, f.gotEmail = a, e
	f.registered++
	return &accountspb.DeveloperRegistration{Name: "accounts/" + a + "/developerRegistration"}, f.err
}
func (f *fakeService) GetDeveloperRegistration(_ context.Context, a string) (*accountspb.DeveloperRegistration, error) {
	f.gotAccount = a
	return &accountspb.DeveloperRegistration{Name: "accounts/" + a + "/developerRegistration"}, f.err
}

func setup(t *testing.T) *fakeService {
	t.Helper()
	f := &fakeService{}
	orig, origID := newService, merchantID
	newService = func() service { return f }
	merchantID = "196138351"
	t.Cleanup(func() { newService, merchantID, registerYes, registerEmail, quotaProject = orig, origID, false, "", "" })
	return f
}

func capture(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := fn()
	_ = w.Close()
	os.Stdout = old
	b, _ := io.ReadAll(r)
	return string(b), err
}

func TestAccountsStatusUsesAccountIssues(t *testing.T) {
	f := setup(t)
	out, err := capture(t, func() error { return accountsStatusCmd.RunE(accountsStatusCmd, nil) })
	if err != nil || f.gotParent != "accounts/196138351" || !strings.Contains(out, "Missing policy") {
		t.Fatalf("err=%v parent=%q out=%s", err, f.gotParent, out)
	}
}

func TestAccountsGet(t *testing.T) {
	f := setup(t)
	if _, err := capture(t, func() error { return accountsGetCmd.RunE(accountsGetCmd, nil) }); err != nil || f.gotName != "accounts/196138351" {
		t.Fatalf("err=%v name=%q", err, f.gotName)
	}
}

func TestInvalidMerchantID(t *testing.T) {
	setup(t)
	merchantID = "abc"
	if err := accountsGetCmd.RunE(accountsGetCmd, nil); err == nil {
		t.Fatal("want error for non-numeric id")
	}
	merchantID = ""
	t.Setenv("MERCHANT_ID", "")
	if err := accountsGetCmd.RunE(accountsGetCmd, nil); err == nil {
		t.Fatal("want error for missing id")
	}
}

func TestProductsGetBuildsName(t *testing.T) {
	f := setup(t)
	if _, err := capture(t, func() error { return productsGetCmd.RunE(productsGetCmd, []string{"it~IT~42"}) }); err != nil {
		t.Fatal(err)
	}
	if f.gotName != "accounts/196138351/products/it~IT~42" {
		t.Fatalf("name=%q", f.gotName)
	}
	if _, err := capture(t, func() error { return productsGetCmd.RunE(productsGetCmd, []string{"accounts/9/products/x"}) }); err != nil || f.gotName != "accounts/9/products/x" {
		t.Fatalf("full name passthrough failed: %q %v", f.gotName, err)
	}
}

func TestProductsStatus(t *testing.T) {
	setup(t)
	out, err := capture(t, func() error { return productsStatusCmd.RunE(productsStatusCmd, nil) })
	if err != nil || !strings.Contains(out, `"offerId": "1"`) {
		t.Fatalf("err=%v out=%s", err, out)
	}
}

func TestProductsSearchUsesMCQLAndEscapes(t *testing.T) {
	f := setup(t)
	productsMaxResults = 7
	if _, err := capture(t, func() error { return productsSearchCmd.RunE(productsSearchCmd, []string{`o'brien`}) }); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.gotQuery, "FROM product_view") || !strings.Contains(f.gotQuery, `%o\'brien%`) || f.gotLimit != 7 || f.gotParent != "accounts/196138351" {
		t.Fatalf("query=%q limit=%d parent=%q", f.gotQuery, f.gotLimit, f.gotParent)
	}
}

func TestGCPNotRegisteredIsActionable(t *testing.T) {
	f := setup(t)
	f.err = errors.New("googleapi: Error 401: GCP project is not registered, reason: GCP_NOT_REGISTERED")
	for name, fn := range map[string]func() error{
		"accounts": func() error { return accountsListCmd.RunE(accountsListCmd, nil) },
		"products": func() error { return productsListCmd.RunE(productsListCmd, nil) },
		"report":   func() error { return runReport("SELECT 1") },
		"status":   func() error { return setupStatusCmd.RunE(setupStatusCmd, nil) },
	} {
		_, err := capture(t, fn)
		if err == nil || !strings.Contains(err.Error(), "merchant-cli setup register-gcp --id 196138351") ||
			!errors.Is(err, f.err) {
			t.Errorf("%s: not actionable: %v", name, err)
		}
	}
	// Other errors are not decorated.
	f.err = errors.New("boom")
	if err := WrapAPIError(f.err); strings.Contains(err.Error(), "register-gcp") {
		t.Error("unrelated error decorated")
	}
}

func TestRegisterGcpDryRunThenYes(t *testing.T) {
	f := setup(t)
	registerEmail = "admin@example.com"
	if err := setupRegisterCmd.RunE(setupRegisterCmd, nil); err != nil || f.registered != 0 {
		t.Fatalf("dry run must not call API: err=%v n=%d", err, f.registered)
	}
	registerYes = true
	out, err := capture(t, func() error { return setupRegisterCmd.RunE(setupRegisterCmd, nil) })
	if err != nil || f.registered != 1 || f.gotAccount != "196138351" || f.gotEmail != "admin@example.com" || !strings.Contains(out, "developerRegistration") {
		t.Fatalf("err=%v n=%d acct=%q out=%s", err, f.registered, f.gotAccount, out)
	}
	registerEmail = ""
	if _, err := capture(t, func() error { return setupRegisterCmd.RunE(setupRegisterCmd, nil) }); err != nil || f.registered != 2 || f.gotEmail != "" {
		t.Fatalf("developer email is optional: err=%v n=%d email=%q", err, f.registered, f.gotEmail)
	}
}

func TestRegisterGcpPermissionHint(t *testing.T) {
	f := setup(t)
	registerYes = true
	f.err = errors.New("googleapi: Error 403: ... REASON:PERMISSION_DENIED_INSUFFICIENT_PERMISSIONS")
	err := setupRegisterCmd.RunE(setupRegisterCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "do not have ADMIN access on accounts/196138351") {
		t.Fatalf("want ADMIN hint, got %v", err)
	}
}
