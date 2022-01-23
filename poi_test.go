package main

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/seachicken/gh-poi/connmock"
	"github.com/stretchr/testify/assert"
)

func Test_ShouldBeDeletableWhenBranchesAssociatedWithMergedPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"356a192b7913b04c54574d18c28d46e6395428ab",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1, []string{
						"356a192b7913b04c54574d18c28d46e6395428ab",
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

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin_upstream", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, nil, nil).
		GetPullRequests("issue1UpMerged", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"356a192b7913b04c54574d18c28d46e6395428ab",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1, []string{
						"356a192b7913b04c54574d18c28d46e6395428ab",
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

func Test_ShouldNotDeletableWhenBranchIsCheckedOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("main_@issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{
			true, "issue1",
			[]string{
				"356a192b7913b04c54574d18c28d46e6395428ab",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1, []string{
						"356a192b7913b04c54574d18c28d46e6395428ab",
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

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, nil, nil).
		GetPullRequests("issue1Closed", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"356a192b7913b04c54574d18c28d46e6395428ab",
			},
			[]PullRequest{
				{
					"issue1", Closed, 1, []string{
						"356a192b7913b04c54574d18c28d46e6395428ab",
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

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, nil, nil).
		GetPullRequests("issue1Merged_issue1Closed", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"356a192b7913b04c54574d18c28d46e6395428ab",
			},
			[]PullRequest{
				{
					"issue1", Closed, 1, []string{
						"356a192b7913b04c54574d18c28d46e6395428ab",
					},
					"https://github.com/owner/repo/pull/1", "owner",
				},
				{
					"issue1", Merged, 2, []string{
						"356a192b7913b04c54574d18c28d46e6395428ab",
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

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main_issue1Merged"}, {"issue1", "issue1CommitAfterMerge"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"cb197ba87e4ad323b1008c611212deb7da2a4a49", "main"},
				{"b8a2645298053fb62ea03e27feea6c483d3fd27e", "issue1"},
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, nil, nil).
		GetPullRequests("issue1Merged", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{
			false, "issue1",
			[]string{
				"b8a2645298053fb62ea03e27feea6c483d3fd27e",
				"356a192b7913b04c54574d18c28d46e6395428ab",
			},
			[]PullRequest{
				{
					"issue1", Merged, 1, []string{
						"356a192b7913b04c54574d18c28d46e6395428ab",
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

func Test_ReturnsAnErrorWhenGetRepoNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		GetRepoNames("origin", errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenCheckReposFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(errors.New("failed to run external command: gh"), nil).
		GetRepoNames("origin", nil, nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetBranchNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", errors.New("failed to run external command: gh"), nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetLogFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog(
			[]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}},
			errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetAssociatedBranchNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, errors.New("failed to run external command: git"), nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetPullRequestsFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBranchNames("@main_issue1", nil, nil).
		GetLog([]connmock.LogStub{{"main", "main"}, {"issue1", "issue1"}}, nil, nil).
		GetAssociatedBranchNames(
			[]connmock.AssociatedBranchNamesStub{
				{"356a192b7913b04c54574d18c28d46e6395428ab", "issue1"},
				{"b6589fc6ab0dc82cf12099d1c2d40ab994e8410c", "main_issue1"},
			}, nil, nil).
		GetPullRequests("issue1Merged", errors.New("failed to run external command: gh"), nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_DeletingDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		GetBranchNames("@main", nil, nil).
		DeleteBranches(nil, connmock.NewConf(&connmock.Times{N: 1}))

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

	s := connmock.Setup(ctrl).
		DeleteBranches(nil, connmock.NewConf(&connmock.Times{N: 0}))

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
