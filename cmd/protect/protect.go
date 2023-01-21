package protect

import (
	"context"
	"fmt"

	"github.com/seachicken/gh-poi/cmd"
	"github.com/seachicken/gh-poi/shared"
)

func ProtectBranches(ctx context.Context, targetBranchNames []string, connection shared.Connection) error {
	branchNameResults, err := connection.GetBranchNames(ctx)
	if err != nil {
		return err
	}
	branches := cmd.ToBranch(cmd.SplitLines(branchNameResults))

	for _, targetName := range targetBranchNames {
		if cmd.BranchNameExists(targetName, branches) {
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName))
			_, err = connection.AddConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName), "true")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func UnprotectBranches(ctx context.Context, targetBranchNames []string, connection shared.Connection) error {
	branchNameResults, err := connection.GetBranchNames(ctx)
	if err != nil {
		return err
	}
	branches := cmd.ToBranch(cmd.SplitLines(branchNameResults))

	for _, targetName := range targetBranchNames {
		if cmd.BranchNameExists(targetName, branches) {
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName))
		}
	}

	return nil
}
