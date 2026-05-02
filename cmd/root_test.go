package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/seachicken/gh-poi/conn"
	"github.com/seachicken/gh-poi/shared"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var ErrCommand = errors.New("failed to run external command")

func Test_GetPreferredRemotes(t *testing.T) {
	t.Run("with quick scan", func(t *testing.T) {
		scan := shared.Quick

		t.Run("returns origin as highest priority", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetRemoteNames("origin_upstream", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "remote.upstream.gh-resolved", Filename: "empty"},
				}, nil, nil)

			actual, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)
			assert.Equal(t, 1, len(actual))
			assert.Equal(t, "origin", actual[0].Name)
		})

		t.Run("returns first remote when origin is missing", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetRemoteNames("midstream", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.midstream.gh-resolved", Filename: "empty"},
				}, nil, nil)

			actual, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)
			assert.Equal(t, 1, len(actual))
			assert.Equal(t, "midstream", actual[0].Name)
		})

		t.Run("returns origin and gh-resolved when gh-resolved is configured", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetRemoteNames("origin_upstream", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "remote.upstream.gh-resolved", Filename: "ghResolved"},
				}, nil, nil)

			actual, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)
			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "origin", actual[0].Name)
			assert.Equal(t, "", actual[0].GhResolved)
			assert.Equal(t, "upstream", actual[1].Name)
			assert.Equal(t, "base", actual[1].GhResolved)
		})
	})

	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		t.Run("returns origin as highest priority", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetRemoteNames("origin_upstream", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "remote.upstream.gh-resolved", Filename: "empty"},
				}, nil, nil)

			actual, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)
			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "origin", actual[0].Name)
			assert.Equal(t, "upstream", actual[1].Name)
		})

		t.Run("returns first remote when origin is missing", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetRemoteNames("midstream", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.midstream.gh-resolved", Filename: "empty"},
				}, nil, nil)

			actual, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)
			assert.Equal(t, 1, len(actual))
			assert.Equal(t, "midstream", actual[0].Name)
		})

		t.Run("returns origin and gh-resolved when gh-resolved is configured", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetRemoteNames("origin_upstream", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "remote.upstream.gh-resolved", Filename: "ghResolved"},
				}, nil, nil)

			actual, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)
			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "origin", actual[0].Name)
			assert.Equal(t, "", actual[0].GhResolved)
			assert.Equal(t, "upstream", actual[1].Name)
			assert.Equal(t, "base", actual[1].GhResolved)
		})
	})
}

/*
// Before
// main  : *---*---*
//          \     /
// topic :   *---* (PR merged)
*/
func Test_GetBranchesWhenMergedPR(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@main_issue1", nil, nil).
				GetMergedBranchNames("@main_issue1", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "issue1", Filename: "issue1Merged"},
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
		}

		t.Run("deletable when branch is merged", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("deletable when branch is merged with associated refs", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "b8a2645298053fb62ea03e27feea6c483d3fd27e", Filename: "main_issue1"},
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "main_issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
				}, nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("not deletable when branch is locked", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetConfig([]conn.ConfigStub{
					{Key: "branch.issue1.gh-poi-locked", Filename: "locked"},
				}, nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, true, actual[0].IsLocked)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, false, actual[1].IsLocked)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		// TODO: Remove after deprecated commands are removed
		t.Run("not deletable when branch is protected", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetConfig([]conn.ConfigStub{
					{Key: "branch.issue1.gh-poi-protected", Filename: "locked"},
				}, nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, true, actual[0].IsLocked)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, false, actual[1].IsLocked)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// main  : *---*---*
