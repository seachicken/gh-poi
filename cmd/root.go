package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/seachicken/gh-poi/shared"
)

type (
	UncommittedChange struct {
		X    string
		Y    string
		Path string
	}
)

const (
	github    = "github.com"
	localhost = "github.localhost"
)

var ErrNotFound = errors.New("not found")

func GetRemote(ctx context.Context, connection shared.Connection) (shared.Remote, error) {
	remoteNames, err := connection.GetRemoteNames(ctx)
	if err != nil {
		return shared.Remote{}, err
	}

	remotes := toRemotes(SplitLines(remoteNames))
	if remote, err := getPrimaryRemote(remotes); err == nil {
		hostname := remote.Hostname
		ghHost := os.Getenv("GH_HOST")
		if ghHost == "" {
			if config, err := connection.GetSshConfig(ctx, hostname); err == nil {
				remote.Hostname = normalizeHostname(findHostname(SplitLines(config), hostname))
			}
		} else {
			remote.Hostname = ghHost
		}
		return remote, nil
	} else {
		return shared.Remote{}, err
	}
}

func GetBranches(ctx context.Context, remote shared.Remote, connection shared.Connection, state shared.PullRequestState, dryRun bool) ([]shared.
	Branch, error) {
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

	branches, err := loadBranches(ctx, remote, defaultBranchName, repoNames, connection)
	if err != nil {
		return nil, err
	}

	branches = checkDeletion(branches, state)

	branches, err = switchToDefaultBranchIfDeleted(ctx, branches, defaultBranchName, connection, dryRun)
	if err != nil {
		return nil, err
	}

	sort.Slice(branches, func(i, j int) bool { return branches[i].Name < branches[j].Name })

	return branches, nil
}

