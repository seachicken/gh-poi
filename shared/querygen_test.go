package shared

import (
	"strings"
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
			{Head: false, Name: "main", IsMerged: false, IsLocked: false,
				Commits:      []string{},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: true, Name: "issue1", IsMerged: false, IsLocked: false,
				Commits: []string{
					"356a192b7913b04c54574d18c28d46e6395428ab",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue2", IsMerged: false, IsLocked: false,
				Commits: []string{
					"da4b9237bacccdf19c0760cab7aec4a8359010b0",
					"08a2aaaadff191eb76974b9b3d8b71f202c0156e",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue3", IsMerged: false, IsLocked: false,
				Commits: []string{
					"77de68daecd823babbb58edb1c8e14d7106e83bb",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue4", IsMerged: false, IsLocked: false,
				Commits: []string{
					"1b6453892473a467d07372d45eb05abc2031647a",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue5", IsMerged: false, IsLocked: false,
				Commits: []string{
					"ac3478d69a3c81fa62e60f5c3696165a4e5e6ac4",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
			{Head: false, Name: "issue6", IsMerged: false, IsLocked: false,
				Commits: []string{
					"c1dfd96eea8cc2b62785275bca38ac261256e278",
				},
				PullRequests: []PullRequest{}, State: Unknown,
			},
		}),
	)
}

func Test_GetQueryHeads(t *testing.T) {
	// Emitted names only. The third one is the fragment that passes 256 characters.
	second := strings.Repeat("b", 110)
	third := strings.Repeat("c", 110)

	assert.Equal(t,
		[]string{
			`head:"feature/stack-1" head:"` + second + `" `,
			`head:"` + third + `"`,
		},
		GetQueryHeads([]Branch{
			{Name: "main", Commits: []string{}},
			{Name: "feature/stack-1", Commits: []string{"aaaaaaaa"}},
			{Name: strings.Repeat("x", 239) + `"`, Commits: []string{"bbbbbbbb"}},
			{Name: second, Commits: []string{"cccccccc"}},
			{Name: third, Commits: []string{"dddddddd"}},
		}),
	)
}

func Test_GetQueryHeadsSkipsNameLongerThanSearchLimit(t *testing.T) {
	// head:"<250 chars>" is 257 characters and cannot be searched on its own.
	longName := strings.Repeat("n", 250)
	assert.Equal(t,
		[]string{},
		GetQueryHeads([]Branch{
			{Name: longName, Commits: []string{"aaaaaaaa"}},
		}),
	)

	// A leading over-long name must not flush an empty, unqualified query.
	assert.Equal(t,
		[]string{`head:"feature/stack-1"`},
		GetQueryHeads([]Branch{
			{Name: longName, Commits: []string{"aaaaaaaa"}},
			{Name: "feature/stack-1", Commits: []string{"bbbbbbbb"}},
		}),
	)

	// 249 characters is exactly 256 with the head:"" wrapper, so it is sent.
	name249 := strings.Repeat("m", 249)
	exact := `head:"` + name249 + `"`
	assert.Equal(t, 256, len(exact))
	assert.Equal(t,
		[]string{exact},
		GetQueryHeads([]Branch{
			{Name: name249, Commits: []string{"aaaaaaaa"}},
		}),
	)
	// The trailing separator would make it 257, so it is its own query.
	assert.Equal(t,
		[]string{
			exact,
			`head:"feature/stack-1"`,
		},
		GetQueryHeads([]Branch{
			{Name: name249, Commits: []string{"aaaaaaaa"}},
			{Name: "feature/stack-1", Commits: []string{"bbbbbbbb"}},
		}),
	)
}
