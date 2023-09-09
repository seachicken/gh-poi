package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateRemoteWithScpLikeUrl(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com",
			RepoName: "org/repo",
		},
		NewRemote("origin	git@github.com:org/repo (fetch)"),
	)
}

func Test_CreateRemoteWithScpLikeUrlAndCustomUserinfo(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com",
			RepoName: "org/repo",
		},
		NewRemote("origin	git0-._~@github.com:org/repo (fetch)"),
	)
}

func Test_CreateRemoteWithScpLikeUrlWithoutUserinfo(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com",
			RepoName: "org/repo",
		},
		NewRemote("origin	github.com:org/repo.git (fetch)"),
	)
}

func Test_CreateRemoteWithHttps(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com",
			RepoName: "org/repo",
		},
		NewRemote("origin	https://github.com/org/repo.git (fetch)"),
	)
}

// https://github.com/seachicken/gh-poi/issues/39
func Test_CreateRemoteWithCustomHostname(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com-work",
			RepoName: "org/repo",
		},
		NewRemote("origin	git@github.com-work:org/repo.git (fetch)"),
	)
}
