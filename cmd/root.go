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

	"github.com/seachicken/gh-poi/conn"
	"github.com/seachicken/gh-poi/shared"
)

const (
	github    = "github.com"
	localhost = "github.localhost"
)

var ErrNotFound = errors.New("not found")

// Returns a list of remotes prioritized for PR discovery.
// Both modes prioritize "origin," and when searching for pull requests,
// the parent (a.k.a upstream) repository is also included in the search.
//
// quick:
//   - Focuses on the most likely PR sources to minimize API calls.
//   - Returns only "origin" and the remote configured via `gh repo set-default`.
//
// deep:
//   - Scans all registered remotes to ensure comprehensive PR discovery.
//   - Useful for complex setups where PRs may span multiple forks or parents.
func GetPreferredRemotes(ctx context.Context, connection shared.Connection, scan shared.ScanMode) ([]shared.Remote, error) {
	remotes, err := conn.GetRemoteNames(ctx, connection)
	if err != nil {
		return []shared.Remote{}, err
	}
	if len(remotes) == 0 {
		return []shared.Remote{}, ErrNotFound
	}

	uniqueRemotes := make(map[string]shared.Remote)
	for _, remote := range remotes {
		uniqueRemotes[remote.Name] = remote
	}

	var primaryRemote *shared.Remote
	var ghResolvedRemote *shared.Remote
	var otherRemotes []shared.Remote
	for key, remote := range uniqueRemotes {
		config, _ := connection.GetConfig(ctx, fmt.Sprintf("remote.%s.gh-resolved", remote.Name))
		splitConfig := SplitLines(config)
		if len(splitConfig) > 0 && len(splitConfig[0]) > 0 {
			remote.GhResolved = splitConfig[0]
			ghResolvedRemote = &remote
			uniqueRemotes[key] = remote
		}
	}
	_, ok := uniqueRemotes["origin"]
	if ok {
		for _, remote := range uniqueRemotes {
			if remote.Name == "origin" {
				primaryRemote = &remote
			} else {
				otherRemotes = append(otherRemotes, remote)
			}
		}
	} else {
		first := true
		for _, remote := range uniqueRemotes {
			if first {
				primaryRemote = &remote
			} else {
				otherRemotes = append(otherRemotes, remote)
			}
			first = false
		}
	}

	preferredRemotes := []shared.Remote{}
	if scan == shared.Quick {
		if ghResolvedRemote == nil || primaryRemote.Name == ghResolvedRemote.Name {
			preferredRemotes = []shared.Remote{*primaryRemote}
		} else {
			preferredRemotes = []shared.Remote{*primaryRemote, *ghResolvedRemote}
		}
	} else {
		preferredRemotes = append([]shared.Remote{*primaryRemote}, otherRemotes...)
	}

	ghHost := os.Getenv("GH_HOST")
	for key, remote := range preferredRemotes {
		if ghHost == "" {
			if config, err := connection.GetSshConfig(ctx, remote.Hostname); err == nil {
				remote.Hostname = normalizeHostname(findHostname(SplitLines(config), remote.Hostname))
			}
		} else {
			remote.Hostname = ghHost
		}
		preferredRemotes[key] = remote
	}

	return preferredRemotes, nil
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

func findHostname(params []string, defaultName string) string {
	for _, param := range params {
		kv := strings.Split(param, " ")
		if kv[0] == "hostname" {
			return kv[1]
		}
	}
	return defaultName
}

func GetBranches(ctx context.Context, remotes []shared.Remote, connection shared.Connection, state shared.PullRequestState, scan shared.ScanMode, dryRun bool) ([]shared.
	Branch, error) {
	var repoNames []string
	var defaultBranchName string
	var err error
	if scan == shared.Quick {
		if repos, e := connection.GetRepoNames(ctx, remotes[0].Hostname, remotes[0].ResolvedRepoName()); e == nil {
			repoNames, defaultBranchName, err = getRepo(repos)
		} else {
			err = e
		}
	} else {
		uniqueRepoNames := make(map[string]bool)
		first := true
		for _, remote := range remotes {
			if repos, e := connection.GetRepoNames(ctx, remote.Hostname, remote.ResolvedRepoName()); e == nil {
				names, defaultName, e := getRepo(repos)
				if e != nil {
					err = e
					continue
				}
				for _, name := range names {
					uniqueRepoNames[name] = true
				}
				if first {
					defaultBranchName = defaultName
				}
			} else {
				err = e
				continue
			}
			first = false
		}
		for repoName := range uniqueRepoNames {
			repoNames = append(repoNames, repoName)
		}
	}
	if err != nil {
		return nil, err
	}

	branches, err := loadBranches(ctx, remotes[0], defaultBranchName, repoNames, connection, scan)
	if err != nil {
		return nil, err
	}

	branches = checkDeletion(branches, state)

	branches, err = switchToDefaultBranchIfDeleted(ctx, remotes, branches, defaultBranchName, connection, dryRun)
	if err != nil {
		return nil, err
	}

	sort.Slice(branches, func(i, j int) bool { return branches[i].Name < branches[j].Name })

	return branches, nil
}

func loadBranches(ctx context.Context, remote shared.Remote, defaultBranchName string, repoNames []string, connection shared.Connection, scan shared.ScanMode) ([]shared.Branch, error) {
	var branches []shared.Branch

	if names, err := connection.GetBranchNames(ctx); err == nil {
		branches = ToBranch(SplitLines(names))
		branches = applyDefault(branches, defaultBranchName)
		mergedNames, err := connection.GetMergedBranchNames(ctx, remote.Name, defaultBranchName)
		if err != nil {
			return nil, err
		}
		branches = applyMerged(branches, extractMergedBranchNames(SplitLines(mergedNames)))
		branches, err = applyLocked(ctx, branches, connection)
		if err != nil {
			return nil, err
		}
		branches, err = applyCommits(ctx, branches, defaultBranchName, connection, scan)
		if err != nil {
			return nil, err
		}
		branches, err = applyWorktrees(ctx, branches, connection)
		if err != nil {
			return nil, err
		}
		branches, err = applyTrackedChanges(ctx, branches, connection)
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
			pullRequests, err := connection.GetPullRequests(ctx, remote.Hostname, orgs, repos, hash)
			if err != nil {
				prChan <- pullRequestResult{err: err}
				return
			}

			pr, err := toPullRequests(pullRequests)
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

func applyLocked(ctx context.Context, branches []shared.Branch, connection shared.Connection) ([]shared.Branch, error) {
	results := []shared.Branch{}

	for _, branch := range branches {
		config, _ := connection.GetConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-locked", branch.Name))
		splitConfig := SplitLines(config)
		if len(splitConfig) > 0 && splitConfig[0] == "true" {
			branch.IsLocked = true
		}

		// TODO: Remove after deprecated commands are removed
		configDeprecated, _ := connection.GetConfig(ctx, fmt.Sprintf("branch.%s.gh-poi-protected", branch.Name))
		splitConfigDeprecated := SplitLines(configDeprecated)
		if len(splitConfigDeprecated) > 0 && splitConfigDeprecated[0] == "true" {
			branch.IsLocked = true
		}

		results = append(results, branch)
	}

	return results, nil
}

func applyCommits(ctx context.Context, branches []shared.Branch, defaultBranchName string, connection shared.Connection, scan shared.ScanMode) ([]shared.Branch, error) {
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

			oids, err := connection.GetLog(ctx, branch.Name)
			if err != nil {
				resultChan <- remoteBranchResult{err: err}
				return
			}

			if logOids := SplitLines(oids); len(logOids) > 0 {
				if scan == shared.Quick {
					branch.Commits = []string{logOids[0]}
				} else {
					trimmedOids, err := trimBranch(ctx, logOids, branch, defaultBranchName, connection)
					if err != nil {
						resultChan <- remoteBranchResult{err: err}
						return
					}
					branch.Commits = trimmedOids
				}
			} else {
				branch.Commits = []string{}
			}
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
	results := []shared.Branch{}

	for _, branch := range branches {
		changes := []shared.UncommittedChange{}
		var err error
		if branch.Head {
			changes, err = conn.GetUncommittedChanges(ctx, connection)
			if err != nil {
				return results, err
			}
		} else if branch.Worktree != nil {
			changes, err = conn.GetUncommittedChanges(ctx, connection, "-C", branch.Worktree.Path)
			if err != nil {
				return results, err
			}
		}

		for _, change := range changes {
			if change.IsUntracked() {
				branch.HasUntrackedFiles = true
			} else {
				branch.HasTrackedChanges = true
			}
		}
		results = append(results, branch)
	}

	return results, nil
}

func applyWorktrees(ctx context.Context, branches []shared.Branch, connection shared.Connection) ([]shared.Branch, error) {
	worktrees, err := conn.GetWorktrees(ctx, connection)
	if err != nil {
		// Worktrees might not be supported or available, continue gracefully
		return branches, nil
	}

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

func trimBranch(ctx context.Context, oids []string, branch shared.Branch, defaultBranchName string, connection shared.Connection) ([]string, error) {
	results := []string{}
	childNames := []string{}

	for i, oid := range oids {
		if branch.IsMerged {
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
				if name != branch.Name {
					childNames = append(childNames, name)
				}
			}
		}

		for _, name := range names {
			if name != branch.Name && !slices.Contains(childNames, name) {
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
			if pr.Url == result.Url && pr.Number == result.Number {
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

func checkDeletion(branches []shared.Branch, state shared.PullRequestState) []shared.Branch {
	results := []shared.Branch{}
	for _, branch := range branches {
		branch.State = getDeleteStatus(branch, state)
		results = append(results, branch)
	}
	return results
}

func getDeleteStatus(branch shared.Branch, state shared.PullRequestState) shared.BranchState {
	if branch.IsLocked {
		return shared.NotDeletable
	}

	if branch.Worktree != nil {
		if branch.Worktree.IsLocked || (branch.Worktree.IsMain && !branch.Head) || (!branch.Worktree.IsMain && branch.Head) || branch.HasUntrackedFiles {
			return shared.NotDeletable
		}
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
	if slices.Contains(pr.Commits, localHeadOid) {
		return true
	}

	return false
}

func switchToDefaultBranchIfDeleted(ctx context.Context, remotes []shared.Remote, branches []shared.Branch, defaultBranchName string, connection shared.Connection, dryRun bool) ([]shared.Branch, error) {
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

	var remoteName = remotes[0].Name
	for _, remote := range remotes {
		if remote.GhResolved != "" {
			remoteName = remote.Name
		}
	}
	newBranchName := defaultBranchName
	if BranchNameExists(defaultBranchName, branches) {
		if !dryRun {
			_, err := connection.CheckoutBranch(ctx, newBranchName, false)
			if err != nil {
				return nil, err
			}
		}
	} else {
		newBranchName = remoteName + "/" + defaultBranchName
		if !dryRun {
			_, err := connection.CheckoutBranch(ctx, newBranchName, true)
			if err != nil {
				return nil, err
			}
		}
	}

	if !BranchNameExists(defaultBranchName, branches) {
		branch := shared.Branch{}
		branch.Head = true
		branch.Name = defaultBranchName
		if newBranchName != defaultBranchName {
			branch.Name = "(HEAD detached at " + remoteName + "/" + defaultBranchName + ")"
		}
		branch.State = shared.NotDeletable
		results = append(results, branch)
	}

	for _, branch := range branches {
		if branch.Name == newBranchName {
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

	_, err := deleteWorktrees(ctx, branches, connection)
	if err != nil {
		return nil, err
	}
	_, err = connection.DeleteBranches(ctx, branchNames)
	if err != nil {
		return nil, err
	}

	branchNamesAfter, err := connection.GetBranchNames(ctx)
	if err != nil {
		return nil, err
	}
	branchesAfter := ToBranch(SplitLines(branchNamesAfter))

	return checkDeleted(branches, branchesAfter), nil
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

func deleteWorktrees(ctx context.Context, branches []shared.Branch, connection shared.Connection) (map[string]bool, error) {
	deleted := make(map[string]bool)
	var errs []error
	for _, branch := range branches {
		if branch.State != shared.Deletable {
			continue
		}
		if branch.Worktree == nil || branch.Worktree.IsMain {
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
	return strings.FieldsFunc(strings.ReplaceAll(text, "\r\n", "\n"),
		func(c rune) bool { return c == '\n' })
}
