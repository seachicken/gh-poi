package shared

import (
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

// https://datatracker.ietf.org/doc/html/rfc3986#section-2.3
var regex = regexp.MustCompile(`^(?:(?:[a-zA-Z0-9-._~]+)(?:://|@))?([a-zA-Z0-9-._~]+)[:/](.+?/.+?)(?:\.git|)$`)

func NewRemote(remoteConfig string) Remote {
	splitConfig := strings.Fields(remoteConfig)
	if len(splitConfig) == 3 {
		found := regex.FindStringSubmatch(splitConfig[1])
		if len(found) == 3 {
			return Remote{splitConfig[0], found[1], found[2]}
		}
	}
	return Remote{}
}
