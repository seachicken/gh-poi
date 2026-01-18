package lock

import (
	"context"
	"fmt"

	"github.com/seachicken/gh-poi/cmd"
	"github.com/seachicken/gh-poi/shared"
)

func LockBranches(ctx context.Context, targetBranchNames []string, connection shared.Connection) error {
	branchNameResults, err := connection.GetBranchNames(ctx)
	if err != nil {
		return err
	}
	branches := cmd.ToBranch(cmd.SplitLines(branchNameResults))

	for _, targetName := range targetBranchNames {
		if cmd.BranchNameExists(targetName, branches) {
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-locked", targetName))
			// TODO: Remove after deprecated commands are removed
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName))
			_, err = connection.AddConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-locked", targetName), "true")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func UnlockBranches(ctx context.Context, targetBranchNames []string, connection shared.Connection) error {
	branchNameResults, err := connection.GetBranchNames(ctx)
	if err != nil {
		return err
	}
	branches := cmd.ToBranch(cmd.SplitLines(branchNameResults))

	for _, targetName := range targetBranchNames {
		if cmd.BranchNameExists(targetName, branches) {
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-locked", targetName))
			// TODO: Remove after deprecated commands are removed
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName))
		}
	}

	return nil
}
