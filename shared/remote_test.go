package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResolvedRepoName(t *testing.T) {
	t.Run("returns repo from git config URL when gh-resolved is unset", func(t *testing.T) {
		assert.Equal(t,
			"owner/repo",
			Remote{RepoName: "owner/repo"}.ResolvedRepoName(),
		)
	})

	t.Run("returns repo from git config URL when gh-resolved base is set", func(t *testing.T) {
		assert.Equal(t,
			"owner/repo",
			Remote{RepoName: "owner/repo", GhResolved: "base"}.ResolvedRepoName(),
		)
	})

	t.Run("returns repo from gh-resolved when gh-resolved repo is set", func(t *testing.T) {
		assert.Equal(t,
			"upstream/repo",
			Remote{RepoName: "owner/repo", GhResolved: "upstream/repo"}.ResolvedRepoName(),
		)
	})
}
