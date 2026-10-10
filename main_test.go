package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/seachicken/gh-poi/conn"
	"github.com/seachicken/gh-poi/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type mockMainConn struct {
	*mocks.MockConnection
	pruneRemoteErr    error
	pruneWorktreesErr error
}

func (m *mockMainConn) PruneRemoteBranches(ctx context.Context, remoteName string) (string, error) {
	return "", m.pruneRemoteErr
}

func (m *mockMainConn) PruneWorktrees(ctx context.Context) (string, error) {
	return "", m.pruneWorktreesErr
}

func newMockMainConn(stub *conn.Stub) *mockMainConn {
	return &mockMainConn{MockConnection: stub.Conn}
}

func TestExitCode(t *testing.T) {
	t.Run("returns 0 for nil error", func(t *testing.T) {
		assert.Equal(t, 0, exitCode(nil))
	})

	t.Run("returns 1 for non-exec generic error", func(t *testing.T) {
		assert.Equal(t, 1, exitCode(errors.New("generic error")))
	})

	t.Run("returns child exit code for exec.ExitError", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 42")
		err := cmd.Run()
		assert.Error(t, err)
		assert.Equal(t, 42, exitCode(err))
	})
}

func TestRunCLI_Flags(t *testing.T) {
	t.Run("returns 0 on --help", func(t *testing.T) {
		err := runCLI([]string{"--help"})
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode(err))
	})

	t.Run("returns non-zero on invalid flag", func(t *testing.T) {
		err := runCLI([]string{"--nonexistent-flag"})
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})

	t.Run("returns non-zero on unknown subcommand", func(t *testing.T) {
		err := runCLI([]string{"unknown-command"})
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})

	t.Run("returns 0 on lock --help", func(t *testing.T) {
		err := runCLI([]string{"lock", "--help"})
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode(err))
	})

	t.Run("returns 0 on unlock --help", func(t *testing.T) {
		err := runCLI([]string{"unlock", "--help"})
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode(err))
	})
}

