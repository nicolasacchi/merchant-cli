package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/nicolasacchi/clicore/cierrors"
	"github.com/nicolasacchi/merchant-cli/internal/commands"
	"google.golang.org/api/googleapi"
)

var version = "dev"

func main() {
	commands.SetVersion(version)
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}

// exitCode maps a Google API error to the fleet-canonical exit-code table
// (auth=2, validation=3, not_found=4, rate_limited=5, else 1) so $? is portable
// across the CLI fleet. Non-Google errors exit 1.
func exitCode(err error) int {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		return cierrors.ExitCodeFor(gerr.Code, "")
	}
	return 1
}
