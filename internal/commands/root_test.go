package commands

import "testing"

func TestGetMerchantID(t *testing.T) {
	orig := merchantID
	t.Cleanup(func() { merchantID = orig })

	// --id flag wins over the env var.
	merchantID = "flag-id"
	t.Setenv("MERCHANT_ID", "env-id")
	if got := getMerchantID(); got != "flag-id" {
		t.Errorf("flag should win: got %q", got)
	}

	// Falls back to MERCHANT_ID when the flag is empty.
	merchantID = ""
	if got := getMerchantID(); got != "env-id" {
		t.Errorf("env fallback: got %q", got)
	}

	// Empty when neither is set.
	merchantID = ""
	t.Setenv("MERCHANT_ID", "")
	if got := getMerchantID(); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}
