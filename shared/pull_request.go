package shared

type (
	PullRequestState int

	PullRequest struct {
		Name    string
		State   PullRequestState
		IsDraft bool
		Number  int
		Commits []string
		Url     string
		Author  string
	}
)

const (
	Closed PullRequestState = iota
	Merged
	Open
)
