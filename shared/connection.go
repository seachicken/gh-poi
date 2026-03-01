//go:generate mockgen -source=connection.go -package=mocks -destination=../mocks/poi_mock.go
package shared

import "context"

type Connection interface {
	CheckRepos(ctx context.Context, hostname string, repoNames []string) error
	GetRemoteNames(ctx context.Context) (string, error)
	GetSshConfig(ctx context.Context, name string) (string, error)
	GetRepoNames(ctx context.Context, hostname string, repoName string) (string, error)
	GetBranchNames(ctx context.Context) (string, error)
	GetMergedBranchNames(ctx context.Context, remoteName string, branchName string) (string, error)
	GetRemoteHeadOid(ctx context.Context, remoteName string, branchName string) (string, error)
	GetLsRemoteHeadOid(ctx context.Context, url string, branchName string) (string, error)
	GetLog(ctx context.Context, branchName string) (string, error)
	GetAssociatedRefNames(ctx context.Context, oid string) (string, error)
	GetPullRequests(ctx context.Context, hostname string, orgs string, repos string, queryHashes string) (string, error)
	GetUncommittedChanges(ctx context.Context) (string, error)
	GetConfig(ctx context.Context, key string) (string, error)
	AddConfig(ctx context.Context, key string, value string) (string, error)
	RemoveConfig(ctx context.Context, key string) (string, error)
	CheckoutBranch(ctx context.Context, branchName string) (string, error)
	DeleteBranches(ctx context.Context, branchNames []string) (string, error)
	GetWorktrees(ctx context.Context) (string, error)
	RemoveWorktree(ctx context.Context, path string) (string, error)
	// IsLocalRepo returns true if the current working directory is inside a git
	// repository.  The concrete implementation should run a lightweight git check
	// such as `git rev-parse --is-inside-work-tree`.
	IsLocalRepo(ctx context.Context) (bool, error)
}
