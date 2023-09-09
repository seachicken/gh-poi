package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRemoteWithScpLikeUrl(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com",
			RepoName: "org/repo",
		},
		NewRemote("origin	git@github.com:org/repo (fetch)"),
	)
}

// https://github.com/seachicken/gh-poi/issues/39
func Test_NewRemoteWithCustomHostname(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com-work",
			RepoName: "org/repo",
		},
		NewRemote("origin	git@github.com-work:org/repo.git (fetch)"),
	)
}

func Test_NewRemoteWithHttps(t *testing.T) {
	assert.Equal(t,
		Remote{
			Name:     "origin",
			Hostname: "github.com",
			RepoName: "org/repo",
		},
		NewRemote("origin	https://github.com/org/repo (fetch)"),
	)
}
