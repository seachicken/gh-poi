package main

import (
	"context"
	"fmt"
)

func ProtectBranches(ctx context.Context, targetBranchNames []string, connection Connection) error {
	branchNameResults, err := connection.GetBranchNames(ctx)
	if err != nil {
		return err
	}
	branches := toBranch(splitLines(branchNameResults))

	for _, targetName := range targetBranchNames {
		if branchNameExists(targetName, branches) {
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName))
			_, err = connection.AddConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName), "true")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func UnprotectBranches(ctx context.Context, targetBranchNames []string, connection Connection) error {
	branchNameResults, err := connection.GetBranchNames(ctx)
	if err != nil {
		return err
	}
	branches := toBranch(splitLines(branchNameResults))

	for _, targetName := range targetBranchNames {
		if branchNameExists(targetName, branches) {
			connection.RemoveConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", targetName))
		}
	}

	return nil
}