func TestRunMainWithConn_Errors(t *testing.T) {
	t.Run("propagates error when GetPreferredRemotes fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("failed to run external command: git, args: [remote -v]")
		s := conn.Setup(ctrl).GetRemoteNames("origin", expectedErr, nil)

		err := runMainWithConn(context.Background(), newMockMainConn(s), Merged, Quick, true, false)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.NotZero(t, exitCode(err))
	})

	t.Run("propagates error when GetBranches fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("failed to run external command: gh repo view")
		s := conn.Setup(ctrl).
			GetRemoteNames("origin", nil, nil).
			GetSshConfig("github.com", nil, nil).
			GetConfig([]conn.ConfigStub{
				{Key: "remote.origin.gh-resolved", Filename: "empty"},
			}, nil, nil).
			GetRepoNames([]conn.RepoNamesStub{
				{RepoName: "owner/repo", Filename: "origin"},
			}, expectedErr, nil)

		err := runMainWithConn(context.Background(), newMockMainConn(s), Merged, Quick, true, false)
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})

	t.Run("propagates error when DeleteBranches fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("failed to delete branches")
		s := conn.Setup(ctrl).
			GetRemoteNames("origin", nil, nil).
			GetSshConfig("github.com", nil, nil).
			GetRepoNames([]conn.RepoNamesStub{
				{RepoName: "owner/repo", Filename: "origin"},
			}, nil, nil).
			GetBranchNames("@main_issue1", nil, nil).
			GetMergedBranchNames("@main_issue1", nil, nil).
			GetLog([]conn.LogStub{
				{BranchName: "main", Filename: "main_issue1Merged"},
				{BranchName: "issue1", Filename: "issue1Merged"},
			}, nil, nil).
			GetPullRequests("issue1Merged", nil, nil).
			GetUncommittedChanges([]conn.UncommittedChangeStub{
				{Path: "", Output: ""},
			}, nil, nil).
			GetWorktrees("none", nil, nil).
			GetConfig([]conn.ConfigStub{
				{Key: "remote.origin.gh-resolved", Filename: "empty"},
				{Key: "branch.main.merge", Filename: "mergeMain"},
				{Key: "branch.main.gh-poi-locked", Filename: "empty"},
				{Key: "branch.main.gh-poi-protected", Filename: "empty"},
				{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
				{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
				{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
			}, nil, nil).
			DeleteBranches(expectedErr, nil)

		err := runMainWithConn(context.Background(), newMockMainConn(s), Merged, Quick, false, false)
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})

	t.Run("propagates error when PruneRemoteBranches fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("failed to prune remote branches")
		s := conn.Setup(ctrl).
			GetRemoteNames("origin", nil, nil).
			GetSshConfig("github.com", nil, nil).
			GetRepoNames([]conn.RepoNamesStub{
				{RepoName: "owner/repo", Filename: "origin"},
			}, nil, nil).
			GetBranchNames("@main_issue1", nil, nil).
			GetMergedBranchNames("@main_issue1", nil, nil).
			GetLog([]conn.LogStub{
				{BranchName: "main", Filename: "main_issue1Merged"},
				{BranchName: "issue1", Filename: "issue1Merged"},
			}, nil, nil).
			GetPullRequests("issue1Merged", nil, nil).
			GetUncommittedChanges([]conn.UncommittedChangeStub{
				{Path: "", Output: ""},
			}, nil, nil).
			GetWorktrees("none", nil, nil).
			GetConfig([]conn.ConfigStub{
				{Key: "remote.origin.gh-resolved", Filename: "empty"},
				{Key: "branch.main.merge", Filename: "mergeMain"},
				{Key: "branch.main.gh-poi-locked", Filename: "empty"},
				{Key: "branch.main.gh-poi-protected", Filename: "empty"},
				{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
				{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
				{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
			}, nil, nil).
			DeleteBranches(nil, nil)

		m := newMockMainConn(s)
		m.pruneRemoteErr = expectedErr

		err := runMainWithConn(context.Background(), m, Merged, Quick, false, false)
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})

	t.Run("propagates error when PruneWorktrees fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("failed to prune worktrees")
		s := conn.Setup(ctrl).
			GetRemoteNames("origin", nil, nil).
			GetSshConfig("github.com", nil, nil).
			GetRepoNames([]conn.RepoNamesStub{
				{RepoName: "owner/repo", Filename: "origin"},
			}, nil, nil).
			GetBranchNames("@main_issue1", nil, nil).
			GetMergedBranchNames("@main_issue1", nil, nil).
			GetLog([]conn.LogStub{
				{BranchName: "main", Filename: "main_issue1Merged"},
				{BranchName: "issue1", Filename: "issue1Merged"},
			}, nil, nil).
			GetPullRequests("issue1Merged", nil, nil).
			GetUncommittedChanges([]conn.UncommittedChangeStub{
				{Path: "", Output: ""},
			}, nil, nil).
			GetWorktrees("none", nil, nil).
			GetConfig([]conn.ConfigStub{
				{Key: "remote.origin.gh-resolved", Filename: "empty"},
				{Key: "branch.main.merge", Filename: "mergeMain"},
				{Key: "branch.main.gh-poi-locked", Filename: "empty"},
				{Key: "branch.main.gh-poi-protected", Filename: "empty"},
				{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
				{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
				{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
			}, nil, nil).
			DeleteBranches(nil, nil)

		m := newMockMainConn(s)
		m.pruneWorktreesErr = expectedErr

		err := runMainWithConn(context.Background(), m, Merged, Quick, false, false)
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})

	t.Run("succeeds on dry run", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := conn.Setup(ctrl).
			GetRemoteNames("origin", nil, nil).
			GetSshConfig("github.com", nil, nil).
			GetRepoNames([]conn.RepoNamesStub{
				{RepoName: "owner/repo", Filename: "origin"},
			}, nil, nil).
			GetBranchNames("@main_issue1", nil, nil).
			GetMergedBranchNames("@main_issue1", nil, nil).
			GetLog([]conn.LogStub{
				{BranchName: "main", Filename: "main_issue1Merged"},
				{BranchName: "issue1", Filename: "issue1Merged"},
			}, nil, nil).
			GetPullRequests("issue1Merged", nil, nil).
			GetUncommittedChanges([]conn.UncommittedChangeStub{
				{Path: "", Output: ""},
			}, nil, nil).
			GetWorktrees("none", nil, nil).
			GetConfig([]conn.ConfigStub{
				{Key: "remote.origin.gh-resolved", Filename: "empty"},
				{Key: "branch.main.merge", Filename: "mergeMain"},
				{Key: "branch.main.gh-poi-locked", Filename: "empty"},
				{Key: "branch.main.gh-poi-protected", Filename: "empty"},
				{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
				{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
				{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
			}, nil, nil)

		err := runMainWithConn(context.Background(), newMockMainConn(s), Merged, Quick, true, false)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode(err))
	})
}

func TestRunLockWithConn_Errors(t *testing.T) {
	t.Run("propagates error when GetBranchNames fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("failed to get branch names")
		s := conn.Setup(ctrl).GetBranchNames("@main_issue1", expectedErr, nil)

		err := runLockWithConn(context.Background(), s.Conn, []string{"main"})
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})
}

func TestRunUnlockWithConn_Errors(t *testing.T) {
	t.Run("propagates error when GetBranchNames fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedErr := errors.New("failed to get branch names")
		s := conn.Setup(ctrl).GetBranchNames("@main_issue1", expectedErr, nil)

		err := runUnlockWithConn(context.Background(), s.Conn, []string{"main"})
		assert.Error(t, err)
		assert.NotZero(t, exitCode(err))
	})
}

func TestRunMain_RealGitFailureOutsideRepo(t *testing.T) {
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(origDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	err = runMain(Merged, Quick, true, false)
	assert.Error(t, err)
	assert.NotZero(t, exitCode(err))
}

func TestCLI_SubprocessExitNonZeroOnGitFailure(t *testing.T) {
	if os.Getenv("TEST_CLI_SUBPROCESS") == "1" {
		os.Args = []string{"gh-poi", "--dry-run"}
		main()
		return
	}

	tempDir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_SubprocessExitNonZeroOnGitFailure")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "TEST_CLI_SUBPROCESS=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.NotZero(t, exitErr.ExitCode())
}

func TestCLI_SubprocessExitNonZeroOnUnknownSubcommand(t *testing.T) {
	if os.Getenv("TEST_CLI_SUBPROCESS_UNKNOWN") == "1" {
		os.Args = []string{"gh-poi", "unknowncmd"}
		main()
		return
	}

	tempDir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_SubprocessExitNonZeroOnUnknownSubcommand")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "TEST_CLI_SUBPROCESS_UNKNOWN=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	assert.True(t, errors.As(err, &exitErr))
	assert.NotZero(t, exitErr.ExitCode())
}
