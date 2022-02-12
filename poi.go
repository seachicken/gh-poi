//go:generate mockgen -source=poi.go -package=mocks -destination=./mocks/poi_mock.go
package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type (
	Connection interface {
		CheckRepos(hostname string, repoNames []string) error
		GetRemoteNames() (string, error)
		GetRepoNames(repoName string) (string, error)
		GetBranchNames() (string, error)
		GetLog(branchName string) (string, error)
		GetAssociatedRefNames(oid string) (string, error)
		GetPullRequests(hostname string, repoNames []string, queryHashes string) (string, error)
		GetUncommittedChanges() (string, error)
		CheckoutBranch(branchName string) (string, error)
		DeleteBranches(branchNames []string) (string, error)
	}

	Remote struct {
		Name     string
		RepoName string
	}

	BranchState int

	Branch struct {
		Head         bool
		Name         string
		Commits      []string
		PullRequests []PullRequest
		State        BranchState
	}

	PullRequestState int

	PullRequest struct {
		Name    string
		State   PullRequestState
		IsDraft bool
		Number  int
		Commits []string
		Url     string
		Author  string
	}
)

const (
	Unknown BranchState = iota
	NotDeletable
	Deletable
	Deleted
)

const (
	Closed PullRequestState = iota
	Merged
	Open
)

var ErrNotFound = errors.New("not found")

func GetBranches(conn Connection, check bool) ([]Branch, error) {
	primaryRepoName := ""
	if remoteNames, err := conn.GetRemoteNames(); err == nil {
		remotes := toRemotes(splitLines(remoteNames))
		if remote, err := getPrimaryRemote(remotes); err == nil {
			primaryRepoName = remote.RepoName
		}
	} else {
		return nil, err
	}

	var hostname string
	var repoNames []string
	var defaultBranchName string
	if json, err := conn.GetRepoNames(primaryRepoName); err == nil {
		hostname, repoNames, defaultBranchName, _ = getRepo(json)
	} else {
		return nil, err
	}

	err := conn.CheckRepos(hostname, repoNames)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	if names, err := conn.GetBranchNames(); err == nil {
		branches = toBranch(splitLines(names))
		branches, err = applyCommits(branches, defaultBranchName, conn)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	prs := []PullRequest{}
	for _, queryHashes := range getQueryHashes(branches) {
		json, err := conn.GetPullRequests(hostname, repoNames, queryHashes)
		if err != nil {
			return nil, err
		}

		if pr, err := toPullRequests(json); err == nil {
			prs = append(prs, pr...)
		}
	}

	branches = applyPullRequest(branches, prs)

	uncommittedChanges, err := conn.GetUncommittedChanges()
	if err != nil {
		return nil, err
	}

	branches = checkDeletion(branches, uncommittedChanges)

	needsCheckout := false
	for _, branch := range branches {
		if branch.Head && branch.State == Deletable {
			needsCheckout = true
			break
		}
	}

	if needsCheckout {
		result := []Branch{}

		if !check {
			_, err := conn.CheckoutBranch(defaultBranchName)
			if err != nil {
				return nil, err
			}
		}

		if !branchNameExists(defaultBranchName, branches) {
			result = append(result, Branch{
				true, defaultBranchName,
				[]string{},
				[]PullRequest{},
				NotDeletable,
			})
		}

		for _, branch := range branches {
			if branch.Name == defaultBranchName {
				branch.Head = true
			} else {
				branch.Head = false
			}
			result = append(result, branch)
		}

		branches = result
	}

	sort.Slice(branches, func(i, j int) bool { return branches[i].Name < branches[j].Name })

	return branches, nil
}

func toRemotes(remoteNames []string) []Remote {
	results := []Remote{}
	r := regexp.MustCompile(`^(.+?)\s+.+(?::|/)(.+?/.+?)(?:\.git|)\s+.+$`)
	for _, name := range remoteNames {
		found := r.FindStringSubmatch(name)
		if len(found) == 3 {
			results = append(results, Remote{found[1], found[2]})
		}
	}
	return results
}

func getPrimaryRemote(remotes []Remote) (Remote, error) {
	if len(remotes) == 0 {
		return Remote{}, ErrNotFound
	}

	for _, remote := range remotes {
		if remote.Name == "origin" {
			return remote, nil
		}
	}
	return remotes[0], nil
}

func applyCommits(branches []Branch, defaultBranchName string, conn Connection) ([]Branch, error) {
	results := []Branch{}

	for _, branch := range branches {
		if branch.Name == defaultBranchName {
			results = append(results, branch)
			continue
		}

		oids, err := conn.GetLog(branch.Name)
		if err != nil {
			return nil, err
		}

		trimmedOids, err := trimBranch(splitLines(oids), branch.Name, conn)
		if err != nil {
			return nil, err
		}

		branch.Commits = trimmedOids
		results = append(results, branch)
	}

	return results, nil
}

func trimBranch(oids []string, branchName string, conn Connection) ([]string, error) {
	results := []string{}
	childNames := []string{}

	for i, oid := range oids {
		refNames, err := conn.GetAssociatedRefNames(oid)
		if err != nil {
			return nil, err
		}
		names := extractBranchNames(splitLines(refNames))

		if i == 0 {
			for _, name := range names {
				if name != branchName {
					childNames = append(childNames, name)
				}
			}
		}

		isChild := func(name string) bool {
			for _, childName := range childNames {
				if name == childName {
					return true
				}
			}
			return false
		}

		for _, name := range names {
			if name != branchName && !isChild(name) {
				return results, nil
			}
		}

		results = append(results, oid)
	}

	return results, nil
}

func extractBranchNames(refNames []string) []string {
	result := []string{}
	r := regexp.MustCompile(`^refs/(?:heads|remotes/.+?)/`)
	for _, name := range refNames {
		result = append(result, r.ReplaceAllString(name, ""))
	}
	return result
}

func applyPullRequest(branches []Branch, prs []PullRequest) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		prs := findMatchedPullRequest(branch.Name, prs)
		sort.Slice(prs, func(i, j int) bool { return prs[i].Number < prs[j].Number })
		branch.PullRequests = prs
		results = append(results, branch)
	}
	return results
}

