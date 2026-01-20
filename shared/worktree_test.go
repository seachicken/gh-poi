package shared

import (
	"testing"

	"github.com/seachicken/gh-poi/conn"
	"github.com/stretchr/testify/assert"
)

func Test_ParseWorktreesWithLinkedWorktree(t *testing.T) {
	stub := (&conn.Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "linked")
	assert.Equal(t,
		[]Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Branch: "main", IsMain: true, IsLocked: false},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Branch: "linkedIssue1", IsMain: false, IsLocked: false},
		},
		ParseWorktrees(stub),
	)
}

func Test_ParseWorktreesWithoutLinkedWorktree(t *testing.T) {
	stub := (&conn.Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "none")
	assert.Equal(t,
		[]Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_basic", Branch: "main", IsMain: true, IsLocked: false},
		},
		ParseWorktrees(stub),
	)
}

func Test_ParseWorktreesWithDetached(t *testing.T) {
	stub := (&conn.Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "detached")
	assert.Equal(t,
		[]Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Branch: "main", IsMain: true, IsLocked: false},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Branch: "", IsMain: false, IsLocked: false},
		},
		ParseWorktrees(stub),
	)
}

func Test_ParseWorktreesWithLocked(t *testing.T) {
	stub := (&conn.Stub{Conn: nil, T: t}).ReadFile("git", "worktree", "locked")
	assert.Equal(t,
		[]Worktree{
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Branch: "main", IsMain: true, IsLocked: false},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Branch: "linkedIssue1", IsMain: false, IsLocked: true},
		},
		ParseWorktrees(stub),
	)
}
