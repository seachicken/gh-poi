package main

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/seachicken/gh-poi/conn"
	"github.com/stretchr/testify/assert"
)

var ErrCommand = errors.New("failed to run external command")

func Test_ShouldBeDeletableWhenRemoteBranchesAssociatedWithMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenLsRemoteBranchesAssociatedWithMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchesAssociatedWithMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchesAssociatedWithSquashAndMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchesAssociatedWithUpstreamSquashAndMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenPRCheckoutBranchesAssociatedWithUpstreamSquashAndMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.fork/main.merge", Filename: "mergeForkMain"},
			{BranchName: "branch.fork/main.remote", Filename: "remote"},
			{BranchName: "branch.fork/main.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "fork/main", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchIsCheckedOutWithTheCheckIsFalse(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, conn.NewConf(&conn.Times{N: 1}))
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchIsCheckedOutWithTheCheckIsTrue(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, conn.NewConf(&conn.Times{N: 0}))
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, true)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchIsCheckedOutWithoutADefaultBranch(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldNotDeletableWhenBranchHasModifiedUncommittedChanges(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchHasUntrackedUncommittedChanges(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil).
		CheckoutBranch(nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldNotDeletableWhenBranchesAssociatedWithClosedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeDeletableWhenBranchesAssociatedWithSquashAndMergedAndClosedPRs(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldNotDeletableWhenBranchesAssociatedWithNotFullyMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldNotDeletableWhenDefaultBranchAssociatedWithMergedPR(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldNotDeletableWhenBranchIsProtected(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "protected"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, true, actual[0].IsProtected)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, false, actual[1].IsProtected)
	assert.Equal(t, NotDeletable, actual[1].State)
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, []PullRequest{}, actual[0].PullRequests)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeNoCommitHistoryWhenTheFirstCommitOfATopicBranchIsAssociatedWithTheDefaultBranch(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, []string{}, actual[0].Commits)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ShouldBeNoCommitHistoryWhenDetachedBranch(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.(HEAD detached at a97e963).gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	actual, _ := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "(HEAD detached at a97e963)", actual[0].Name)
	assert.Equal(t, []string{}, actual[0].Commits)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_ReturnsAnErrorWhenGetRemoteNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", ErrCommand, nil)

	_, err := GetRemote(context.Background(), s.Conn)

	assert.NotNil(t, err)
}

func Test_DoesNotReturnsAnErrorWhenGetSshConfigFails(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.Nil(t, err)
}

func Test_ReturnsAnErrorWhenGetRepoNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", ErrCommand, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenCheckReposFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(ErrCommand, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetBranchNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetSshConfig("github.com", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", ErrCommand, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetMergedBranchNames(t *testing.T) {
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

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetLogFails(t *testing.T) {
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
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetAssociatedRefNamesFails(t *testing.T) {
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
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetPullRequestsFails(t *testing.T) {
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
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetUncommittedChangesFails(t *testing.T) {
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
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenCheckoutBranchFails(t *testing.T) {
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
		CheckoutBranch(ErrCommand, nil).
		GetConfig([]conn.ConfigStub{
			{BranchName: "branch.main.merge", Filename: "mergeMain"},
			{BranchName: "branch.main.gh-poi-protected", Filename: "empty"},
			{BranchName: "branch.issue1.merge", Filename: "mergeIssue1"},
			{BranchName: "branch.issue1.remote", Filename: "remote"},
			{BranchName: "branch.issue1.gh-poi-protected", Filename: "empty"},
		}, nil, nil)
	remote, _ := GetRemote(context.Background(), s.Conn)

	_, err := GetBranches(context.Background(), remote, s.Conn, false)

	assert.NotNil(t, err)
}

func Test_DeletingDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetBranchNames("@main", nil, nil).
		DeleteBranches(nil, conn.NewConf(&conn.Times{N: 1}))

	branches := []Branch{
		{false, "issue1", false, false, "", []string{}, []PullRequest{}, Deletable},
		{true, "main", true, false, "", []string{}, []PullRequest{}, NotDeletable},
	}

	actual, _ := DeleteBranches(context.Background(), branches, s.Conn)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, Deleted, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}

func Test_DoNotDeleteNotDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		DeleteBranches(nil, conn.NewConf(&conn.Times{N: 0}))

	branches := []Branch{
		{false, "issue1", false, false, "", []string{}, []PullRequest{}, NotDeletable},
		{true, "main", true, false, "", []string{}, []PullRequest{}, NotDeletable},
	}

	actual, _ := DeleteBranches(context.Background(), branches, s.Conn)

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, "issue1", actual[0].Name)
	assert.Equal(t, NotDeletable, actual[0].State)
	assert.Equal(t, "main", actual[1].Name)
	assert.Equal(t, NotDeletable, actual[1].State)
}