func findMatchedPullRequest(branchName string, prs []PullRequest) []PullRequest {
	results := []PullRequest{}

	exists := func(pr PullRequest) bool {
		for _, result := range results {
			if pr.Number == result.Number {
				return true
			}
		}
		return false
	}

	for _, pr := range prs {
		if pr.Name == branchName && !exists(pr) {
			results = append(results, pr)
		}
	}
	return results
}

func checkDeletion(branches []Branch, uncommittedChanges string) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		branch.State = getDeleteStatus(branch, uncommittedChanges)
		results = append(results, branch)
	}
	return results
}

func getDeleteStatus(branch Branch, uncommittedChanges string) BranchState {
	if branch.Head && len(uncommittedChanges) > 0 {
		return NotDeletable
	}

	if len(branch.PullRequests) == 0 {
		return NotDeletable
	}

	fullyMergedCnt := 0
	for _, pr := range branch.PullRequests {
		if pr.State == Open {
			return NotDeletable
		}
		if isFullyMerged(branch, pr) {
			fullyMergedCnt++
		}
	}
	if fullyMergedCnt == 0 {
		return NotDeletable
	}

	return Deletable
}

func isFullyMerged(branch Branch, pr PullRequest) bool {
	if pr.State != Merged || len(branch.Commits) == 0 {
		return false
	}

	localHeadOid := branch.Commits[0]
	for _, oid := range pr.Commits {
		if oid == localHeadOid {
			return true
		}
	}

	return false
}

func toBranch(branchNames []string) []Branch {
	results := []Branch{}

	for _, branchName := range branchNames {
		splitedNames := strings.Split(branchName, ",")
		head := false
		if splitedNames[0] == "*" {
			head = true
		}
		results = append(results, Branch{
			head,
			splitedNames[1],
			[]string{},
			[]PullRequest{},
			Unknown,
		})
	}

	return results
}

