package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseWorktrees(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		input := `worktree /home/user/project
HEAD abc123
branch refs/heads/main

worktree /home/user/project-feature
HEAD def456
branch refs/heads/feature-branch`

		worktrees := ParseWorktrees(input)

		assert.Equal(t, 2, len(worktrees))
		assert.Equal(t, "/home/user/project", worktrees[0].Path)
		assert.Equal(t, "main", worktrees[0].Branch)
		assert.True(t, worktrees[0].IsMain)
		assert.Equal(t, "/home/user/project-feature", worktrees[1].Path)
		assert.Equal(t, "feature-branch", worktrees[1].Branch)
		assert.False(t, worktrees[1].IsMain)
	})

	t.Run("empty", func(t *testing.T) {
		worktrees := ParseWorktrees("")
		assert.Equal(t, 0, len(worktrees))
	})

	t.Run("detached HEAD", func(t *testing.T) {
		input := `worktree /home/user/project
HEAD abc123
detached`

		worktrees := ParseWorktrees(input)

		assert.Equal(t, 1, len(worktrees))
		assert.Equal(t, "/home/user/project", worktrees[0].Path)
		assert.Equal(t, "", worktrees[0].Branch)
		assert.True(t, worktrees[0].IsMain)
	})
}