func loadBranches(ctx context.Context, remote shared.Remote, defaultBranchName string, repoNames []string, connection shared.Connection) ([]shared.Branch, error) {
	var branches []shared.Branch

	if names, err := connection.GetBranchNames(ctx); err == nil {
		branches = ToBranch(SplitLines(names))
		branches = applyDefault(branches, defaultBranchName)
		mergedNames, err := connection.GetMergedBranchNames(ctx, remote.Name, defaultBranchName)
		if err != nil {
			return nil, err
		}
		branches = applyMerged(branches, extractMergedBranchNames(SplitLines(mergedNames)))
		branches, err = applyProtected(ctx, branches, connection)
		if err != nil {
			return nil, err
		}
		branches, err = applyCommits(ctx, remote, branches, defaultBranchName, connection)
		if err != nil {
			return nil, err
		}
		branches, err = applyTrackedChanges(ctx, branches, connection)
		if err != nil {
			return nil, err
		}
		branches, err = applyWorktrees(ctx, branches, connection)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	prs := []shared.PullRequest{}
	orgs := shared.GetQueryOrgs(repoNames)
	repos := shared.GetQueryRepos(repoNames)

	type pullRequestResult struct {
		prs []shared.PullRequest
		err error
	}

	queryHashes := shared.GetQueryHashes(branches)
	prChan := make(chan pullRequestResult, len(queryHashes))
	var wg sync.WaitGroup

	for _, queryHash := range queryHashes {
		wg.Add(1)
		go func(hash string) {
			defer wg.Done()
			json, err := connection.GetPullRequests(ctx, remote.Hostname, orgs, repos, hash)
			if err != nil {
				prChan <- pullRequestResult{err: err}
				return
			}

			pr, err := toPullRequests(json)
			if err != nil {
				prChan <- pullRequestResult{err: err}
				return
			}

			prChan <- pullRequestResult{prs: pr}
		}(queryHash)
	}

	go func() {
		wg.Wait()
		close(prChan)
	}()

	for result := range prChan {
		if result.err != nil {
			return nil, result.err
		}
		prs = append(prs, result.prs...)
	}

	branches = applyPullRequest(ctx, branches, prs, connection)

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

func toRemotes(remoteConfigs []string) []shared.Remote {
	results := []shared.Remote{}
	for _, remoteConfig := range remoteConfigs {
		results = append(results, shared.NewRemote(remoteConfig))
	}
	return results
}

func getPrimaryRemote(remotes []shared.Remote) (shared.Remote, error) {
	if len(remotes) == 0 {
		return shared.Remote{}, ErrNotFound
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

func applyDefault(branches []shared.Branch, defaultBranchName string) []shared.Branch {
	results := []shared.Branch{}
	for _, branch := range branches {
		if branch.Name == defaultBranchName {
			branch.IsDefault = true
		}
		results = append(results, branch)
	}
	return results
}

func applyMerged(branches []shared.Branch, mergedNames []string) []shared.Branch {
	results := []shared.Branch{}
	for _, branch := range branches {
		branch.IsMerged = slices.Contains(mergedNames, branch.Name)
		results = append(results, branch)
	}
	return results
}

func applyProtected(ctx context.Context, branches []shared.Branch, connection shared.Connection) ([]shared.Branch, error) {
	results := []shared.Branch{}

	for _, branch := range branches {
		config, _ := connection.GetConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", branch.Name))
		splitConfig := SplitLines(config)
		if len(splitConfig) > 0 && splitConfig[0] == "true" {
			branch.IsProtected = true
		}
		results = append(results, branch)
	}

	return results, nil
}

func applyCommits(ctx context.Context, remote shared.Remote, branches []shared.Branch, defaultBranchName string, connection shared.Connection) ([]shared.Branch, error) {
	var wg sync.WaitGroup

	type remoteBranchResult struct {
		branch shared.Branch
		err    error
	}

	results := []shared.Branch{}
	resultChan := make(chan remoteBranchResult, len(branches))

	for _, branch := range branches {
		wg.Add(1)
		go func(branch shared.Branch) {
			defer wg.Done()

			if branch.Name == defaultBranchName || branch.IsDetached() {
				branch.Commits = []string{}
				resultChan <- remoteBranchResult{branch: branch}
				return
			}

			if remoteHeadOid, err := connection.GetRemoteHeadOid(ctx, remote.Name, branch.Name); err == nil {
				branch.RemoteHeadOid = SplitLines(remoteHeadOid)[0]
			} else {
				result, _ := connection.GetConfig(ctx, fmt.Sprintf("branch.%s.remote", branch.Name))
				splitResults := SplitLines(result)
				if len(splitResults) > 0 {
					remoteUrl := splitResults[0]
					if result, err := connection.GetLsRemoteHeadOid(ctx, remoteUrl, branch.Name); err == nil {
						splitResults := strings.Fields(result)
						if len(splitResults) > 0 {
							branch.RemoteHeadOid = splitResults[0]
						}
					}
				}
			}

			oids, err := connection.GetLog(ctx, branch.Name)
			if err != nil {
				resultChan <- remoteBranchResult{err: err}
				return
			}

			trimmedOids, err := trimBranch(
				ctx, SplitLines(oids), branch.RemoteHeadOid, branch.IsMerged,
				branch.Name, defaultBranchName, connection)
			if err != nil {
				resultChan <- remoteBranchResult{err: err}
				return
			}

			branch.Commits = trimmedOids
			resultChan <- remoteBranchResult{branch: branch}
		}(branch)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result.err != nil {
			return nil, result.err
		}
		results = append(results, result.branch)
	}

	return results, nil
}

func applyTrackedChanges(ctx context.Context, branches []shared.Branch, connection shared.Connection) ([]shared.Branch, error) {
	var uncommittedChanges []UncommittedChange
	if changes, err := connection.GetUncommittedChanges(ctx); err == nil {
		uncommittedChanges = toUncommittedChange(SplitLines(changes))
	} else {
		return nil, err
	}

	results := []shared.Branch{}
	for _, branch := range branches {
		if branch.Head {
			hasTrackedChanges := false
			for _, change := range uncommittedChanges {
				if !change.IsUntracked() {
					hasTrackedChanges = true
					break
				}
			}
			if hasTrackedChanges {
				branch.HasTrackedChanges = true
			}
		}
		results = append(results, branch)
	}
	return results, nil
}

func applyWorktrees(ctx context.Context, branches []shared.Branch, connection shared.Connection) ([]shared.Branch, error) {
	worktreeOutput, err := connection.GetWorktrees(ctx)
	if err != nil {
		// Worktrees might not be supported or available, continue gracefully
		return branches, nil
	}

	worktrees := shared.ParseWorktrees(worktreeOutput)

	// Create a map for quick branch-to-worktree lookup
	worktreeMap := make(map[string]*shared.Worktree)
	for i := range worktrees {
		if worktrees[i].Branch != "" {
			worktreeMap[worktrees[i].Branch] = &worktrees[i]
		}
	}

	results := []shared.Branch{}
	for _, branch := range branches {
		if wt, ok := worktreeMap[branch.Name]; ok {
			branch.Worktree = wt
		}
		results = append(results, branch)
	}

	return results, nil
}

func trimBranch(ctx context.Context, oids []string, remoteHeadOid string, isMerged bool,
	branchName string, defaultBranchName string, connection shared.Connection) ([]string, error) {
	results := []string{}
	childNames := []string{}

	for i, oid := range oids {
		if len(remoteHeadOid) > 0 || isMerged {
			results = append(results, oid)
			break
		}

		refNames, err := connection.GetAssociatedRefNames(ctx, oid)
		if err != nil {
			return nil, err
		}
		names := extractBranchNames(SplitLines(refNames))

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

func applyPullRequest(ctx context.Context, branches []shared.Branch, prs []shared.PullRequest, connection shared.Connection) []shared.Branch {
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

	results := []shared.Branch{}
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

func findMatchedPullRequest(branchName string, prs []shared.PullRequest, prNumbers map[string]int) []shared.PullRequest {
	results := []shared.PullRequest{}

	prExists := func(pr shared.PullRequest) bool {
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

func checkDeletion(branches []shared.Branch, state shared.PullRequestState) []shared.Branch {
	results := []shared.Branch{}
	for _, branch := range branches {
		branch.State = getDeleteStatus(branch, state)
		results = append(results, branch)
	}
	return results
}

func getDeleteStatus(branch shared.Branch, state shared.PullRequestState) shared.BranchState {
	if branch.IsProtected {
		return shared.NotDeletable
	}

	if branch.HasTrackedChanges {
		return shared.NotDeletable
	}

	if len(branch.PullRequests) == 0 {
		return shared.NotDeletable
	}

	fullyMergedCnt := 0
	for _, pr := range branch.PullRequests {
		if pr.State == shared.Open {
			return shared.NotDeletable
		}
		if isFullyMerged(branch, pr, state) {
			fullyMergedCnt++
		}
	}
	if fullyMergedCnt == 0 {
		return shared.NotDeletable
	}

	return shared.Deletable
}

func isFullyMerged(branch shared.Branch, pr shared.PullRequest, state shared.PullRequestState) bool {
	if len(branch.Commits) == 0 {
		return false
	}
	if (state == shared.Merged && pr.State != shared.Merged) ||
		// In the GitHub interface, closed status includes merged status, so we make it behave the same way.
		// https://github.com/cli/cli/issues/8102
		(state == shared.Closed && pr.State != shared.Closed && pr.State != shared.Merged) {
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

func switchToDefaultBranchIfDeleted(ctx context.Context, branches []shared.Branch, defaultBranchName string, connection shared.Connection, dryRun bool) ([]shared.Branch, error) {
	needsCheckout := false
	for _, branch := range branches {
		if branch.Head && branch.State == shared.Deletable {
			needsCheckout = true
			break
		}
	}

	if !needsCheckout {
		return branches, nil
	}

	results := []shared.Branch{}

	if !dryRun {
		_, err := connection.CheckoutBranch(ctx, defaultBranchName)
		if err != nil {
			return nil, err
		}
	}

	if !BranchNameExists(defaultBranchName, branches) {
		branch := shared.Branch{}
		branch.Head = true
		branch.Name = defaultBranchName
		branch.State = shared.NotDeletable
		results = append(results, branch)
	}

	for _, branch := range branches {
		if branch.Name == defaultBranchName {
			branch.Head = true
		} else {
			branch.Head = false
		}
		results = append(results, branch)
	}

	return results, nil
}

func ToBranch(branchNames []string) []shared.Branch {
	results := []shared.Branch{}

	for _, branchName := range branchNames {
		branch := shared.Branch{}
		splitNames := strings.Split(branchName, ":")
		branch.Head = splitNames[0] == "*"
		branch.Name = splitNames[1]
		results = append(results, branch)
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

func toPullRequests(jsonResp string) ([]shared.PullRequest, error) {
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

	results := []shared.PullRequest{}
	for _, edge := range resp.Data.Search.Edges {
		state, err := toPullRequestState(edge.Node.State)
		if err == ErrNotFound {
			return nil, fmt.Errorf("unexpected pull request state: %s", edge.Node.State)
		}

		commits := []string{}
		for _, node := range edge.Node.Commits.Nodes {
			commits = append(commits, node.Commit.Oid)
		}

		results = append(results, shared.PullRequest{
			Name:    edge.Node.HeadRefName,
			State:   state,
			IsDraft: edge.Node.IsDraft,
			Number:  edge.Node.Number,
			Commits: commits,
			Url:     edge.Node.Url,
			Author:  edge.Node.Author.Login,
		})
	}

	return results, nil
}

func toPullRequestState(state string) (shared.PullRequestState, error) {
	switch state {
	case "CLOSED":
		return shared.Closed, nil
	case "MERGED":
		return shared.Merged, nil
	case "OPEN":
		return shared.Open, nil
	default:
		return 0, ErrNotFound
	}
}

func DeleteBranches(ctx context.Context, branches []shared.Branch, connection shared.Connection) ([]shared.Branch, error) {
	branchNames := getBranchNames(branches, shared.Deletable)
	if len(branchNames) == 0 {
		return branches, nil
	}

	connection.DeleteBranches(ctx, branchNames)

	branchNamesAfter, err := connection.GetBranchNames(ctx)
	if err != nil {
		return nil, err
	}
	branchesAfter := ToBranch(SplitLines(branchNamesAfter))

	return checkDeleted(branches, branchesAfter), nil
}

func DeleteWorktrees(ctx context.Context, branches []shared.Branch, connection shared.Connection) (map[string]bool, error) {
	deleted := make(map[string]bool)
	var errs []error
	for _, branch := range branches {
		if branch.State != shared.Deletable {
			continue
		}

		if branch.Worktree == nil {
			continue
		}

		// Cannot remove main worktree
		if branch.Worktree.IsMain {
			continue
		}

		_, err := connection.RemoveWorktree(ctx, branch.Worktree.Path)
		if err != nil {
			errs = append(errs, err)
		} else {
			deleted[branch.Name] = true
		}
	}
	return deleted, errors.Join(errs...)
}

func getBranchNames(branches []shared.Branch, state shared.BranchState) []string {
	results := []string{}
	for _, branch := range branches {
		if branch.State == state {
			results = append(results, branch.Name)
		}
	}
	return results
}

func checkDeleted(branchesBefore []shared.Branch, branchesAfter []shared.Branch) []shared.Branch {
	results := []shared.Branch{}
	for _, branch := range branchesBefore {
		if branch.State == shared.Deletable {
			if !BranchNameExists(branch.Name, branchesAfter) {
				branch.State = shared.Deleted
			}
		}
		results = append(results, branch)
	}
	return results
}

func BranchNameExists(branchName string, branches []shared.Branch) bool {
	for _, branch := range branches {
		if branch.Name == branchName {
			return true
		}
	}
	return false
}

func SplitLines(text string) []string {
	return strings.FieldsFunc(strings.Replace(text, "\r\n", "\n", -1),
		func(c rune) bool { return c == '\n' })
}

func (uc *UncommittedChange) IsUntracked() bool {
	return uc.Y == "?"
}
