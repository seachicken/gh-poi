//go:generate mockgen -source=poi.go -package=mocks -destination=./mocks/poi_mock.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	Connection interface {
		CheckRepos(ctx context.Context, hostname string, repoNames []string) error
		GetRemoteNames(ctx context.Context) (string, error)
		GetSshConfig(ctx context.Context, name string) (string, error)
		GetRepoNames(ctx context.Context, hostname string, repoName string) (string, error)
		GetBranchNames(ctx context.Context) (string, error)
		GetMergedBranchNames(ctx context.Context, remoteName string, branchName string) (string, error)
		GetLog(ctx context.Context, branchName string) (string, error)
		GetAssociatedRefNames(ctx context.Context, oid string) (string, error)
		GetPullRequests(ctx context.Context, hostname string, repoNames []string, queryHashes string) (string, error)
		GetUncommittedChanges(ctx context.Context) (string, error)
		GetConfig(ctx context.Context, key string) (string, error)
		CheckoutBranch(ctx context.Context, branchName string) (string, error)
		DeleteBranches(ctx context.Context, branchNames []string) (string, error)
	}

	Remote struct {
		Name     string
		Hostname string
		RepoName string
	}

	BranchState int

	Branch struct {
		Head         bool
		Name         string
		IsMerged     bool
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

	UncommittedChange struct {
		X    string
		Y    string
		Path string
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

const (
	github    = "github.com"
	localhost = "github.localhost"
)

var detachedBranchNameRegex = regexp.MustCompile(`^\(.+\)`)
var ErrNotFound = errors.New("not found")

func GetRemote(ctx context.Context, connection Connection) (Remote, error) {
	remoteNames, err := connection.GetRemoteNames(ctx)
	if err != nil {
		return Remote{}, err
	}

	remotes := toRemotes(splitLines(remoteNames))
	if remote, err := getPrimaryRemote(remotes); err == nil {
		hostname := remote.Hostname
		if config, err := connection.GetSshConfig(ctx, hostname); err == nil {
			remote.Hostname = normalizeHostname(findHostname(splitLines(config), hostname))
		}
		return remote, nil
	} else {
		return Remote{}, err
	}
}

func GetBranches(ctx context.Context, remote Remote, connection Connection, dryRun bool) ([]Branch, error) {
	var repoNames []string
	var defaultBranchName string
	if json, err := connection.GetRepoNames(ctx, remote.Hostname, remote.RepoName); err == nil {
		repoNames, defaultBranchName, err = getRepo(json)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	err := connection.CheckRepos(ctx, remote.Hostname, repoNames)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	if names, err := connection.GetBranchNames(ctx); err == nil {
		branches = toBranch(splitLines(names))
		mergedNames, err := connection.GetMergedBranchNames(ctx, remote.Name, defaultBranchName)
		if err != nil {
			return nil, err
		}
		branches = applyMerged(branches, extractMergedBranchNames(splitLines(mergedNames)))
		branches, err = applyCommits(ctx, branches, defaultBranchName, connection)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	prs := []PullRequest{}
	for _, queryHashes := range getQueryHashes(branches) {
		json, err := connection.GetPullRequests(ctx, remote.Hostname, repoNames, queryHashes)
		if err != nil {
			return nil, err
		}

		if pr, err := toPullRequests(json); err == nil {
			prs = append(prs, pr...)
		}
	}

	branches = applyPullRequest(ctx, branches, prs, connection)

	var uncommittedChanges []UncommittedChange
	if changes, err := connection.GetUncommittedChanges(ctx); err == nil {
		uncommittedChanges = toUncommittedChange(splitLines(changes))
	} else {
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

		if !dryRun {
			_, err := connection.CheckoutBranch(ctx, defaultBranchName)
			if err != nil {
				return nil, err
			}
		}

		if !branchNameExists(defaultBranchName, branches) {
			result = append(result, Branch{
				true, defaultBranchName,
				false,
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

// https://github.com/cli/cli/blob/8f28d1f9d5b112b222f96eb793682ff0b5a7927d/internal/ghinstance/host.go#L26
func normalizeHostname(host string) string {
	hostname := strings.ToLower(host)
	if strings.HasSuffix(hostname, "."+github) {
		return github
	}
	if strings.HasSuffix(hostname, "."+localhost) {
		return localhost
	}
	return hostname
}

func toRemotes(remoteNames []string) []Remote {
	results := []Remote{}
	r := regexp.MustCompile(`^(.+?)\s+.+(?:@|//)(.+?)(?::|/)(.+?/.+?)(?:\.git|)\s+.+$`)
	for _, name := range remoteNames {
		found := r.FindStringSubmatch(name)
		if len(found) == 4 {
			results = append(results, Remote{found[1], found[2], found[3]})
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

func findHostname(params []string, defaultName string) string {
	for _, param := range params {
		kv := strings.Split(param, " ")
		if kv[0] == "hostname" {
			return kv[1]
		}
	}
	return defaultName
}

func extractMergedBranchNames(mergedNames []string) []string {
	result := []string{}
	r := regexp.MustCompile(`^[ *]+(.+)`)
	for _, name := range mergedNames {
		found := r.FindStringSubmatch(name)
		if len(found) > 1 {
			result = append(result, found[1])
		}
	}
	return result
}

func applyMerged(branches []Branch, mergedNames []string) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		branch.IsMerged = nameExists(branch.Name, mergedNames)
		results = append(results, branch)
	}
	return results
}

func nameExists(name string, names []string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

func applyCommits(ctx context.Context, branches []Branch, defaultBranchName string, connection Connection) ([]Branch, error) {
	results := []Branch{}

	for _, branch := range branches {
		if branch.Name == defaultBranchName || branch.IsDetached() {
			results = append(results, branch)
			continue
		}

		oids, err := connection.GetLog(ctx, branch.Name)
		if err != nil {
			return nil, err
		}

		trimmedOids, err := trimBranch(ctx, splitLines(oids), branch.IsMerged,
			branch.Name, defaultBranchName, connection)
		if err != nil {
			return nil, err
		}

		branch.Commits = trimmedOids
		results = append(results, branch)
	}

	return results, nil
}

func trimBranch(ctx context.Context, oids []string, isMerged bool,
	branchName string, defaultBranchName string, connection Connection) ([]string, error) {
	results := []string{}
	childNames := []string{}

	for i, oid := range oids {
		if isMerged {
			results = append(results, oid)
			break
		}

		refNames, err := connection.GetAssociatedRefNames(ctx, oid)
		if err != nil {
			return nil, err
		}
		names := extractBranchNames(splitLines(refNames))

		if i == 0 {
			for _, name := range names {
				if name == defaultBranchName {
					return []string{}, nil
				}
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

func applyPullRequest(ctx context.Context, branches []Branch, prs []PullRequest, connection Connection) []Branch {
	prNumbers := map[string]int{}
	for _, branch := range branches {
		if branch.IsDetached() {
			continue
		}
		mergeConfig, _ := connection.GetConfig(ctx, fmt.Sprintf("branch.%s.merge", branch.Name))
		if n := getPRNumber(mergeConfig); n > 0 {
			prNumbers[branch.Name] = n
		}
	}

	results := []Branch{}
	for _, branch := range branches {
		prs := findMatchedPullRequest(branch.Name, prs, prNumbers)
		sort.Slice(prs, func(i, j int) bool { return prs[i].Number < prs[j].Number })
		branch.PullRequests = prs
		results = append(results, branch)
	}
	return results
}

func getPRNumber(mergeConfig string) int {
	r := regexp.MustCompile(`^refs/pull/(\d+)`)
	found := r.FindStringSubmatch(mergeConfig)
	if len(found) > 0 {
		num, err := strconv.Atoi(found[1])
		if err != nil {
			return 0
		}
		return num
	} else {
		return 0
	}
}

func findMatchedPullRequest(branchName string, prs []PullRequest, prNumbers map[string]int) []PullRequest {
	results := []PullRequest{}

	prExists := func(pr PullRequest) bool {
		for _, result := range results {
			if pr.Number == result.Number {
				return true
			}
		}
		return false
	}

	prNumberExists := func(prNumber int) bool {
		for _, n := range prNumbers {
			if n == prNumber {
				return true
			}
		}
		return false
	}

	for _, pr := range prs {
		if prExists(pr) {
			continue
		}

		if prNumberExists(pr.Number) {
			if pr.Number == prNumbers[branchName] {
				results = append(results, pr)
			}
		} else if pr.Name == branchName {
			results = append(results, pr)
		}
	}

	return results
}

func toUncommittedChange(changes []string) []UncommittedChange {
	results := []UncommittedChange{}
	for _, change := range changes {
		results = append(results, UncommittedChange{
			string(change[0]),
			string(change[1]),
			string(change[3:]),
		})
	}
	return results
}

func checkDeletion(branches []Branch, uncommittedChanges []UncommittedChange) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		branch.State = getDeleteStatus(branch, uncommittedChanges)
		results = append(results, branch)
	}
	return results
}

func getDeleteStatus(branch Branch, uncommittedChanges []UncommittedChange) BranchState {
	hasTrackedChanges := false
	for _, change := range uncommittedChanges {
		if !change.IsUntracked() {
			hasTrackedChanges = true
			break
		}
	}
	if branch.Head && hasTrackedChanges {
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
		splitedNames := strings.Split(branchName, ":")

		results = append(results, Branch{
			splitedNames[0] == "*",
			splitedNames[1],
			false,
			[]string{},
			[]PullRequest{},
			Unknown,
		})
	}

	return results
}

func getRepo(jsonResp string) ([]string, string, error) {
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
	}

	var resp response
	if err := json.Unmarshal([]byte(jsonResp), &resp); err != nil {
		return nil, "", fmt.Errorf("error unmarshaling response: %w", err)
	}

	repoNames := []string{
		resp.Owner.Login + "/" + resp.Name,
	}
	if len(resp.Parent.Name) > 0 {
		repoNames = append(repoNames, resp.Parent.Owner.Login+"/"+resp.Parent.Name)
	}

	return repoNames, resp.DefaultBranchRef.Name, nil
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

func DeleteBranches(ctx context.Context, branches []Branch, connection Connection) ([]Branch, error) {
	branchNames := getBranchNames(branches, Deletable)
	if len(branchNames) == 0 {
		return branches, nil
	}

	connection.DeleteBranches(ctx, branchNames)

	branchNamesAfter, err := connection.GetBranchNames(ctx)
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
	return strings.FieldsFunc(strings.Replace(text, "\r\n", "\n", -1),
		func(c rune) bool { return c == '\n' })
}

func (b Branch) IsDetached() bool {
	return detachedBranchNameRegex.MatchString(b.Name)
}

func (uc *UncommittedChange) IsUntracked() bool {
	return uc.Y == "?"
}
