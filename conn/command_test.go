package conn

import (
	"testing"

	"github.com/seachicken/gh-poi/shared"
	"github.com/stretchr/testify/assert"
)

func Test_CreateRemoteWithScpLikeUrl(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	git@github.com:org/repo (fetch)"),
	)
}

func Test_CreateRemoteWithScpLikeUrlAndCustomUserinfo(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	git0-._~@github.com:org/repo (fetch)"),
	)
}

func Test_CreateRemoteWithSshUrl(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	ssh://git@github.com/org/repo.git (fetch)"),
	)
}

func Test_CreateRemoteWithScpLikeUrlWithoutUserinfo(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	github.com:org/repo.git (fetch)"),
	)
}

func Test_CreateRemoteWithHttps(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	https://github.com/org/repo.git (fetch)"),
	)
}

// https://github.com/seachicken/gh-poi/issues/152
func Test_CreateRemoteWithHttpsTrailingSlash(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	https://github.com/org/repo.git/ (fetch)"),
	)
}

// https://github.com/seachicken/gh-poi/issues/152
func Test_CreateRemoteWithHttpsTrailingSlashWithoutDotGit(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	https://github.com/org/repo/ (fetch)"),
	)
}

// https://github.com/seachicken/gh-poi/issues/152
func Test_CreateRemoteWithSshUrlTrailingSlash(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	ssh://git@github.com/org/repo.git/ (fetch)"),
	)
}

// https://github.com/seachicken/gh-poi/issues/152
func Test_CreateRemoteWithScpLikeUrlTrailingSlash(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	git@github.com:org/repo.git/ (fetch)"),
	)
}

// https://github.com/seachicken/gh-poi/issues/39
func Test_CreateRemoteWithCustomHostname(t *testing.T) {
	assert.Equal(t,
		[]shared.Remote{
			{
				Name:     "origin",
				Hostname: "github.com-work",
				RepoName: "org/repo",
			},
		},
		parseRemotes("origin	git@github.com-work:org/repo.git (fetch)"),
	)
}

func Test_ParseWorktreesWithLinkedWorktree(t *testing.T) {
	stub := (&Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "@main_+linkedIssue1")
	assert.Equal(t,
		[]shared.Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Branch: "main", IsMain: true, IsLocked: false},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Branch: "linkedIssue1", IsMain: false, IsLocked: false},
		},
		parseWorktrees(stub),
	)
}

func Test_ParseWorktreesWithoutLinkedWorktree(t *testing.T) {
	stub := (&Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "none")
	assert.Equal(t,
		[]shared.Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_basic", Branch: "main", IsMain: true, IsLocked: false},
		},
		parseWorktrees(stub),
	)
}

func Test_ParseWorktreesWithDetached(t *testing.T) {
	stub := (&Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "detached")
	assert.Equal(t,
		[]shared.Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Branch: "main", IsMain: true, IsLocked: false},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Branch: "", IsMain: false, IsLocked: false},
		},
		parseWorktrees(stub),
	)
}

func Test_ParseWorktreesWithLocked(t *testing.T) {
	stub := (&Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "locked")
	assert.Equal(t,
		[]shared.Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Branch: "main", IsMain: true, IsLocked: false},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Branch: "linkedIssue1", IsMain: false, IsLocked: true},
		},
		parseWorktrees(stub),
	)
}
