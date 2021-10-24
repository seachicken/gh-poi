package main

import (
	"fmt"
	"strings"
)

func GetQueryHashes(branches []Branch) []string {
	results := []string{}

	var hashes strings.Builder
	for i, branch := range branches {
		separator := " "
		if i == len(branches)-1 {
			separator = ""
		}
		hash := fmt.Sprintf("hash:%s%s", branch.LastObjectId, separator)

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
