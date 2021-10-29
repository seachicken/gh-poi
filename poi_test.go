package main

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/seachicken/gh-poi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := mocks.NewMockConnection(ctrl)
	conn.
		EXPECT().
		GetRemoteName().
		Return(`origin  git@github.com:owner/name.git (fetch)
origin  git@github.com:owner/name.git (push)
`).
		AnyTimes()
	conn.
		EXPECT().
		GetBrancheNames().
		Return(` ,issue1,356a192b7913b04c54574d18c28d46e6395428ab
*,main,b6589fc6ab0dc82cf12099d1c2d40ab994e8410c
`).
		AnyTimes()
	conn.
		EXPECT().
		FetchRepoNames().
		Return("owner/name").
		AnyTimes()
	conn.
		EXPECT().
		FetchPrStates(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(`{
  "data": {
    "search": {
      "issueCount": 1,
      "edges": [
        {
          "node": {
            "number": 1,
            "url": "https://github.com/owner/name/pull/1",
            "state": "MERGED",
            "headRefName": "issue1",
            "author": {
              "login": "owner"
            }
          }
        }
      ]
    }
  }
}`).
		AnyTimes()

	actual, _ := GetBranches(conn)

	assert.Equal(t, []Branch{
		{false, "issue1", "356a192b7913b04c54574d18c28d46e6395428ab",
			[]PullRequest{
				PullRequest{"issue1", Merged, 1, "https://github.com/owner/name/pull/1", "owner"},
			},
			Deletable,
		},
		{true, "main", "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func TestGetBranches_pullRequestsAreNotMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := mocks.NewMockConnection(ctrl)
	conn.
		EXPECT().
		GetRemoteName().
		Return(`origin  git@github.com:owner/name.git (fetch)
origin  git@github.com:owner/name.git (push)
`).
		AnyTimes()
	conn.
		EXPECT().
		GetBrancheNames().
		Return(` ,issue1,356a192b7913b04c54574d18c28d46e6395428ab
*,main,b6589fc6ab0dc82cf12099d1c2d40ab994e8410c
`).
		AnyTimes()
	conn.
		EXPECT().
		FetchRepoNames().
		Return("owner/name").
		AnyTimes()
	conn.
		EXPECT().
		FetchPrStates(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(`{
  "data": {
    "search": {
      "issueCount": 2,
      "edges": [
        {
          "node": {
            "number": 1,
            "url": "https://github.com/owner/name/pull/1",
            "state": "CLOSED",
            "headRefName": "issue1",
            "author": {
              "login": "owner"
            }
          }
        },
        {
          "node": {
            "number": 2,
            "url": "https://github.com/owner/name/pull/2",
            "state": "CLOSED",
            "headRefName": "issue1",
            "author": {
              "login": "owner"
            }
          }
        }
      ]
    }
  }
}`).
		AnyTimes()

	actual, _ := GetBranches(conn)

	assert.Equal(t, []Branch{
		{false, "issue1", "356a192b7913b04c54574d18c28d46e6395428ab",
			[]PullRequest{
				PullRequest{"issue1", Closed, 1, "https://github.com/owner/name/pull/1", "owner"},
				PullRequest{"issue1", Closed, 2, "https://github.com/owner/name/pull/2", "owner"},
			},
			NotDeletable,
		},
		{true, "main", "b6589fc6ab0dc82cf12099d1c2d40ab994e8410c",
			[]PullRequest{},
			NotDeletable,
		},
	}, actual)
}

func TestDeleteBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := mocks.NewMockConnection(ctrl)
	conn.
		EXPECT().
		GetBrancheNames().
		Return(`*,issue1,356a192b7913b04c54574d18c28d46e6395428ab
 ,main,b6589fc6ab0dc82cf12099d1c2d40ab994e8410c
`).
		AnyTimes()
	conn.
		EXPECT().
		DeleteBranches(gomock.Any()).
		Return("").
		Times(1)

	branches := []Branch{
		{false, "issue1", "", []PullRequest{}, Deletable},
		{false, "issue2", "", []PullRequest{}, Deletable},
		{true, "main", "", []PullRequest{}, NotDeletable},
	}

	expected := []Branch{
		{false, "issue1", "", []PullRequest{}, Deletable},
		{false, "issue2", "", []PullRequest{}, Deleted},
		{true, "main", "", []PullRequest{}, NotDeletable},
	}
	assert.Equal(t, expected, DeleteBranches(branches, conn))
}

func TestDeleteBranches_doesNotExistsDeletableBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	conn := mocks.NewMockConnection(ctrl)
	conn.
		EXPECT().
		GetBrancheNames().
		Return(` ,issue1,356a192b7913b04c54574d18c28d46e6395428ab
*,main,b6589fc6ab0dc82cf12099d1c2d40ab994e8410c
`).
		AnyTimes()
	conn.
		EXPECT().
		DeleteBranches(gomock.Any()).
		Return("").
		Times(0)

	branches := []Branch{
		{true, "issue1", "", []PullRequest{}, NotDeletable},
		{false, "main", "", []PullRequest{}, NotDeletable},
	}

	DeleteBranches(branches, conn)
}
