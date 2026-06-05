package output

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn with os.Stdout redirected to a pipe and returns what it
// wrote. PrintJSON writes to os.Stdout directly, so this is the way to observe it.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out), runErr
}

func TestPrintJSON_ValidIndentedOutput(t *testing.T) {
	out, err := captureStdout(t, func() error {
		return PrintJSON(map[string]any{"id": "196138351", "ok": true})
	})
	if err != nil {
		t.Fatalf("PrintJSON: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if decoded["id"] != "196138351" {
		t.Errorf("round-trip lost id: %#v", decoded)
	}
	if !strings.Contains(out, "\n  ") {
		t.Errorf("output is not 2-space indented:\n%s", out)
	}
}

func TestPrintJSON_UnencodableValueErrors(t *testing.T) {
	_, err := captureStdout(t, func() error {
		return PrintJSON(make(chan int)) // channels can't be JSON-encoded
	})
	if err == nil {
		t.Fatal("want an error for an unencodable value")
	}
}
