//go:generate mockgen -source=poi.go -package=mocks -destination=./mocks/poi_mock.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/cli/safeexec"
	"github.com/pkg/errors"
)

type (
	Connection interface {
		GetRemoteName() string
		GetBrancheNames() string
		DeleteBranches(branchNames []string) string
		FetchRepoNames() string
		FetchPrStates(hostname string, repoNames []string, queryHashes string) string
	}

	ConnectionImpl struct {
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
	hostname := getHostname(conn.GetRemoteName())
	repoNames := strings.Split(conn.FetchRepoNames(), ",")
	branches := toBranch(strings.Split(conn.GetBrancheNames(), "\n"))

	prs := []PullRequest{}
	for _, queryHashes := range GetQueryHashes(branches) {
		pr, err := fromJson(conn.FetchPrStates(hostname, repoNames, queryHashes))
		if err != nil {
			return []Branch{}, err
		}
		prs = append(prs, pr...)
	}

	branches = applyPullRequest(branches, prs)
	branches = checkDeletion(branches)
	return branches, nil
}

func getHostname(remoteName string) string {
	r := regexp.MustCompile("(?:@|//)(.+):")
	found := r.FindSubmatch([]byte(remoteName))
	return string(found[1])
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

func fromJson(jsonResp string) ([]PullRequest, error) {
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
					}
				}
			}
		}
	}

	var resp response
	err := json.Unmarshal([]byte(jsonResp), &resp)
	if err != nil {
		return []PullRequest{}, err
	}

	results := []PullRequest{}
	for _, edge := range resp.Data.Search.Edges {
		state, err := toPullRequestState(edge.Node.State)
		if err != nil {
			return []PullRequest{}, err
		}

		results = append(results, PullRequest{
			edge.Node.HeadRefName,
			state,
			edge.Node.Number, edge.Node.Url,
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

func DeleteBranches(branches []Branch, conn Connection) []Branch {
	branchNames := getBranchNames(branches, Deletable)
	if len(branchNames) == 0 {
		return branches
	}

	conn.DeleteBranches(branchNames)

	branchesAfter := toBranch(strings.Split(conn.GetBrancheNames(), "\n"))

	return checkDeleted(branches, branchesAfter)
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

func checkDeleted(branches []Branch, branchesAfter []Branch) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		if branch.State == Deletable {
			_, err := findMatchedBranch(branch.Name, branchesAfter)
			if err == ErrNotFound {
				branch.State = Deleted
			}
		}
		results = append(results, branch)
	}
	return results
}

func findMatchedBranch(branchName string, branches []Branch) (Branch, error) {
	for _, branch := range branches {
		if branch.Name == branchName {
			return branch, nil
		}
	}
	return Branch{}, ErrNotFound
}

func (conn *ConnectionImpl) GetRemoteName() string {
	args := []string{
		"remote", "-v",
	}
	stdout, _, _ := run("git", args)

	return stdout.String()
}

func (conn *ConnectionImpl) GetBrancheNames() string {
	args := []string{
		"branch", "-v", "--no-abbrev",
		"--format=%(HEAD),%(refname:lstrip=2),%(objectname)",
	}
	stdout, _, _ := run("git", args)

	return stdout.String()
}

func (conn *ConnectionImpl) DeleteBranches(branchNames []string) string {
	args := append([]string{
		"branch", "-D"},
		branchNames...)
	stdout, _, _ := run("git", args)

	return stdout.String()
}

func (conn *ConnectionImpl) FetchRepoNames() string {
	args := []string{
		"repo", "view",
		"--json", "owner",
		"--json", "name",
		"--json", "parent",
		"--template", "{{ .owner.login }}/{{ .name }}{{ if.parent }},{{ .parent.owner.login }}/{{ .parent.name }}{{ end }}",
	}
	stdout, _, _ := run("gh", args)

	return stdout.String()
}

func (conn *ConnectionImpl) FetchPrStates(
	hostname string, repoNames []string, queryHashes string) string {
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
                title
                url
                state
                headRefName
                author { login avatarUrl }
              }
            }
          }
        }
      }`,
			getQueryRepos(repoNames),
			queryHashes,
		),
	}
	stdout, _, _ := run("gh", args)

	return stdout.String()
}

func getQueryRepos(repoNames []string) string {
	var repos strings.Builder
	for _, name := range repoNames {
		repos.WriteString(fmt.Sprintf("repo:%s ", name))
	}
	return repos.String()
}

func run(file string, args []string) (stdout, stderr bytes.Buffer, err error) {
	bin, err := safeexec.LookPath(file)
	if err != nil {
		return
	}

	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return
	}

	return
}
