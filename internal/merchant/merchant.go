// Package merchant wraps the Merchant API v1 clients behind a small interface
// so commands can be tested with a fake.
package merchant

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	accounts "cloud.google.com/go/shopping/merchant/accounts/apiv1"
	"cloud.google.com/go/shopping/merchant/accounts/apiv1/accountspb"
	products "cloud.google.com/go/shopping/merchant/products/apiv1"
	"cloud.google.com/go/shopping/merchant/products/apiv1/productspb"
	reports "cloud.google.com/go/shopping/merchant/reports/apiv1"
	"cloud.google.com/go/shopping/merchant/reports/apiv1/reportspb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Service is the Merchant API v1 surface used by the CLI.
type Service interface {
	ListAccounts(ctx context.Context, pageSize int32, pageToken string) ([]*accountspb.Account, string, error)
	GetAccount(ctx context.Context, name string) (*accountspb.Account, error)
	ListAccountIssues(ctx context.Context, parent string, pageSize int32, pageToken string) ([]*accountspb.AccountIssue, string, error)
	ListProducts(ctx context.Context, parent string, pageSize int32, pageToken string) ([]*productspb.Product, string, error)
	GetProduct(ctx context.Context, name string) (*productspb.Product, error)
	SearchReport(ctx context.Context, parent, query string, limit int) ([]*reportspb.ReportRow, error)
	RegisterGcp(ctx context.Context, account, developerEmail string) (*accountspb.DeveloperRegistration, error)
	GetDeveloperRegistration(ctx context.Context, account string) (*accountspb.DeveloperRegistration, error)
}

// Options controls how credentials are resolved.
type Options struct {
	// CredentialsFile overrides GOOGLE_APPLICATION_CREDENTIALS (service account
	// key or authorized-user JSON).
	CredentialsFile string
	// QuotaProject is the GCP project billed/registered for the calls.
	QuotaProject string
}

func (o Options) clientOptions() []option.ClientOption {
	var opts []option.ClientOption
	path := o.CredentialsFile
	if path == "" {
		path = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}
	if path != "" {
		opts = append(opts, option.WithAuthCredentialsFile(credType(path), path))
	}
	if o.QuotaProject != "" {
		opts = append(opts, option.WithQuotaProject(o.QuotaProject))
	}
	return opts
}

// Client is the production Service.
type Client struct {
	opts []option.ClientOption
}

// New returns a Client that lazily creates REST clients per call.
func New(o Options) *Client { return &Client{opts: o.clientOptions()} }

func (c *Client) ListAccounts(ctx context.Context, pageSize int32, pageToken string) ([]*accountspb.Account, string, error) {
	cl, err := accounts.NewRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, "", err
	}
	defer cl.Close()
	it := cl.ListAccounts(ctx, &accountspb.ListAccountsRequest{PageSize: pageSize, PageToken: pageToken})
	return drain(it, int(pageSize))
}

func (c *Client) GetAccount(ctx context.Context, name string) (*accountspb.Account, error) {
	cl, err := accounts.NewRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	return cl.GetAccount(ctx, &accountspb.GetAccountRequest{Name: name})
}

func (c *Client) ListAccountIssues(ctx context.Context, parent string, pageSize int32, pageToken string) ([]*accountspb.AccountIssue, string, error) {
	cl, err := accounts.NewAccountIssueRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, "", err
	}
	defer cl.Close()
	it := cl.ListAccountIssues(ctx, &accountspb.ListAccountIssuesRequest{Parent: parent, PageSize: pageSize, PageToken: pageToken})
	return drain(it, int(pageSize))
}

func (c *Client) ListProducts(ctx context.Context, parent string, pageSize int32, pageToken string) ([]*productspb.Product, string, error) {
	cl, err := products.NewRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, "", err
	}
	defer cl.Close()
	it := cl.ListProducts(ctx, &productspb.ListProductsRequest{Parent: parent, PageSize: pageSize, PageToken: pageToken})
	return drain(it, int(pageSize))
}

func (c *Client) GetProduct(ctx context.Context, name string) (*productspb.Product, error) {
	cl, err := products.NewRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	return cl.GetProduct(ctx, &productspb.GetProductRequest{Name: name})
}

func (c *Client) SearchReport(ctx context.Context, parent, query string, limit int) ([]*reportspb.ReportRow, error) {
	cl, err := reports.NewReportRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	it := cl.Search(ctx, &reportspb.SearchRequest{Parent: parent, Query: query})
	var rows []*reportspb.ReportRow
	for {
		row, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
		if limit > 0 && len(rows) >= limit {
			break
		}
	}
	return rows, nil
}

func (c *Client) RegisterGcp(ctx context.Context, account, developerEmail string) (*accountspb.DeveloperRegistration, error) {
	cl, err := accounts.NewDeveloperRegistrationRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	return cl.RegisterGcp(ctx, &accountspb.RegisterGcpRequest{
		Name:           fmt.Sprintf("accounts/%s/developerRegistration", account),
		DeveloperEmail: developerEmail,
	})
}

func (c *Client) GetDeveloperRegistration(ctx context.Context, account string) (*accountspb.DeveloperRegistration, error) {
	cl, err := accounts.NewDeveloperRegistrationRESTClient(ctx, c.opts...)
	if err != nil {
		return nil, err
	}
	defer cl.Close()
	return cl.GetDeveloperRegistration(ctx, &accountspb.GetDeveloperRegistrationRequest{
		Name: fmt.Sprintf("accounts/%s/developerRegistration", account),
	})
}

type pageIterator[T any] interface {
	Next() (T, error)
	PageInfo() *iterator.PageInfo
}

// drain reads at most one page (up to limit items) and returns the next page token.
func drain[T any](it pageIterator[T], limit int) ([]T, string, error) {
	var out []T
	for {
		v, err := it.Next()
		if errors.Is(err, iterator.Done) {
			return out, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		out = append(out, v)
		if limit > 0 && len(out) >= limit {
			return out, it.PageInfo().Token, nil
		}
	}
}

func credType(path string) option.CredentialsType {
	b, err := os.ReadFile(path)
	if err == nil && strings.Contains(string(b), `"authorized_user"`) {
		return option.AuthorizedUser
	}
	return option.ServiceAccount
}
