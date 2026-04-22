package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
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
				FetchBranch(nil, nil).
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
				FetchBranch(nil, nil).
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
				FetchBranch(nil, nil).
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

			assert.Equal(t, 3, len(actual))
			assert.Equal(t, "(HEAD detached at origin/main)", actual[0].Name)
			assert.Equal(t, shared.NotDeletable, actual[0].State)
			assert.Equal(t, "issue1", actual[1].Name)
			assert.Equal(t, shared.Deletable, actual[1].State)
			assert.Equal(t, "issue2", actual[2].Name)
			assert.Equal(t, shared.NotDeletable, actual[2].State)
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

func Test_IsRepoDeletable(t *testing.T) {
	t.Run("returns true when all branches are deletable or deleted", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deletable, IsMerged: true},
			{Name: "bugfix", State: shared.Deleted, IsMerged: true},
		}
		assert.True(t, IsRepoDeletable(branches))
	})

	t.Run("returns true when one deleted and default branch is merged", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deleted, IsMerged: true},
			{Name: "main", State: shared.NotDeletable, IsMerged: true, IsDefault: true},
		}
		assert.True(t, IsRepoDeletable(branches))
	})

	t.Run("returns true when one deletable and only detached HEAD remains", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deletable, IsMerged: true},
			{Name: "(HEAD detached at origin/main)", State: shared.NotDeletable, IsMerged: false},
		}
		assert.True(t, IsRepoDeletable(branches))
	})

	t.Run("returns false when no branches are deletable or deleted", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "main", State: shared.NotDeletable, IsMerged: true},
			{Name: "feature", State: shared.NotDeletable, IsMerged: false},
		}
		assert.False(t, IsRepoDeletable(branches))
	})

	t.Run("returns false when remaining branch is not merged", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deletable, IsMerged: true},
			{Name: "main", State: shared.NotDeletable, IsMerged: false},
		}
		assert.False(t, IsRepoDeletable(branches))
	})

	t.Run("returns false when remaining branch has tracked changes", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deletable, IsMerged: true},
			{Name: "main", State: shared.NotDeletable, IsMerged: true, HasTrackedChanges: true},
		}
		assert.False(t, IsRepoDeletable(branches))
	})

	t.Run("returns true when one deleted and default branch is behind upstream", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deleted, IsMerged: true},
			{Name: "main", State: shared.NotDeletable, IsMerged: true, IsDefault: true},
		}
		assert.True(t, IsRepoDeletable(branches))
	})

	t.Run("returns false when default branch is ahead of upstream", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deletable, IsMerged: true},
			{Name: "main", State: shared.NotDeletable, IsMerged: false, IsDefault: true},
		}
		assert.False(t, IsRepoDeletable(branches))
	})

	t.Run("returns false when locked branch is not merged", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deletable, IsMerged: true},
			{Name: "locked-branch", State: shared.NotDeletable, IsMerged: false, IsLocked: true},
		}
		assert.False(t, IsRepoDeletable(branches))
	})

	t.Run("returns true with mixed deletable and merged remaining branches", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature1", State: shared.Deleted, IsMerged: true},
			{Name: "feature2", State: shared.Deletable, IsMerged: true},
			{Name: "main", State: shared.NotDeletable, IsMerged: true, IsDefault: true},
		}
		assert.True(t, IsRepoDeletable(branches))
	})

	t.Run("returns false with empty branch list", func(t *testing.T) {
		branches := []shared.Branch{}
		assert.False(t, IsRepoDeletable(branches))
	})

	t.Run("skips multiple detached HEADs", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature", State: shared.Deleted, IsMerged: true},
			{Name: "(HEAD detached at origin/main)", State: shared.NotDeletable, IsMerged: false},
			{Name: "(HEAD detached at abc123)", State: shared.NotDeletable, IsMerged: false, HasTrackedChanges: true},
		}
		assert.True(t, IsRepoDeletable(branches))
	})

	t.Run("returns false when one remaining branch not merged among many merged", func(t *testing.T) {
		branches := []shared.Branch{
			{Name: "feature1", State: shared.Deleted, IsMerged: true},
			{Name: "main", State: shared.NotDeletable, IsMerged: true, IsDefault: true},
			{Name: "wip-branch", State: shared.NotDeletable, IsMerged: false},
		}
		assert.False(t, IsRepoDeletable(branches))
	})
}

