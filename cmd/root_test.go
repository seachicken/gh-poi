package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/seachicken/gh-poi/conn"
	"github.com/seachicken/gh-poi/shared"
	"github.com/stretchr/testify/assert"
)

var ErrCommand = errors.New("failed to run external command")

func Test_BranchIsDeletableWhenRemoteBranchesAssociatedWithMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main_issue1", nil, nil).
		GetRemoteHeadOid([]conn.RemoteHeadStub{
			{BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "issue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenLsRemoteBranchesAssociatedWithMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main_issue1", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid([]conn.LsRemoteHeadStub{
			{BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "issue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenBranchesAssociatedWithMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main_issue1", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "issue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "b8a2645298053fb62ea03e27feea6c483d3fd27e", Filename: "main_issue1"},
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "main_issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenBranchesAssociatedWithSquashAndMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenBranchesAssociatedWithUpstreamSquashAndMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin_upstream", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1UpMerged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenPRCheckoutBranchesAssociatedWithUpstreamSquashAndMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin_upstream", nil, nil).
		GetBranchNames("@main_forkMain", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "fork/main", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "forkMain"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_forkMain"},
		}, nil, nil).
		GetPullRequests("forkMainUpMerged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.fork/main.merge", Filename: "mergeForkMain"},
			{BranchName: "branch.fork/main.remote", Filename: "remote"},
			{BranchName: "branch.fork/main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.fork/main.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "fork/main", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenBranchIsCheckedOutWithCheckIsFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, conn.NewConf(&conn.Times{N: 1}))
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenBranchIsCheckedOutWithCheckIsTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, conn.NewConf(&conn.Times{N: 0}))
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, true)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenBranchIsCheckedOutWithoutDefaultBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@issue1", nil, nil).
		GetMergedBranchNames("empty", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "issue1_originMain"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsNotDeletableWhenBranchHasModifiedUncommittedChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges(" M README.md", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenBranchHasUntrackedUncommittedChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("?? new.txt", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsNotDeletableWhenPRIsClosedAndStateOptionIsMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Closed", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenPRIsClosedAndStateOptionIsClosed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Closed", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Closed, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenPRHasMergedAndClosedAndStateOptionIsMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged_issue1Closed", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWhenPRHasMergedAndClosedAndStateOptionIsClosed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged_issue1Closed", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Closed, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsNotDeletableWhenBranchesAssociatedWithNotFullyMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
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
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsNotDeletableWhenDefaultBranchAssociatedWithMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("mainMerged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsNotDeletableWhenBranchIsLocked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main_issue1", nil, nil).
		GetRemoteHeadOid([]conn.RemoteHeadStub{
			{BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "issue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "locked"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, true, actual[0].IsLocked)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, false, actual[1].IsLocked)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

// TODO: Remove after deprecated commands are removed
func Test_BranchIsNotDeletableWhenBranchIsLockedForCompatibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main_issue1", nil, nil).
		GetRemoteHeadOid([]conn.RemoteHeadStub{
			{BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "issue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "locked"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, true, actual[0].IsLocked)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, false, actual[1].IsLocked)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsDeletableWithBaseWorktreeCheckedOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@linkedIssue1", nil, nil).
		GetMergedBranchNames("main_@linkedIssue1", nil, nil).
		GetRemoteHeadOid([]conn.RemoteHeadStub{
			{BranchName: "linkedIssue1", Filename: "issue1"},
		}, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "linkedIssue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetPullRequests("linkedIssue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("@linkedIssue1", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.linkedIssue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.linkedIssue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.linkedIssue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "linkedIssue1", actual[0].Name)
	assert.Equal(t, shared.Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsNotDeletableWithLinkedWorktreeCheckedOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@linkedIssue1", nil, nil).
		GetMergedBranchNames("main_@linkedIssue1", nil, nil).
		GetRemoteHeadOid([]conn.RemoteHeadStub{
			{BranchName: "linkedIssue1", Filename: "issue1"},
		}, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "linkedIssue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetPullRequests("linkedIssue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("@main_+linkedIssue1", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.linkedIssue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.linkedIssue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.linkedIssue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "linkedIssue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchIsNotDeletableWithLockedWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_linkedIssue1", nil, nil).
		GetMergedBranchNames("@main_linkedIssue1", nil, nil).
		GetRemoteHeadOid([]conn.RemoteHeadStub{
			{BranchName: "linkedIssue1", Filename: "issue1"},
		}, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main_issue1Merged"}, {BranchName: "linkedIssue1", Filename: "issue1Merged"},
		}, nil, nil).
		GetPullRequests("linkedIssue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("locked", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.linkedIssue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.linkedIssue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.linkedIssue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "linkedIssue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_BranchesAndPRsAreNotAssociatedWhenManyLocalCommitsAreAhead(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
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
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

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
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "main"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("notFound", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

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
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@detached", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("notFound", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.(HEAD detached at a97e963).gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.(HEAD detached at a97e963).gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

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

	_, err := GetRemote(context.Background(), s.Conn)

	assert.NotNil(t, err)
}

func Test_DoesNotReturnsErrorWhenGetSshConfigFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", ErrCommand, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.Nil(t, err)
}

func Test_ReturnsErrorWhenGetRepoNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", ErrCommand, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenCheckReposFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(ErrCommand, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetBranchNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", ErrCommand, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetMergedBranchNames(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", ErrCommand, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetLogFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetAssociatedRefNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetPullRequestsFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", ErrCommand, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenGetUncommittedChangesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetMergedBranchNames("@main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_ReturnsErrorWhenCheckoutBranchFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetMergedBranchNames("main", nil, nil).
		GetRemoteHeadOid(nil, ErrCommand, nil).
		GetLsRemoteHeadOid(nil, nil, nil).
		GetLog([]conn.LogStub{
			{BranchName: "main", Filename: "main"}, {BranchName: "issue1", Filename: "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{Oid: "a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", Filename: "issue1"},
			{Oid: "6ebe3d30d23531af56bd23b5a098d3ccae2a534a", Filename: "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		GetWorktrees("none", nil, nil).
		CheckoutBranch(ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-locked", Filename: "empty"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, shared.Merged, false)

	assert.NotNil(t, err)
}

func Test_DeletingDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetBranchNames("@main", nil, nil).
		DeleteBranches(nil, conn.NewConf(&conn.Times{N: 1}))

	branches := []shared.Branch{
		{Head: false, Name: "issue1", IsMerged: false, IsLocked: false, RemoteHeadOid: "", Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.Deletable},
		{Head: true, Name: "main", IsMerged: true, IsLocked: false, RemoteHeadOid: "", Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.NotDeletable},
	}

	actual, _ := DeleteBranches(context.Background(), branches, s.Conn)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.Deleted, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}

func Test_DoNotDeleteNotDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		DeleteBranches(nil, conn.NewConf(&conn.Times{N: 0}))

	branches := []shared.Branch{
		{Head: false, Name: "issue1", IsMerged: false, IsLocked: false, RemoteHeadOid: "", Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.NotDeletable},
		{Head: true, Name: "main", IsMerged: true, IsLocked: false, RemoteHeadOid: "", Commits: []string{}, PullRequests: []shared.PullRequest{}, State: shared.NotDeletable},
	}

	actual, _ := DeleteBranches(context.Background(), branches, s.Conn)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, shared.NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, shared.NotDeletable, actual[1].State)
}
