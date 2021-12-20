//go:generate mockgen -source=poi.go -package=mocks -destination=./mocks/poi_mock.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	exec "golang.org/x/sys/execabs"
)

type (
	Connection interface {
		CheckRepos(hostname string, repoNames []string) error
		GetRepoNames() (string, error)
		GetBrancheNames() (string, error)
		GetPullRequests(hostname string, repoNames []string, queryHashes string) (string, error)
		DeleteBranches(branchNames []string) (string, error)
	}

	ConnectionImpl struct {
	}

	Repo struct {
		Hostname string
		Origin   string
		Upstream string
	}

	BranchState int

	Branch struct {
		Head         bool
		Name         string
		LastObjectId string
		PullRequests []PullRequest
		State        BranchState
	}

	PullRequestState int

	PullRequest struct {
		Name   string
		State  PullRequestState
		Number int
		Url    string
		Author string
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

func GetBranches(conn Connection) ([]Branch, error) {
	var hostname string
	var repoNames []string
	var defaultBranchName string
	if json, err := conn.GetRepoNames(); err == nil {
		hostname, repoNames, defaultBranchName, _ = getRepo(json)
	} else {
		return nil, err
	}

	err := conn.CheckRepos(hostname, repoNames)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	if names, err := conn.GetBrancheNames(); err == nil {
		branches = toBranch(strings.Split(names, "\n"))
	} else {
		return nil, err
	}

	prs := []PullRequest{}
	for _, queryHashes := range GetQueryHashes(branches, defaultBranchName) {
		json, err := conn.GetPullRequests(hostname, repoNames, queryHashes)
		if err != nil {
			return nil, err
		}

		if pr, err := toPullRequests(json); err == nil {
			prs = append(prs, pr...)
		}
	}

	branches = applyPullRequest(branches, prs)
	branches = checkDeletion(branches)
	return branches, nil
}

func applyPullRequest(branches []Branch, prs []PullRequest) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		prs := findMatchedPullRequest(branch.Name, prs)
		branch.PullRequests = prs
		results = append(results, branch)
	}
	return results
}

func findMatchedPullRequest(branchName string, prs []PullRequest) []PullRequest {
	results := []PullRequest{}
	for _, pr := range prs {
		if pr.Name == branchName {
			results = append(results, pr)
		}
	}
	return results
}

func checkDeletion(branches []Branch) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		branch.State = getDeleteStatus(branch)
		results = append(results, branch)
	}
	return results
}

func getDeleteStatus(branch Branch) BranchState {
	if branch.Head {
		return NotDeletable
	}

	if len(branch.PullRequests) == 0 {
		return NotDeletable
	}

	mergedCnt := 0
	for _, pr := range branch.PullRequests {
		if pr.State == Open {
			return NotDeletable
		}
		if pr.State == Merged {
			mergedCnt++
		}
	}
	if mergedCnt == 0 {
		return NotDeletable
	}

	return Deletable
}

func toBranch(branchNames []string) []Branch {
	branchNames = branchNames[:len(branchNames)-1]

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
			splitedNames[2],
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
	r := regexp.MustCompile("//(.+?)/")
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
						Url         string
						State       string
						Author      struct {
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

		results = append(results, PullRequest{
			edge.Node.HeadRefName,
			state,
			edge.Node.Number, edge.Node.Url, edge.Node.Author.Login,
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

	branchNamesAfter, err := conn.GetBrancheNames()
	if err != nil {
		return nil, err
	}
	branchesAfter := toBranch(strings.Split(branchNamesAfter, "\n"))

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

func (conn *ConnectionImpl) CheckRepos(hostname string, repoNames []string) error {
	for _, name := range repoNames {
		args := []string{
			"api",
			"--hostname", hostname,
			"repos/" + name,
			"--silent",
		}
		if _, err := run("gh", args); err != nil {
			return err
		}
	}
	return nil
}

func (conn *ConnectionImpl) GetRepoNames() (string, error) {
	args := []string{
		"repo", "view",
		"--json", "url",
		"--json", "owner",
		"--json", "name",
		"--json", "parent",
		"--json", "defaultBranchRef",
	}
	return run("gh", args)
}

func (conn *ConnectionImpl) GetBrancheNames() (string, error) {
	args := []string{
		"branch", "-v", "--no-abbrev",
		"--format=%(HEAD),%(refname:lstrip=2),%(objectname)",
	}
	return run("git", args)
}

func (conn *ConnectionImpl) GetPullRequests(
	hostname string, repoNames []string, queryHashes string) (string, error) {
	args := []string{
		"api", "graphql",
		"--hostname", hostname,
		"-f", fmt.Sprintf(`query=query {
  search(type: ISSUE, query: "is:pr %s %s", last: 100) {
    issueCount
    edges {
      node {
        ... on PullRequest {
          number
          url
          state
          headRefName
          author { login }
        }
      }
    }
  }
}`,
			getQueryRepos(repoNames),
			queryHashes,
		),
	}
	return run("gh", args)
}

func (conn *ConnectionImpl) DeleteBranches(branchNames []string) (string, error) {
	args := append([]string{
		"branch", "-D"},
		branchNames...)
	return run("git", args)
}

func getQueryRepos(repoNames []string) string {
	var repos strings.Builder
	for _, name := range repoNames {
		repos.WriteString(fmt.Sprintf("repo:%s ", name))
	}
	return repos.String()
}

func run(name string, args []string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to run external command: %s", name)
	}
	cmd.Wait()

	if stderr.Len() > 0 {
		return "", fmt.Errorf("failed to run external command: %s\n%s", name, stderr.String())
	}

	return stdout.String(), nil
}