//          \   ../
// topic :   *---* (PR merged)
*/
func Test_GetBranchesWhenSquashAndMergedPR(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@main_issue1", nil, nil).
				GetMergedBranchNames("@main", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
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
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil)
		}

		t.Run("deletable", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("deletable with dry-run option", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				CheckoutBranch(nil, conn.NewConf(&conn.Times{N: 0}))
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, true)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// upstream/main : *---*---*
//                  \   ../
// topic         :   *---* (PR merged)
*/
func Test_GetBranchesWhenSquashAndMergedPRByUpstream(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin_upstream"},
				}, nil, nil).
				GetBranchNames("@main_issue1", nil, nil).
				GetMergedBranchNames("@main", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
				}, nil, nil).
				GetPullRequests("issue1UpMerged", nil, nil).
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
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil)
		}

		t.Run("deletable", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("deletable with gh pr checkout branch", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetBranchNames("@main_forkMain", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main"}, {BranchName: "fork/main", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "forkMain"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_forkMain"},
				}, nil, nil).
				GetPullRequests("forkMainUpMerged", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "branch.fork/main.merge", Filename: "mergeForkMain"},
					{Key: "branch.fork/main.remote", Filename: "remote"},
					{Key: "branch.fork/main.gh-poi-locked", Filename: "empty"},
					{Key: "branch.fork/main.gh-poi-protected", Filename: "empty"},
				}, nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "fork/main", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// upstream/main : *---*---*
//                  \     /
// main          :   *---* (PR merged)
*/
func Test_GetBranchesWhenMergedPRWithDefaultBranchAsHeadRef(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@main_issue1", nil, nil).
				GetMergedBranchNames("@main", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
				}, nil, nil).
				GetPullRequests("mainMerged", nil, nil).
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
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil)
		}

		t.Run("not deletable when head ref is default branch", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// origin/main : *---*---*
//                \   ../
// topic       :   *---* (PR merged)
*/
func Test_GetBranchesWhenSquashAndMergedToOriginAndMissingDefaultBranch(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@issue1", nil, nil).
				GetMergedBranchNames("empty", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "issue1_originMain"},
				}, nil, nil).
				GetPullRequests("issue1Merged", nil, nil).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "", Output: ""},
				}, nil, nil).
				GetWorktrees("none", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil).
				CheckoutBranch(nil, nil)
		}

		t.Run("deletable when default branch does not exists", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "(HEAD detached at origin/main)", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "issue1", actual[1].Name)
			assert.Equal(t, shared.Deletable, actual[1].State)
		})
	})
}

/*
// Before
// upstream/main : *-------*
//                  \   ../
// topic         :   *---* (PR merged)
*/
func Test_GetBranchesWhenSquashAndMergedToUpstreamAndMissingDefaultBranch(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin_upstream", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@issue1", nil, nil).
				GetMergedBranchNames("empty", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "issue1_originMain"},
				}, nil, nil).
				GetPullRequests("issue1Merged", nil, nil).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "", Output: ""},
				}, nil, nil).
				GetWorktrees("none", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "remote.upstream.gh-resolved", Filename: "ghResolved"},
					{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil).
				CheckoutBranch(nil, nil)
		}

		t.Run("deletable when default branch does not exists", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "(HEAD detached at upstream/main)", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "issue1", actual[1].Name)
			assert.Equal(t, shared.Deletable, actual[1].State)
		})
	})
}

/*
// Before:
// main  : *---*---*
//          \   ../
// topic :   *---* (PR merged) ---+
*/
func Test_GetBranchesWhenSquashAndMergedPRWithChanges(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("main_@issue1", nil, nil).
				GetMergedBranchNames("main", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
				}, nil, nil).
				GetPullRequests("issue1Merged", nil, nil).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "", Output: " M README.md"},
					{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_basic", Output: ""},
				}, nil, nil).
				GetWorktrees("none", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "branch.main.merge", Filename: "mergeMain"},
					{Key: "branch.main.gh-poi-locked", Filename: "empty"},
					{Key: "branch.main.gh-poi-protected", Filename: "empty"},
					{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil).
				CheckoutBranch(nil, nil)
		}

		t.Run("not deletable with uncommitted changes", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("deletable with untracked files", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "", Output: "?? new.txt"},
				}, nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// main  : *---*---*
