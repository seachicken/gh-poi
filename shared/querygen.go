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
		if len(branch.Commits) == 0 {
			continue
		}

		separator := " "
		if i == len(branches)-1 {
			separator = ""
		}
		hash := fmt.Sprintf("hash:%s%s", branch.Commits[len(branch.Commits)-1], separator)

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

func GetQueryHeads(branches []Branch) []string {
	results := []string{}

	names := make([]string, 0, len(branches))
	for _, branch := range branches {
		if len(branch.Commits) == 0 || strings.Contains(branch.Name, `"`) {
			continue
		}
		// head:"name" is 7 characters plus the name. A longer name cannot fit
		// in one search query; emitting it would exceed the length limit, and
		// flushing an empty buffer would send an unqualified is:pr search.
		if len(branch.Name)+len(`head:""`) > 256 {
			continue
		}
		names = append(names, branch.Name)
	}

	var heads strings.Builder
	for i, name := range names {
		separator := " "
		if i == len(names)-1 {
			separator = ""
		}
		fragment := fmt.Sprintf("head:\"%s\"", name)
		head := fragment + separator

		// https://docs.github.com/en/rest/reference/search#limitations-on-query-length
		if heads.Len()+len(head) > 256 {
			if heads.Len() > 0 {
				results = append(results, heads.String())
				heads.Reset()
			}
			// The trailing separator is the only reason this fragment does not
			// fit. A name this long has to be its own query.
			if len(head) > 256 {
				head = fragment
			}
		}

		heads.WriteString(head)
	}
	if len(heads.String()) > 0 {
		results = append(results, heads.String())
	}

	return results
}
