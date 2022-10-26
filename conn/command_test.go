package conn

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

// $ git log --all --graph --pretty=oneline
// * a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0 (issue1) 1-1
// * 6ebe3d30d23531af56bd23b5a098d3ccae2a534a (HEAD -> main) Initial commit
func Test_RepoBasic(t *testing.T) {
	setGitDir("repo_basic", t)
	conn := &Connection{}
	stub := &Stub{nil, t}

	t.Run("GetRemoteNames", func(t *testing.T) {
		actual, _ := conn.GetRemoteNames(context.Background())
		assert.Equal(t,
			stub.readFile("git", "remote", "origin"),
			actual,
		)
	})

	t.Run("GetBranchNames", func(t *testing.T) {
		actual, _ := conn.GetBranchNames(context.Background())
		assert.Equal(t,
			stub.readFile("git", "branch", "@main_issue1"),
			actual,
		)
	})

	t.Run("GetLog", func(t *testing.T) {

		t.Run("main", func(t *testing.T) {
			actual, _ := conn.GetLog(context.Background(), "main")
			assert.Equal(t,
				stub.readFile("git", "log", "main"),
				actual,
			)
		})

		t.Run("issue1", func(t *testing.T) {
			actual, _ := conn.GetLog(context.Background(), "issue1")
			assert.Equal(t,
				stub.readFile("git", "log", "issue1"),
				actual,
			)
		})
	})

	t.Run("GetAssociatedRefNames", func(t *testing.T) {

		t.Run("issue1", func(t *testing.T) {
			actual, _ := conn.GetAssociatedRefNames(context.Background(), "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0")
			assert.Equal(t,
				stub.readFile("git", "abranch", "issue1"),
				actual,
			)
		})

		t.Run("main_issue1", func(t *testing.T) {
			actual, _ := conn.GetAssociatedRefNames(context.Background(), "6ebe3d30d23531af56bd23b5a098d3ccae2a534a")
			assert.Equal(t,
				stub.readFile("git", "abranch", "main_issue1"),
				actual,
			)
		})
	})

	t.Run("GetUncommittedChanges", func(t *testing.T) {
		actual, _ := conn.GetUncommittedChanges(context.Background())
		assert.Equal(t, "A  README.md\n", actual)
	})

	t.Run("GetConfig", func(t *testing.T) {
		actual, _ := conn.GetConfig(context.Background(), "branch.main.merge")
		assert.Equal(t,
			stub.readFile("git", "configMerge", "main"),
			actual,
		)
	})
}

func setGitDir(repoName string, t *testing.T) {
	gitDirOrg := os.Getenv("GIT_DIR")
	gitWorkTreeOrg := os.Getenv("GIT_WORK_TREE")

	os.Setenv("GIT_DIR", path.Join(fixturePath, repoName, ".git"))
	os.Setenv("GIT_WORK_TREE", path.Join(fixturePath, repoName))

	t.Cleanup(func() {
		os.Setenv("GIT_DIR", gitDirOrg)
		os.Setenv("GIT_WORK_TREE", gitWorkTreeOrg)
	})
}
