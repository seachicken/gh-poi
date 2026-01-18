package shared

import (
	"fmt"
	"strings"
)

func GetQueryOrgs(repoNames []string) string {
	var repos strings.Builder
	for _, name := range repoNames {
		fmt.Fprintf(&repos, "org:%s ", strings.Split(name, "/")[0])
	}
	return strings.TrimSpace(repos.String())
}

func GetQueryRepos(repoNames []string) string {
	var repos strings.Builder
	for _, name := range repoNames {
		fmt.Fprintf(&repos, "repo:%s ", name)
	}
	return strings.TrimSpace(repos.String())
}

func GetQueryHashes(branches []Branch) []string {
	results := []string{}

	var hashes strings.Builder
	for i, branch := range branches {
		if branch.RemoteHeadOid == "" && len(branch.Commits) == 0 {
			continue
		}

		separator := " "
		if i == len(branches)-1 {
			separator = ""
		}
		oid := ""
		if branch.RemoteHeadOid == "" {
			oid = branch.Commits[len(branch.Commits)-1]
		} else {
			oid = branch.RemoteHeadOid
		}
		hash := fmt.Sprintf("hash:%s%s", oid, separator)

		// https://docs.github.com/en/rest/reference/search#limitations-on-query-length
		if len(hashes.String())+len(hash) > 256 {
			results = append(results, hashes.String())
			hashes.Reset()
		}

		hashes.WriteString(hash)
	}
	if len(hashes.String()) > 0 {
		results = append(results, hashes.String())
	}

	return results
}
