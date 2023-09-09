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

var regex = regexp.MustCompile(`^(?:\w+://|\w+@)?([a-zA-Z0-9\.-]+)[:/](.+?/.+?)(?:\.git|)$`)

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
