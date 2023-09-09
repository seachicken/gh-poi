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

var detachedBranchNameRegex = regexp.MustCompile(`^\(.+\)`)

func (b Branch) IsDetached() bool {
	return detachedBranchNameRegex.MatchString(b.Name)
}
