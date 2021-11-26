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
		GetBrancheNames("@main_issue1", nil, nil).
		GetPullRequests("issue1Merged", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{false, "issue1", "356a192b7913b04c54574d18c28d46e6395428ab",
			[]PullRequest{
				{"issue1", Merged, 1, "https://github.com/owner/repo/pull/1", "owner"},
			},
			Deletable,
		},
		{true, "main", "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
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
		GetBrancheNames("@main_issue1", nil, nil).
		GetPullRequests("issue1UpMerged", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{false, "issue1", "356a192b7913b04c54574d18c28d46e6395428ab",
			[]PullRequest{
				{"issue1", Merged, 1, "https://github.com/parent-owner/repo/pull/1", "owner"},
			},
			Deletable,
		},
		{true, "main", "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
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
		GetBrancheNames("@main_issue1", nil, nil).
		GetPullRequests("issue1Closed", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{false, "issue1", "356a192b7913b04c54574d18c28d46e6395428ab",
			[]PullRequest{
				{"issue1", Closed, 1, "https://github.com/owner/repo/pull/1", "owner"},
			},
			NotDeletable,
		},
		{true, "main", "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
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
		GetBrancheNames("@main_issue1", nil, nil).
		GetPullRequests("issue1Merged_issue1Closed", nil, nil)

	actual, _ := GetBranches(s.Conn)

	assert.Equal(t, []Branch{
		{false, "issue1", "356a192b7913b04c54574d18c28d46e6395428ab",
			[]PullRequest{
				{"issue1", Closed, 1, "https://github.com/owner/repo/pull/1", "owner"},
				{"issue1", Merged, 2, "https://github.com/owner/repo/pull/2", "owner"},
			},
			Deletable,
		},
		{true, "main", "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
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

func Test_ReturnsAnErrorWhenGetBrancheNamesFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBrancheNames("@main_issue1", errors.New("failed to run external command: gh"), nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_ReturnsAnErrorWhenGetPullRequestsFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		CheckRepos(nil, nil).
		GetRepoNames("origin", nil, nil).
		GetBrancheNames("@main_issue1", nil, nil).
		GetPullRequests("issue1Merged", errors.New("failed to run external command: gh"), nil)

	_, err := GetBranches(s.Conn)

	assert.NotNil(t, err)
}

func Test_DeletingDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		GetBrancheNames("@main", nil, nil).
		DeleteBranches(nil, connmock.NewConf(&connmock.Times{N: 1}))

	branches := []Branch{
		{false, "issue1", "", []PullRequest{}, Deletable},
		{true, "main", "", []PullRequest{}, NotDeletable},
	}

	actual, _ := DeleteBranches(branches, s.Conn)

	expected := []Branch{
		{false, "issue1", "", []PullRequest{}, Deleted},
		{true, "main", "", []PullRequest{}, NotDeletable},
	}
	assert.Equal(t, expected, actual)
}

func Test_DoNotDeleteNotDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s := connmock.Setup(ctrl).
		DeleteBranches(nil, connmock.NewConf(&connmock.Times{N: 0}))

	branches := []Branch{
		{false, "issue1", "", []PullRequest{}, NotDeletable},
		{true, "main", "", []PullRequest{}, NotDeletable},
	}

	actual, _ := DeleteBranches(branches, s.Conn)

	expected := []Branch{
		{false, "issue1", "", []PullRequest{}, NotDeletable},
		{true, "main", "", []PullRequest{}, NotDeletable},
	}
	assert.Equal(t, expected, actual)
}
