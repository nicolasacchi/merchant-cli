package main

import (
	"errors"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestExitCode(t *testing.T) {
	cases := map[int]int{401: 2, 403: 2, 400: 3, 404: 4, 429: 5, 500: 1, 200: 1}
	for code, want := range cases {
		if got := exitCode(&googleapi.Error{Code: code}); got != want {
			t.Errorf("exitCode(googleapi %d) = %d, want %d", code, got, want)
		}
	}
	if got := exitCode(errors.New("plain")); got != 1 {
		t.Errorf("non-googleapi error should exit 1, got %d", got)
	}
}
