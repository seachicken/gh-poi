package main

import (
	"fmt"
	"strings"
)

func getQueryHashes(branches []Branch) []string {
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
