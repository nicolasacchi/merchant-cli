package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/nicolasacchi/clicore/cierrors"
	"github.com/nicolasacchi/merchant-cli/internal/commands"
	"google.golang.org/api/googleapi"
)

var version = "dev"

// resolveVersion prefers the -ldflags value (Makefile, goreleaser) and falls
// back to the module version so `go install ...@vX.Y.Z` reports the tag.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}

func main() {
	commands.SetVersion(resolveVersion())
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
