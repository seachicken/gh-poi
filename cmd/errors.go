package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/seachicken/gh-poi/conn"
	"github.com/seachicken/gh-poi/shared"
)

// HandleRepoError checks for the sentinel repository-missing error and prints
// a single, user-friendly message. It returns true when it handled the error.
func HandleRepoError(err error) bool {
	if errors.Is(err, conn.ErrNotAGitRepository) {
		fmt.Fprintln(os.Stderr, shared.NoRepoMsg)
		return true
	}
	return false
}
