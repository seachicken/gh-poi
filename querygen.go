package main

import (
	"fmt"
	"strings"
)

func GetQueryHashes(branches []Branch, defaultBranchName string) []string {
	results := []string{}

	var hashes strings.Builder
	for i, branch := range branches {
		if branch.Name == defaultBranchName || len(branch.Commits) == 0 {
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