func getRepo(jsonResp string) (string, []string, string, error) {
	type response struct {
		DefaultBranchRef struct {
			Name string
		}
		Name  string
		Owner struct {
			Login string
		}
		Parent struct {
			Name  string
			Owner struct {
				Login string
			}
			DefaultBranchName string
		}
		Url string
	}

	var resp response
	if err := json.Unmarshal([]byte(jsonResp), &resp); err != nil {
		return "", nil, "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	repoNames := []string{
		resp.Owner.Login + "/" + resp.Name,
	}
	if len(resp.Parent.Name) > 0 {
		repoNames = append(repoNames, resp.Parent.Owner.Login+"/"+resp.Parent.Name)
	}

	return getHostname(resp.Url), repoNames, resp.DefaultBranchRef.Name, nil
}

func getHostname(url string) string {
	r := regexp.MustCompile(`//(.+?)/`)
	found := r.FindSubmatch([]byte(url))
	return string(found[1])
}

func toPullRequests(jsonResp string) ([]PullRequest, error) {
	type response struct {
		Data struct {
			Search struct {
				IssueCount int
				Edges      []struct {
					Node struct {
						Number      int
						HeadRefName string
						HeadRefOid  string
						Url         string
						State       string
						IsDraft     bool
						Commits     struct {
							Nodes []struct {
								Commit struct {
									Oid string
								}
							}
						}
						Author struct {
							Login string
						}
					}
				}
			}
		}
	}

	var resp response
	if err := json.Unmarshal([]byte(jsonResp), &resp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	results := []PullRequest{}
	for _, edge := range resp.Data.Search.Edges {
		state, err := toPullRequestState(edge.Node.State)
		if err == ErrNotFound {
			return nil, fmt.Errorf("unexpected pull request state: %s", edge.Node.State)
		}

		commits := []string{}
		for _, node := range edge.Node.Commits.Nodes {
			commits = append(commits, node.Commit.Oid)
		}

		results = append(results, PullRequest{
			edge.Node.HeadRefName,
			state,
			edge.Node.IsDraft,
			edge.Node.Number,
			commits,
			edge.Node.Url,
			edge.Node.Author.Login,
		})
	}

	return results, nil
}

func toPullRequestState(state string) (PullRequestState, error) {
	switch state {
	case "CLOSED":
		return Closed, nil
	case "MERGED":
		return Merged, nil
	case "OPEN":
		return Open, nil
	default:
		return 0, ErrNotFound
	}
}

func DeleteBranches(branches []Branch, conn Connection) ([]Branch, error) {
	branchNames := getBranchNames(branches, Deletable)
	if len(branchNames) == 0 {
		return branches, nil
	}

	conn.DeleteBranches(branchNames)

	branchNamesAfter, err := conn.GetBranchNames()
	if err != nil {
		return nil, err
	}
	branchesAfter := toBranch(splitLines(branchNamesAfter))

	return checkDeleted(branches, branchesAfter), nil
}

func getBranchNames(branches []Branch, state BranchState) []string {
	results := []string{}
	for _, branch := range branches {
		if branch.State == state {
			results = append(results, branch.Name)
		}
	}
	return results
}

func checkDeleted(branchesBefore []Branch, branchesAfter []Branch) []Branch {
	results := []Branch{}
	for _, branch := range branchesBefore {
		if branch.State == Deletable {
			if !branchNameExists(branch.Name, branchesAfter) {
				branch.State = Deleted
			}
		}
		results = append(results, branch)
	}
	return results
}

func branchNameExists(branchName string, branches []Branch) bool {
	for _, branch := range branches {
		if branch.Name == branchName {
			return true
		}
	}
	return false
}

func splitLines(text string) []string {
	return strings.FieldsFunc(text, func(c rune) bool { return c == '\n' })
}
