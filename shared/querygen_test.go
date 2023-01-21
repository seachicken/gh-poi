package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetQueryOrgs(t *testing.T) {
	assert.Equal(t,
		"org:parent-owner org:owner",
		GetQueryOrgs([]string{"parent-owner/repo", "owner/repo"}),
	)
}

func Test_GetQueryRepos(t *testing.T) {
	assert.Equal(t,
		"repo:parent-owner/repo repo:owner/repo",
		GetQueryRepos([]string{"parent-owner/repo", "owner/repo"}),
	)
}

func Test_GetQueryHashesWithCommitOid(t *testing.T) {
	assert.Equal(t,
		[]string{
			"hash:356a192b7913b04c54574d18c28d46e6395428ab " +
				"hash:08a2aaaadff191eb76974b9b3d8b71f202c0156e " +
				"hash:77de68daecd823babbb58edb1c8e14d7106e83bb " +
				"hash:1b6453892473a467d07372d45eb05abc2031647a " +
				"hash:ac3478d69a3c81fa62e60f5c3696165a4e5e6ac4 ",
			"hash:c1dfd96eea8cc2b62785275bca38ac261256e278",
		},
		GetQueryHashes([]Branch{
			{Head: false, Name: "main", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "",
				Commits:       []string{},
				PullRequests:  []PullRequest{}, State: Unknown,
			},
			{Head: true, Name: "issue1", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "",
				Commits: []string{
					"356a192b7913b04c54574d18c28d46e6395428ab",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue2", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "",
				Commits: []string{
					"da4b9237bacccdf19c0760cab7aec4a8359010b0",
					"08a2aaaadff191eb76974b9b3d8b71f202c0156e",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue3", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "",
				Commits: []string{
					"77de68daecd823babbb58edb1c8e14d7106e83bb",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue4", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "",
				Commits: []string{
					"1b6453892473a467d07372d45eb05abc2031647a",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue5", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "",
				Commits: []string{
					"ac3478d69a3c81fa62e60f5c3696165a4e5e6ac4",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue6", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "",
				Commits: []string{
					"c1dfd96eea8cc2b62785275bca38ac261256e278",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
		}),
	)
}

func Test_GetQueryHashesWithRemoteOid(t *testing.T) {
	assert.Equal(t,
		[]string{
			"hash:356a192b7913b04c54574d18c28d46e6395428ab",
		},
		GetQueryHashes([]Branch{
			{Head: true, Name: "issue1", IsMerged: false, IsProtected: false,
				RemoteHeadOid: "356a192b7913b04c54574d18c28d46e6395428ab",
				Commits: []string{
					"da4b9237bacccdf19c0760cab7aec4a8359010b0",
					"08a2aaaadff191eb76974b9b3d8b71f202c0156e",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
		}),
	)
}