func Test_FindGitRepos(t *testing.T) {
	t.Run("finds repos in directory tree", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "repo1", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "repo2", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "not-a-repo"), 0755)

		repos, err := FindGitRepos(tmpDir)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(repos))
		slices.Sort(repos)
		assert.Equal(t, filepath.Join(tmpDir, "repo1"), repos[0])
		assert.Equal(t, filepath.Join(tmpDir, "repo2"), repos[1])
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "visible", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, ".hidden", "repo", ".git"), 0755)

		repos, err := FindGitRepos(tmpDir)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(repos))
		assert.Equal(t, filepath.Join(tmpDir, "visible"), repos[0])
	})

	t.Run("skips nested git directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "repo", ".git"), 0755)
		os.MkdirAll(filepath.Join(tmpDir, "repo", "subdir", ".git"), 0755)

		repos, err := FindGitRepos(tmpDir)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(repos))
		assert.Equal(t, filepath.Join(tmpDir, "repo"), repos[0])
	})

	t.Run("returns empty for directory with no repos", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "plain-dir"), 0755)

		repos, err := FindGitRepos(tmpDir)

		assert.Nil(t, err)
		assert.Equal(t, 0, len(repos))
	})

	t.Run("handles deeply nested repos", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.MkdirAll(filepath.Join(tmpDir, "a", "b", "c", "deep-repo", ".git"), 0755)

		repos, err := FindGitRepos(tmpDir)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(repos))
		assert.Equal(t, filepath.Join(tmpDir, "a", "b", "c", "deep-repo"), repos[0])
	})
		t.Run("filters nested repos via post-walk deduplication", func(t *testing.T) {
			tmpDir := t.TempDir()
			os.MkdirAll(filepath.Join(tmpDir, "outer", "inner", ".git"), 0755)
			os.MkdirAll(filepath.Join(tmpDir, "outer", ".git"), 0755)

			repos, err := FindGitRepos(tmpDir)

			assert.Nil(t, err)
			assert.Equal(t, 1, len(repos))
			assert.Equal(t, filepath.Join(tmpDir, "outer"), repos[0])
		})
}

func Test_ParseUserRepos(t *testing.T) {
	t.Run("parses valid JSON response", func(t *testing.T) {
		jsonResp := `[{"nameWithOwner":"owner/repo1"},{"nameWithOwner":"owner/repo2"}]`
		repos, err := ParseUserRepos(jsonResp)
		assert.Nil(t, err)
		assert.Equal(t, []string{"owner/repo1", "owner/repo2"}, repos)
	})

	t.Run("returns empty for empty array", func(t *testing.T) {
		jsonResp := `[]`
		repos, err := ParseUserRepos(jsonResp)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(repos))
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		_, err := ParseUserRepos("not json")
		assert.NotNil(t, err)
	})
}


func Test_ParsePullRequestsList(t *testing.T) {
	t.Run("parses valid JSON response", func(t *testing.T) {
		jsonResp := `[{"number":1,"head":{"ref":"feature"},"user":{"login":"me"},"state":"closed","merged_at":"2024-01-01T00:00:00Z"},{"number":2,"head":{"ref":"open-pr"},"user":{"login":"other"},"state":"open","merged_at":null}]`
		prs, err := ParsePullRequestsList(jsonResp)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(prs))
		assert.Equal(t, 1, prs[0].Number)
		assert.Equal(t, "feature", prs[0].HeadRefName)
		assert.Equal(t, "closed", prs[0].State)
		assert.Equal(t, "2024-01-01T00:00:00Z", prs[0].MergedAt)
		assert.Equal(t, "me", prs[0].Author)
		assert.Equal(t, "open", prs[1].State)
		assert.Equal(t, "", prs[1].MergedAt)
		assert.Equal(t, "other", prs[1].Author)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		_, err := ParsePullRequestsList("not json")
		assert.NotNil(t, err)
	})
}

