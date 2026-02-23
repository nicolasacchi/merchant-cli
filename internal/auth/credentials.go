package auth

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/shopping/merchant/reports/apiv1beta"
	content "google.golang.org/api/content/v2.1"
	"google.golang.org/api/option"
)

func credentialOption() option.ClientOption {
	path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if path != "" {
		return option.WithCredentialsFile(path)
	}
	return nil
}

func NewContentService(ctx context.Context) (*content.APIService, error) {
	opt := credentialOption()
	var opts []option.ClientOption
	if opt != nil {
		opts = append(opts, opt)
	}
	svc, err := content.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("content api service: %w", err)
	}
	return svc, nil
}

func NewReportClient(ctx context.Context) (*reports.ReportClient, error) {
	opt := credentialOption()
	var opts []option.ClientOption
	if opt != nil {
		opts = append(opts, opt)
	}
	client, err := reports.NewReportRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("report client: %w", err)
	}
	return client, nil
}
