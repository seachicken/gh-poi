package shared

import "regexp"

type (
	BranchState int

	Branch struct {
		Head          bool
		Name          string
		IsMerged      bool
		IsProtected   bool
		RemoteHeadOid string
		Commits       []string
		PullRequests  []PullRequest
		State         BranchState
	}
)

const (
	Unknown BranchState = iota
	NotDeletable
	Deletable
	Deleted
)

func (b Branch) IsDetached() bool {
	detachedBranchNameRegex := regexp.MustCompile(`^\(.+\)`)
	return detachedBranchNameRegex.MatchString(b.Name)
}
