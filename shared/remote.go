package shared

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type (
	Remote struct {
		Name     string
		Hostname string
		RepoName string
	}
)

var (
	hasSchemePattern  = regexp.MustCompile("^[^:]+://")
	scpLikeURLPattern = regexp.MustCompile("^([^@]+@)?([^:]+):(/?.+)$")
)

// NewRemote parses the result of `git remote -v` and returns a Remote struct.
//
// acceptable url formats:
//
//	ssh://[user@]host.xz[:port]/path/to/repo.git/
//	git://host.xz[:port]/path/to/repo.git/
//	http[s]://host.xz[:port]/path/to/repo.git/
//	ftp[s]://host.xz[:port]/path/to/repo.git/
//
// An alternative scp-like syntax may also be used with the ssh protocol:
//
//	[user@]host.xz:path/to/repo.git/
//
// ref. http://git-scm.com/docs/git-fetch#_git_urls
// the code is heavily inspired by https://github.com/x-motemen/ghq/blob/7163e61e2309a039241ad40b4a25bea35671ea6f/url.go
func NewRemote(remoteConfig string) Remote {
	splitConfig := strings.Fields(remoteConfig)
	if len(splitConfig) != 3 {
		return Remote{}
	}

	ref := splitConfig[1]
	if !hasSchemePattern.MatchString(ref) {
		if scpLikeURLPattern.MatchString(ref) {
			matched := scpLikeURLPattern.FindStringSubmatch(ref)
			user := matched[1]
			host := matched[2]
			path := matched[3]
			ref = fmt.Sprintf("ssh://%s%s/%s", user, host, strings.TrimPrefix(path, "/"))
		}
	}
	u, err := url.Parse(ref)
	if err != nil {
		return Remote{}
	}

	repo := u.Path
	repo = strings.TrimPrefix(repo, "/")
	repo = strings.TrimSuffix(repo, ".git")

	return Remote{
		Name:     splitConfig[0],
		Hostname: u.Host,
		RepoName: repo,
	}
}