func Test_IsRepoDeletableAPI(t *testing.T) {
	t.Run("returns true for fork with no open viewer PRs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[{"number":1,"head":{"ref":"feature"},"user":{"login":"me"},"state":"closed","merged_at":"2024-01-01T00:00:00Z"}]`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "", "", true, "", "", "me")
		assert.Nil(t, err)
		assert.True(t, result)
	})

	t.Run("returns false for non-fork even with merged PR", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[{"number":1,"head":{"ref":"feature"},"user":{"login":"me"},"state":"closed","merged_at":"2024-01-01T00:00:00Z"}]`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "", "", false, "", "", "me")
		assert.Nil(t, err)
		assert.False(t, result)
	})

	t.Run("returns false for fork with viewer open PRs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[{"number":1,"head":{"ref":"feature"},"user":{"login":"me"},"state":"open","merged_at":null}]`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "", "", true, "", "", "me")
		assert.Nil(t, err)
		assert.False(t, result)
	})

	t.Run("returns true for fork with no PRs at all", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[]`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "", "", true, "", "", "me")
		assert.Nil(t, err)
		assert.True(t, result)
	})

	t.Run("returns true for fork when only other users have open PRs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[{"number":1,"head":{"ref":"feature"},"user":{"login":"me"},"state":"closed","merged_at":"2024-01-01T00:00:00Z"},{"number":2,"head":{"ref":"other"},"user":{"login":"other"},"state":"open","merged_at":null}]`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "", "", true, "", "", "me")
		assert.Nil(t, err)
		assert.True(t, result)
	})
	t.Run("returns false for fork when viewer has open PR on parent repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		forkPRs := `[]`
		parentPRs := `[{"number":1,"head":{"ref":"feature"},"user":{"login":"me"},"state":"open","merged_at":null}]`

		s.GetRepoPullRequestsList("owner", "repo", forkPRs, nil, nil).
			GetRepoPullRequestsList("jdx", "hk", parentPRs, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "", "master", true, "jdx", "hk", "me")
		assert.Nil(t, err)
		assert.False(t, result)
	})

	t.Run("returns true for fork when viewer PR on parent is merged", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		forkPRs := `[]`
		parentPRs := `[{"number":1,"head":{"ref":"feature"},"user":{"login":"me"},"state":"closed","merged_at":"2024-01-01T00:00:00Z"}]`

		s.GetRepoPullRequestsList("owner", "repo", forkPRs, nil, nil).
			GetRepoPullRequestsList("jdx", "hk", parentPRs, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "", "master", true, "jdx", "hk", "me")
		assert.Nil(t, err)
		assert.True(t, result)
	})
	t.Run("returns false for fork with branch ahead of default", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[]`
		branchesJSON := `[{"name":"main"},{"name":"patch-1"}]`
		compareJSON := `{"ahead_by":1,"behind_by":0,"status":"ahead","commits":[{"author":{"login":"me"}}]}`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil).
			GetRepoBranches("owner", "repo", branchesJSON, nil, nil).
			CompareCommits("owner", "repo", "main", "patch-1", compareJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "main", "", true, "", "", "me")
		assert.Nil(t, err)
		assert.False(t, result)
	})

	t.Run("returns true for fork with no branches ahead", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[]`
		branchesJSON := `[{"name":"main"},{"name":"patch-1"}]`
		compareJSON := `{"ahead_by":0,"behind_by":0,"status":"identical","commits":[]}`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil).
			GetRepoBranches("owner", "repo", branchesJSON, nil, nil).
			CompareCommits("owner", "repo", "main", "patch-1", compareJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "main", "", true, "", "", "me")
		assert.Nil(t, err)
		assert.True(t, result)
	})

	t.Run("returns true for fork with branch ahead by other author", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[]`
		branchesJSON := `[{"name":"main"},{"name":"patch-1"}]`
		compareJSON := `{"ahead_by":1,"behind_by":0,"status":"ahead","commits":[{"author":{"login":"someone-else"}}]}`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil).
			GetRepoBranches("owner", "repo", branchesJSON, nil, nil).
			CompareCommits("owner", "repo", "main", "patch-1", compareJSON, nil, nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "main", "", true, "", "", "me")
		assert.Nil(t, err)
		assert.True(t, result)
	})

	t.Run("returns error when branch list fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[]`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil).
			GetRepoBranches("owner", "repo", "", errors.New("api error"), nil)

		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "main", "", true, "", "", "me")
		assert.NotNil(t, err)
		assert.False(t, result)
	})

	t.Run("compares against parent default branch when available", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		s := conn.Setup(ctrl)

		prsJSON := `[]`
		branchesJSON := `[{"name":"main"},{"name":"patch-1"}]`
		compareJSON := `{"ahead_by":1,"behind_by":0,"status":"ahead","commits":[{"author":{"login":"me"}}]}`

		s.GetRepoPullRequestsList("owner", "repo", prsJSON, nil, nil).
			GetRepoPullRequestsList("parent", "repo", prsJSON, nil, nil).
			GetRepoBranches("owner", "repo", branchesJSON, nil, nil).
			CompareCommits("owner", "repo", "parent:master", "main", compareJSON, nil, nil).
			CompareCommits("owner", "repo", "parent:master", "patch-1", compareJSON, nil, nil)

		// Uses parentOwner:parentDefaultBranch as compare base
		result, err := IsRepoDeletableAPI(context.Background(), s.Conn, "owner", "repo", "main", "master", true, "parent", "repo", "me")
		assert.Nil(t, err)
		assert.False(t, result)
	})
	}


