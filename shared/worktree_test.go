package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseWorktreesBasic(t *testing.T) {
	assert.Equal(t,
		[]Worktree{
			{Path: "/home/user/project", Branch: "main", IsMain: true},
			{Path: "/home/user/project-feature", Branch: "feature-branch", IsMain: false},
		},
		ParseWorktrees(`worktree /home/user/project
HEAD abc123
branch refs/heads/main

worktree /home/user/project-feature
HEAD def456
branch refs/heads/feature-branch`),
	)
}

func Test_ParseWorktreesEmpty(t *testing.T) {
	assert.Equal(t,
		[]Worktree{},
		ParseWorktrees(""),
	)
}

func Test_ParseWorktreesDetachedHead(t *testing.T) {
	assert.Equal(t,
		[]Worktree{
			{Path: "/home/user/project", Branch: "", IsMain: true},
		},
		ParseWorktrees(`worktree /home/user/project
HEAD abc123
detached`),
	)
}
