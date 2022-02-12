package main

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/seachicken/gh-poi/conn"
	"github.com/stretchr/testify/assert"
)

func Test_ShouldBeDeletableWhenBranchesAssociatedWithMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			Deletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldBeDeletableWhenBranchesAssociatedWithUpstreamMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin_upstream", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1UpMerged", nil, nil).
		GetUncommittedChanges("", nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/parent-owner/repo/pull/1", "owner",
				},
			},
			Deletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldBeDeletableWhenBranchIsCheckedOutWithTheCheckIsFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		CheckoutBranch(nil, conn.NewConf(&conn.Times{N: 1}))

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			Deletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldBeDeletableWhenBranchIsCheckedOutWithTheCheckIsTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		CheckoutBranch(nil, conn.NewConf(&conn.Times{N: 0}))

	actual, _ := GetBranches(s.Conn, true)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			Deletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldBeDeletableWhenBranchIsCheckedOutWithoutADefaultBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "issue1_origin-main"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		CheckoutBranch(nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			Deletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldNotDeletableWhenBranchHasUncommittedChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges(" M README.md", nil, nil).
		CheckoutBranch(nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			true, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			NotDeletable,
		},
		{
			false, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldNotDeletableWhenBranchesAssociatedWithClosedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Closed", nil, nil).
		GetUncommittedChanges("", nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Closed, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			NotDeletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldBeDeletableWhenBranchesAssociatedWithMergedAndClosedPRs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged_issue1Closed", nil, nil).
		GetUncommittedChanges("", nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Closed, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
				{
					"issue1", Merged, 2,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/2", "owner",
				},
			},
			Deletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldNotDeletableWhenBranchesAssociatedWithNotFullyMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main_issue1Merged"}, {"issue1", "issue1CommitAfterMerge"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"cb197ba87e4ad323b1008c611212deb7da2a4a49", "main"},
			{"b8a2645298053fb62ea03e27feea6c483d3fd27e", "issue1"},
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"b8a2645298053fb62ea03e27feea6c483d3fd27e",
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1,
					[]string{
						"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			NotDeletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func Test_ShouldNotDeletableWhenDefaultBranchAssociatedWithMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("mainMerged", nil, nil).
		GetUncommittedChanges("", nil, nil)

	actual, _ := GetBranches(s.Conn, false)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0",
			},
			[]PullRequest{},
			NotDeletable,
		},
		{
			true, "main",
			[]string{},
			[]PullRequest{
				{
					"main", Merged, 1,
					[]string{
						"6ebe3d30d23531af56bd23b5a098d3ccae2a534a",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
			},
			NotDeletable,
		},
	}, actual)
}

func Test_ReturnsAnErrorWhenGetRemoteNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetRepoNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenCheckReposFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(errors.New("failed to run external command: gh"), nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetBranchNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", errors.New("failed to run external command: gh"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetLogFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetAssociatedRefNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetPullRequestsFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", errors.New("failed to run external command: gh"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetUncommittedChangesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenCheckoutBranchFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRemoteNames("origin", nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetLog([]conn.LogStub{
			{"main", "main"}, {"issue1", "issue1"},
		}, nil, nil).
		GetAssociatedRefNames([]conn.AssociatedBranchNamesStub{
			{"a97e9630426df5d34ca9ee77ae1159bdfd5ff8f0", "issue1"},
			{"6ebe3d30d23531af56bd23b5a098d3ccae2a534a", "main_issue1"},
		}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil).
		GetUncommittedChanges("", nil, nil).
		CheckoutBranch(errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn, false)

	assert.NotNil(t, err)
}

func Test_DeletingDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		GetBranchNames("@main", nil, nil).
		DeleteBranches(nil, conn.NewConf(&conn.Times{N: 1}))

	branches := []Branch{
		{false, "issue1", []string{}, []PullRequest{}, Deletable},
		{true, "main", []string{}, []PullRequest{}, NotDeletable},
	}

	actual, _ := DeleteBranches(branches, s.Conn)

	expected := []Branch{
		{false, "issue1", []string{}, []PullRequest{}, Deleted},
		{true, "main", []string{}, []PullRequest{}, NotDeletable},
	}
	assert.Equal(t, expected, actual)
}

func Test_DoNotDeleteNotDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := conn.Setup(ctrl).
		DeleteBranches(nil, conn.NewConf(&conn.Times{N: 0}))

	branches := []Branch{
		{false, "issue1", []string{}, []PullRequest{}, NotDeletable},
		{true, "main", []string{}, []PullRequest{}, NotDeletable},
	}

	actual, _ := DeleteBranches(branches, s.Conn)

	expected := []Branch{
		{false, "issue1", []string{}, []PullRequest{}, NotDeletable},
		{true, "main", []string{}, []PullRequest{}, NotDeletable},
	}
	assert.Equal(t, expected, actual)
}