func Test_ParseRepoName(t *testing.T) {
	t.Run("parses owner/repo from single repo name", func(t *testing.T) {
	owner, repo, isFork, _, _ := ParseRepoName([]string{"owner/repo"})
	assert.Equal(t, "owner", owner)
	assert.Equal(t, "repo", repo)
	assert.False(t, isFork)
	})

	t.Run("parses owner/repo and detects fork", func(t *testing.T) {
	owner, repo, isFork, _, _ := ParseRepoName([]string{"owner/repo", "parent/parent-repo"})
	assert.Equal(t, "owner", owner)
	assert.Equal(t, "repo", repo)
	assert.True(t, isFork)
	})

	t.Run("handles empty repo names", func(t *testing.T) {
	owner, repo, isFork, _, _ := ParseRepoName([]string{})
	assert.Equal(t, "", owner)
	assert.Equal(t, "", repo)
	assert.False(t, isFork)
	})
}

func Test_ParseAuthScopes(t *testing.T) {
	t.Run("parses scopes from auth status output", func(t *testing.T) {
		output := "github.com\n  ✓ Logged in to github.com account jhult (keyring)\n  - Token scopes: 'gist', 'read:org', 'repo'\n"
		scopes := ParseAuthScopes(output)
		assert.Equal(t, []string{"gist", "read:org", "repo"}, scopes)
	})

	t.Run("returns nil when no scopes line found", func(t *testing.T) {
		output := "github.com\n  ✓ Logged in to github.com account jhult\n"
		scopes := ParseAuthScopes(output)
		assert.Nil(t, scopes)
	})

	t.Run("returns nil for empty output", func(t *testing.T) {
		scopes := ParseAuthScopes("")
		assert.Nil(t, scopes)
	})
}

func Test_HasDeleteRepoScope(t *testing.T) {
	t.Run("returns true when delete_repo is present", func(t *testing.T) {
		assert.True(t, HasDeleteRepoScope([]string{"gist", "read:org", "repo", "delete_repo"}))
	})

	t.Run("returns false when delete_repo is missing", func(t *testing.T) {
		assert.False(t, HasDeleteRepoScope([]string{"gist", "read:org", "repo"}))
	})

	t.Run("returns false for empty scopes", func(t *testing.T) {
		assert.False(t, HasDeleteRepoScope([]string{}))
	})

	t.Run("returns false for nil scopes", func(t *testing.T) {
		assert.False(t, HasDeleteRepoScope(nil))
	})
}
