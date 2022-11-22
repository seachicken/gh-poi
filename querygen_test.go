package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetQueryOrgs(t *testing.T) {
	assert.Equal(t,
		"org:parent-owner org:owner",
		getQueryOrgs([]string{"parent-owner/repo", "owner/repo"}),
	)
}

func Test_GetQueryRepos(t *testing.T) {
	assert.Equal(t,
		"repo:parent-owner/repo repo:owner/repo",
		getQueryRepos([]string{"parent-owner/repo", "owner/repo"}),
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
		getQueryHashes([]Branch{
			{false, "main", false,
				"",
				[]string{},
				[]PullRequest{}, Unknown,
			},
			{true, "issue1", false,
				"",
				[]string{
					"356a192b7913b04c54574d18c28d46e6395428ab",
				},
				[]PullRequest{}, Unknown,
			},
			{false, "issue2", false,
				"",
				[]string{
					"da4b9237bacccdf19c0760cab7aec4a8359010b0",
					"08a2aaaadff191eb76974b9b3d8b71f202c0156e",
				},
				[]PullRequest{}, Unknown,
			},
			{false, "issue3", false,
				"",
				[]string{
					"77de68daecd823babbb58edb1c8e14d7106e83bb",
				},
				[]PullRequest{}, Unknown,
			},
			{false, "issue4", false,
				"",
				[]string{
					"1b6453892473a467d07372d45eb05abc2031647a",
				},
				[]PullRequest{}, Unknown,
			},
			{false, "issue5", false,
				"",
				[]string{
					"ac3478d69a3c81fa62e60f5c3696165a4e5e6ac4",
				},
				[]PullRequest{}, Unknown,
			},
			{false, "issue6", false,
				"",
				[]string{
					"c1dfd96eea8cc2b62785275bca38ac261256e278",
				},
				[]PullRequest{}, Unknown,
			},
		}),
	)
}

func Test_GetQueryHashesWithRemoteOid(t *testing.T) {
	assert.Equal(t,
		[]string{
			"hash:356a192b7913b04c54574d18c28d46e6395428ab",
		},
		getQueryHashes([]Branch{
			{true, "issue1", false,
				"356a192b7913b04c54574d18c28d46e6395428ab",
				[]string{
					"da4b9237bacccdf19c0760cab7aec4a8359010b0",
					"08a2aaaadff191eb76974b9b3d8b71f202c0156e",
				},
				[]PullRequest{}, Unknown,
			},
		}),
	)
}