//          \   ../
// topic :   *---* (PR merged) ---*
*/
func Test_GetBranchesWhenMergedPRWithNotFullyMerged(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@main_issue1", nil, nil).
				GetMergedBranchNames("@main", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main_issue1SquashAndMerged"}, {BranchName: "issue1", Filename: "issue1CommitAfterMerge"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "cb197ba87e4ad323b1008c611212deb7da2a4a49", Filename: "main"},
					{Oid: "b8a2645298053fb62ea03e27feea6c483d3fd27e", Filename: "issue1"},
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
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
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil)
		}

		t.Run("not deletable", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// main  : *---*---*
//          \
// topic :   *---* (PR closed)
*/
func Test_GetBranchesWhenClosedPR(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@main_issue1", nil, nil).
				GetMergedBranchNames("@main", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
				}, nil, nil).
				GetPullRequests("issue1Closed", nil, nil).
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
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil)
		}

		t.Run("not deletable with state option is merged", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("deletable with state option is closed", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Closed, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// main  : *---*---*
//          \   ../
// topic :   *---* (PR #1 closed, PR #2 merged)
*/
func Test_GetBranchesWhenClosedAndMergedPRs(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@main_issue1", nil, nil).
				GetMergedBranchNames("@main", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
				}, nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
				}, nil, nil).
				GetPullRequests("issue1Merged_issue1Closed", nil, nil).
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
					{Key: "branch.issue1.remote", Filename: "remote"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil)
		}

		t.Run("deletable with state option is merged", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("deletable with state option is closed", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Closed, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// main  (main worktree)   : *---*---*
//                            \     /
// topic (linked worktree) :   *---* (PR merged)
*/
func Test_GetBranchesWhenMergedPRIsLinkedWorktree(t *testing.T) {
	t.Run("with deep scan", func(t *testing.T) {
		scan := shared.Deep

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@main_linkedIssue1", nil, nil).
				GetMergedBranchNames("main_@linkedIssue1", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "linkedIssue1", Filename: "issue1Merged"},
				}, nil, nil).
				GetPullRequests("linkedIssue1Merged", nil, nil).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "", Output: ""},
					{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Output: ""},
					{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Output: ""},
				}, nil, nil).
				GetWorktrees("@main_+linkedIssue1", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "branch.main.merge", Filename: "mergeMain"},
					{Key: "branch.main.gh-poi-locked", Filename: "empty"},
					{Key: "branch.main.gh-poi-protected", Filename: "empty"},
					{Key: "branch.linkedIssue1.merge", Filename: "mergeIssue1"},
					{Key: "branch.linkedIssue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.linkedIssue1.gh-poi-protected", Filename: "empty"},
				}, nil, nil).
				CheckoutBranch(nil, nil)
		}

		t.Run("deletable when HEAD is not delete target worktree", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetBranchNames("@main_linkedIssue1", nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "linkedIssue1", actual[0].Name)
			assert.Equal(t, shared.Deletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("not deletable when HEAD is linked worktree", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetBranchNames("main_@linkedIssue1", nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "linkedIssue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("not deletable with uncommited changes", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Output: " M README.md"},
				}, nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "linkedIssue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("not deletable with untracked files", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Output: "?? new.txt"},
				}, nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "linkedIssue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})

		t.Run("not deletable with locked worktree", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetBranchNames("@main_linkedIssue1", nil, nil).
				GetMergedBranchNames("@main_linkedIssue1", nil, nil).
				GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
					{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
					{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
				}, nil, nil).
				GetWorktrees("locked", nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "linkedIssue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "main", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

/*
// Before
// origin/main             : *---------*
//                            \\      /
// topic (main worktree)   :   \*----* (PR merged)
//                              \
// topic (linked worktree) :     *---*
*/
func Test_GetBranchesWhenMergedPRIsMainWorktree(t *testing.T) {
	t.Run("with quick scan", func(t *testing.T) {
		scan := shared.Quick

		setupDefault := func(s *conn.Stub) *conn.Stub {
			return s.
				GetRemoteNames("origin", nil, nil).
				GetSshConfig("github.com", nil, nil).
				GetRepoNames([]conn.RepoNamesStub{
					{RepoName: "owner/repo", Filename: "origin"},
				}, nil, nil).
				GetBranchNames("@issue1_issue2", nil, nil).
				GetMergedBranchNames("@main_issue1", nil, nil).
				GetLog([]conn.LogStub{
					{BranchName: "issue1", Filename: "issue1Merged"}, {BranchName: "issue2", Filename: "issue1"},
				}, nil, nil).
				GetPullRequests("issue1Merged", nil, nil).
				GetUncommittedChanges([]conn.UncommittedChangeStub{
					{Path: "", Output: ""},
					{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_main", Output: ""},
					{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_worktree_linkedIssue1", Output: ""},
				}, nil, nil).
				GetWorktrees("@mainIssue1_+linkedIssue2", nil, nil).
				GetConfig([]conn.ConfigStub{
					{Key: "remote.origin.gh-resolved", Filename: "empty"},
					{Key: "branch.issue1.merge", Filename: "mergeMain"},
					{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
					{Key: "branch.issue2.merge", Filename: "mergeIssue1"},
					{Key: "branch.issue2.gh-poi-locked", Filename: "empty"},
					{Key: "branch.issue2.gh-poi-protected", Filename: "empty"},
				}, nil, nil).
				CheckoutBranch(nil, nil)
		}

		t.Run("deletable when HEAD is main worktree", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 3, len(actual))
			assert.Equal(t, "(HEAD detached at origin/main)", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "issue1", actual[1].Name)
			assert.Equal(t, shared.Deletable, actual[1].State)
			assert.Equal(t, "issue2", actual[2].Name)
			assert.Equal(t, shared.NotDeletable, actual[2].State)
		})

		t.Run("not deletable when HEAD is not main worktree", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := conn.Setup(ctrl).
				GetBranchNames("issue1_@issue2", nil, nil)
			setupDefault(s)
			remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, scan)

			actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, scan, false)

			assert.Equal(t, 2, len(actual))
			assert.Equal(t, "issue1", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "issue2", actual[1].Name)
			assert.Equal(t, shared.NotDeletable, actual[1].State)
		})
	})
}

// issue1's head commit is the same as main's, so no distinct PR is found and
// the branch stays NotDeletable.
func Test_BranchIsNotDeletableWhenFirstCommitOfTopicBranchIsAssociatedWithDefaultBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "main"},
		}, nil, nil).
		GetPullRequests("notFound", nil, nil).
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
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Quick)

	actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Quick, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchesAndPRsAreNotAssociatedWhenManyLocalCommitsAreAhead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"},
			{BranchName: "issue1", Filename: "issue1ManyCommits"}, // return with '--max-count=3'
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "62d5d8280031f607f1db058da959a97f6a8e6d90", Filename: "issue1"},
			{Oid: "b8a2645298053fb62ea03e27feea6c483d3fd27e", Filename: "issue1"},
			{Oid: "d787669ee4a103fe0b361fe31c10ea037c72f27c", Filename: "issue1"},
		}, nil, nil).
		GetPullRequests("notFound", nil, nil).
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
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, []shared.PullRequest{}, actual[0].PullRequests)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_NoCommitHistoryWhenFirstCommitOfTopicBranchIsAssociatedWithDefaultBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "main"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("notFound", nil, nil).
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
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, []string{}, actual[0].Commits)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_NoCommitHistoryWhenDetachedBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("main_@detached", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("notFound", nil, nil).
		GetUncommittedChanges([]conn.UncommittedChangeStub{
			{Path: "", Output: ""},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_basic", Output: ""},
		}, nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
			{Key: "branch.main.merge", Filename: "mergeMain"},
			{Key: "branch.main.gh-poi-locked", Filename: "empty"},
			{Key: "branch.main.gh-poi-protected", Filename: "empty"},
			{Key: "branch.(HEAD detached at a97e963).gh-poi-locked", Filename: "empty"},
			{Key: "branch.(HEAD detached at a97e963).gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	actual, _ := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "(HEAD detached at a97e963)", actual[0].Name)
	assert.Equal(t, []string{}, actual[0].Commits)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_ReturnsErrorWhenGetRemoteNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", ErrCommand, nil)

	_, err := GetPreferredRemotes(context.Background(), s.Conn, shared.Quick)

	assert.NotNil(t, err)
}

func Test_DoesNotReturnErrorWhenGetSshConfigFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", ErrCommand, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
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
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.Nil(t, err)
}

func Test_ReturnsErrorWhenGetRepoNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetBranchNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetMergedBranchNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetLogFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
			{Key: "branch.main.gh-poi-locked", Filename: "empty"},
			{Key: "branch.main.gh-poi-protected", Filename: "empty"},
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetAssociatedRefNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
			{Key: "branch.main.gh-poi-locked", Filename: "empty"},
			{Key: "branch.main.gh-poi-protected", Filename: "empty"},
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetPullRequestsFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", ErrCommand, nil).
		GetUncommittedChanges([]conn.UncommittedChangeStub{
			{Path: "", Output: ""},
		}, nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
			{Key: "branch.main.gh-poi-locked", Filename: "empty"},
			{Key: "branch.main.gh-poi-protected", Filename: "empty"},
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetUncommittedChangesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges([]conn.UncommittedChangeStub{
			{Path: "", Output: ""},
		}, ErrCommand, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
			{Key: "branch.main.merge", Filename: "mergeMain"},
			{Key: "branch.main.gh-poi-locked", Filename: "empty"},
			{Key: "branch.main.gh-poi-protected", Filename: "empty"},
			{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenCheckoutBranchFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames([]conn.RepoNamesStub{
			{RepoName: "owner/repo", Filename: "origin"},
		}, nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges([]conn.UncommittedChangeStub{
			{Path: "", Output: ""},
			{Path: "/home/runner/work/gh-poi/gh-poi/conn/fixtures/repo_basic", Output: ""},
		}, nil, nil).
		GetWorktrees("none", nil, nil).
		CheckoutBranch(ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{Key: "remote.origin.gh-resolved", Filename: "empty"},
			{Key: "branch.main.merge", Filename: "mergeMain"},
			{Key: "branch.main.gh-poi-locked", Filename: "empty"},
			{Key: "branch.main.gh-poi-protected", Filename: "empty"},
			{Key: "branch.issue1.merge", Filename: "mergeIssue1"},
			{Key: "branch.issue1.remote", Filename: "remote"},
			{Key: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{Key: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remotes, _ := GetPreferredRemotes(context.Background(), s.Conn, shared.Deep)

	_, err := GetBranches(context.Background(), remotes, s.Conn, shared.Merged, shared.Deep, false)

	assert.NotNil(t, err)
}

func Test_DeleteBranches(t *testing.T) {
	t.Run("delete deletable branches", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := conn.Setup(ctrl).
			GetBranchNames("@main", nil, nil).
			DeleteBranches(nil, conn.NewConf(&conn.Times{N: 1}))

		branches := []shared.Branch{
			{Head: false, Name: "issue1", IsMerged: false, IsLocked: false, Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.Deletable},
			{Head: true, Name: "main", IsMerged: true, IsLocked: false, Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.NotDeletable},
		}

		actual, _ := DeleteBranches(context.Background(), branches, s.Conn)

		assert.Equal(t, 2, len(actual))
		assert.Equal(t, "issue1", actual[0].Name)
		assert.Equal(t, shared.Deleted, actual[0].State)
		assert.Equal(t, "main", actual[1].Name)
		assert.Equal(t, shared.NotDeletable, actual[1].State)
	})

	t.Run("does not delete not deletable branches", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		s := conn.Setup(ctrl).
			DeleteBranches(nil, conn.NewConf(&conn.Times{N: 0}))

		branches := []shared.Branch{
			{Head: false, Name: "issue1", IsMerged: false, IsLocked: false, Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.NotDeletable},
			{Head: true, Name: "main", IsMerged: true, IsLocked: false, Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.NotDeletable},
		}

		actual, _ := DeleteBranches(context.Background(), branches, s.Conn)

		assert.Equal(t, 2, len(actual))
		assert.Equal(t, "issue1", actual[0].Name)
		assert.Equal(t, shared.NotDeletable, actual[0].State)
		assert.Equal(t, "main", actual[1].Name)
		assert.Equal(t, shared.NotDeletable, actual[1].State)
	})
}
