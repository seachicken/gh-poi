package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/seachicken/gh-poi/cmd"
	"github.com/seachicken/gh-poi/cmd/lock"
	"github.com/seachicken/gh-poi/conn"
	"github.com/seachicken/gh-poi/shared"
)

var (
	bold    = color.New(color.Bold).SprintFunc()
	hiBlack = color.New(color.FgHiBlack).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
)

type StateFlag string

const (
	Closed StateFlag = "closed"
	Merged StateFlag = "merged"
)

func (s *StateFlag) String() string {
	return string(*s)
}

func (s *StateFlag) Set(value string) error {
	for _, state := range []StateFlag{Closed, Merged} {
		if value == string(state) {
			*s = StateFlag(value)
			return nil
		}
	}
	return errors.New("invalid state")
}

func (s StateFlag) toModel() shared.PullRequestState {
	switch s {
	case Closed:
		return shared.Closed
	default:
		return shared.Merged
	}
}

type ScanFlag string

const (
	Quick ScanFlag = "quick"
	Deep  ScanFlag = "deep"
)

func (s *ScanFlag) String() string {
	return string(*s)
}

func (s *ScanFlag) Set(value string) error {
	for _, mode := range []ScanFlag{Quick, Deep} {
		if value == string(mode) {
			*s = ScanFlag(value)
			return nil
		}
	}
	return errors.New("invalid scan mode")
}

func (s ScanFlag) toModel() shared.ScanMode {
	switch s {
	case Deep:
		return shared.Deep
	default:
		return shared.Quick
	}
}

func main() {
	state := Merged
	scan := Quick
	var dryRun bool
	var debug bool
	flag.Var(&state, "state", "Specify the PR state to delete by {closed|merged}")
	flag.Var(&scan, "scan", "Specify the scan mode by {quick|deep}")
	flag.BoolVar(&dryRun, "dry-run", false, "Show branches to delete without actually deleting it")
	flag.BoolVar(&debug, "debug", false, "Enable debug logs")
	flag.Usage = func() {
		fmt.Fprintf(color.Output, "%s\n\n", "Delete merged local branches.")
		fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
		fmt.Fprintf(color.Output, "  %s\n\n", "gh poi <command> [flags]")
		fmt.Fprintf(color.Output, "%s", bold("COMMANDS"))
		fmt.Fprintf(color.Output, "%s\n", `
  branches (default)  Delete merged local branches
  forks               Delete GitHub forks fully merged into upstream
  lock                Lock branches to prevent them from being deleted
  unlock              Unlock branches to allow them to be deleted
  protect             (Deprecated) use 'lock' instead
  unprotect           (Deprecated) use 'unlock' instead
  `)
		fmt.Fprintf(color.Output, "%s\n", bold("FLAGS"))
		maxLen := 0
		flag.VisitAll(func(f *flag.Flag) {
			if len(f.Name) > maxLen {
				maxLen = len(f.Name)
			}
		})
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(color.Output, "  --%-*s %s\n", maxLen+2, f.Name, f.Usage)
		})
		fmt.Println()
	}
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		runMain(state, scan, dryRun, debug)
	} else {
		subcmd, args := args[0], args[1:]
		switch subcmd {
		case "branches":
			branchesCmd := flag.NewFlagSet("branches", flag.ExitOnError)
			var recursive bool
			branchesCmd.BoolVar(&recursive, "recursive", false, "Includes ALL repos found by recursively scanning [directory]")
			branchesCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", "Delete merged local branches. Defaults to current directory.")
				fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", "gh poi branches [directory] [flags]")
				fmt.Fprintf(color.Output, "%s\n", bold("FLAGS"))
				fmt.Fprintf(color.Output, "  --%-*s %s\n", 10, "recursive", "Includes ALL repos found by recursively scanning [directory] (default false)")
			}
			branchesCmd.Parse(args)

			branchesDir := ""
			remaining := branchesCmd.Args()
			if len(remaining) > 0 {
				branchesDir = remaining[0]
			}

			if recursive {
				if branchesDir == "" {
					branchesDir = "."
				}
				runRepos(branchesDir, state, scan, dryRun, debug)
			} else {
				if branchesDir != "" {
					origDir, _ := os.Getwd()
					if err := os.Chdir(branchesDir); err != nil {
						fmt.Fprintf(os.Stderr, "error: %v\n", err)
						os.Exit(1)
					}
					defer os.Chdir(origDir)
				}
				runMain(state, scan, dryRun, debug)
			}

		case "forks":
			forksCmd := flag.NewFlagSet("forks", flag.ExitOnError)
			var concurrency int
			forksCmd.IntVar(&concurrency, "concurrency", 20, "Number of repos to scan in parallel")
			forksCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", "Delete GitHub forks whose branches are all merged into upstream. Recursively scans [directory] for local clones with unpushed work before allowing deletion. Defaults to current directory.")
				fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", "gh poi forks [directory] [flags]")
				fmt.Fprintf(color.Output, "%s\n", bold("FLAGS"))
				fmt.Fprintf(color.Output, "  --%-*s %s\n", 10, "concurrency", "Number of repos to scan in parallel (default 20)")
			}
			forksCmd.Parse(args)

			forksDir := "."
			remaining := forksCmd.Args()
			if len(remaining) > 0 {
				forksDir = remaining[0]
			}
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			connection := &conn.Connection{Debug: debug}
			runForks(ctx, connection, forksDir, state, dryRun, debug, concurrency)
		case "lock", "protect":
			lockCmd := flag.NewFlagSet("lock", flag.ExitOnError)
			lockCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", "Lock branches to prevent them from being deleted")
				fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", "gh poi lock <branchname>...")
			}
			lockCmd.Parse(args)

			// TODO: Remove after deprecated commands are removed
			if subcmd == "protect" {
				fmt.Fprintln(os.Stderr, "warning: 'protect' is deprecated, please use 'lock' instead")
			}
			runLock(args, debug)
		case "unlock", "unprotect":
			unlockCmd := flag.NewFlagSet("unlock", flag.ExitOnError)
			unlockCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", "Unlock branches to allow them to be deleted")
				fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", "gh poi unlock <branchname>...")
			}
			unlockCmd.Parse(args)

			// TODO: Remove after deprecated commands are removed
			if subcmd == "unprotect" {
				fmt.Fprintln(os.Stderr, "warning: 'unprotect' is deprecated, please use 'unlock' instead")
			}
			runUnlock(args, debug)
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q for poi\n", subcmd)
		}
	}
}

func runMain(state StateFlag, scan ScanFlag, dryRun bool, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if dryRun {
		fmt.Fprintf(color.Output, "%s\n", bold("== DRY RUN =="))
	}

	connection := &conn.Connection{Debug: debug}
	sp := spinner.New(spinner.CharSets[14], 40*time.Millisecond)
	defer sp.Stop()

	fetchingMsg := " Fetching pull requests..."
	sp.Suffix = fetchingMsg
	if !debug {
		sp.Start()
	}
	var fetchingErr error

	remotes, err := cmd.GetPreferredRemotes(ctx, connection, scan.toModel())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	branches, fetchingErr := cmd.GetBranches(ctx, remotes, connection, state.toModel(), scan.toModel(), dryRun)

	sp.Stop()

	if fetchingErr == nil {
		fmt.Fprintf(color.Output, "%s%s\n", green("✔"), fetchingMsg)
	} else {
		fmt.Fprintf(color.Output, "%s%s\n", red("✕"), fetchingMsg)
		fmt.Fprintln(os.Stderr, fetchingErr)
		return
	}

	deletingMsg := " Deleting branches..."

	if dryRun {
		fmt.Fprintf(color.Output, "%s%s\n", hiBlack("-"), deletingMsg)
	} else {
		sp.Suffix = deletingMsg
		if !debug {
			sp.Restart()
		}

		var deletingErr error
		branches, deletingErr = cmd.DeleteBranches(ctx, branches, connection)
		connection.PruneRemoteBranches(ctx, remotes[0].Name)
		connection.PruneWorktrees(ctx)

		sp.Stop()

		if deletingErr == nil {
			fmt.Fprintf(color.Output, "%s%s\n", green("✔"), deletingMsg)
		} else {
			fmt.Fprintf(color.Output, "%s%s\n", red("✕"), deletingMsg)
			fmt.Fprintln(os.Stderr, deletingErr)
			return
		}
	}

	fmt.Println()

	var deletedStates []shared.BranchState
	var notDeletedStates []shared.BranchState
	if dryRun {
		deletedStates = []shared.BranchState{shared.Deletable}
		notDeletedStates = []shared.BranchState{shared.NotDeletable}
	} else {
		deletedStates = []shared.BranchState{shared.Deleted}
		notDeletedStates = []shared.BranchState{shared.Deletable, shared.NotDeletable}
	}

	fmt.Fprintf(color.Output, "%s\n", bold("Deleted branches"))
	printBranches(getBranches(branches, deletedStates))
	fmt.Println()

	fmt.Fprintf(color.Output, "%s\n", bold("Branches not deleted"))
	printBranches(getBranches(branches, notDeletedStates))
	fmt.Println()

	if cmd.IsRepoDeletable(branches) {
		repoRoot, err := connection.GetRepoRoot(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		repoRoot = strings.TrimSpace(repoRoot)

		if dryRun {
			fmt.Fprintf(color.Output, "%s\n", bold("Repo is deletable"))
			fmt.Fprintf(color.Output, "  All branches are merged into the upstream.\n")
			fmt.Fprintf(color.Output, "  %s\n", hiBlack(fmt.Sprintf("Would delete: %s", repoRoot)))
		} else {
			fmt.Fprintf(color.Output, "%s\n", bold("Repo is deletable"))
			fmt.Fprintf(color.Output, "  All branches are merged into the upstream.\n")
			fmt.Fprintf(color.Output, "  Delete repo directory %s? [y/N] ", repoRoot)

			drainStdin()

			reader := bufio.NewReader(os.Stdin)
			input, err := readInput(ctx, reader)
			if err != nil {
				return
			}
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "y" || input == "yes" {
				if err := os.RemoveAll(repoRoot); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
		}
	}
}

func runLock(branchNames []string, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &conn.Connection{Debug: debug}

	err := lock.LockBranches(ctx, branchNames, connection)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func runUnlock(branchNames []string, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &conn.Connection{Debug: debug}

	err := lock.UnlockBranches(ctx, branchNames, connection)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func printBranches(branches []shared.Branch) {
	if len(branches) == 0 {
		fmt.Fprintf(color.Output, "%s\n",
			hiBlack("  There are no branches in the current directory"))
	}

	for _, branch := range branches {
		if branch.Head {
			fmt.Fprintf(color.Output, "* %s", green(branch.Name))
		} else {
			fmt.Fprintf(color.Output, "  %s", branch.Name)
		}

		// Show worktree info for any branch with an associated worktree
		if branch.Worktree != nil && !branch.Worktree.IsMain {
			fmt.Fprintf(color.Output, " %s", hiBlack("(worktree: "+branch.Worktree.Path+")"))
		}

		reason := ""
		if branch.State == shared.NotDeletable {
			if branch.IsLocked {
				reason = "locked"
			} else if branch.Worktree != nil && branch.Worktree.IsLocked {
				reason = "worktree locked"
			} else if branch.Worktree != nil && branch.Worktree.IsMain && !branch.Head {
				reason = "main worktree"
			} else if branch.Worktree != nil && !branch.Worktree.IsMain && branch.Head {
				reason = "worktree here"
			} else if branch.Worktree != nil && branch.HasUntrackedFiles {
				reason = "untracked files"
			} else if !branch.IsDefault && len(branch.PullRequests) > 0 && branch.HasTrackedChanges {
				reason = "uncommitted changes"
			}
		}
		if reason == "" {
			fmt.Fprintln(color.Output, "")
		} else {
			fmt.Fprintf(color.Output, " %s\n", hiBlack("["+reason+"]"))
		}

		for i, pr := range branch.PullRequests {
			number := fmt.Sprintf("#%v", pr.Number)
			issueNoColor := getIssueNoColor(pr.State, pr.IsDraft)
			var line string
			if i == len(branch.PullRequests)-1 {
				line = "└─"
			} else {
				line = "├─"
			}

			fmt.Fprintf(color.Output, "    %s %s  %s %s\n",
				line,
				color.New(issueNoColor).SprintFunc()(number),
				pr.Url,
				hiBlack(pr.Author),
			)
		}
	}
}

func getIssueNoColor(state shared.PullRequestState, isDraft bool) color.Attribute {
	switch state {
	case shared.Open:
		if isDraft {
			return color.FgHiBlack
		} else {
			return color.FgGreen
		}
	case shared.Merged:
		return color.FgMagenta
	case shared.Closed:
		return color.FgRed
	default:
		return color.FgHiBlack
	}
}

func getBranches(branches []shared.Branch, states []shared.BranchState) []shared.Branch {
	results := []shared.Branch{}
	for _, branch := range branches {
		if slices.Contains(states, branch.State) {
			results = append(results, branch)
		}
	}
	return results
}

// drainStdin discards any pending typeahead from stdin that may have been
// accidentally entered while a spinner or long operation was running.
// This prevents stale input from interfering with subsequent prompts.
func drainStdin() {
	fd := int(os.Stdin.Fd())
	// Set stdin to non-blocking mode, drain any pending data, then restore.
	if err := syscall.SetNonblock(fd, true); err != nil {
		return
	}
	defer syscall.SetNonblock(fd, false)

	var buf [4096]byte
	for {
		_, err := os.Stdin.Read(buf[:])
		if err != nil {
			break
		}
	}
}

func readInput(ctx context.Context, reader *bufio.Reader) (string, error) {
	type result struct {
		val string
		err error
	}
	ch := make(chan result)
	go func() {
		input, err := reader.ReadString('\n')
		ch <- result{input, err}
	}()
	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

type RepoResult struct {
	Path      string
	Branches  []shared.Branch
	Deletable bool
	Error     error
}

func analyzeRepo(ctx context.Context, dir string, state shared.PullRequestState, scan shared.ScanMode, dryRun bool, debug bool) RepoResult {
	connection := &conn.Connection{Debug: debug}

	remotes, err := cmd.GetPreferredRemotes(ctx, connection, scan)
	if err != nil {
		return RepoResult{Path: dir, Error: err}
	}

	branches, err := cmd.GetBranches(ctx, remotes, connection, state, scan, dryRun)
	if err != nil {
		return RepoResult{Path: dir, Error: err}
	}

	if !dryRun {
		branches, err = cmd.DeleteBranches(ctx, branches, connection)
		if err != nil {
			return RepoResult{Path: dir, Branches: branches, Error: err}
		}
		connection.PruneRemoteBranches(ctx, remotes[0].Name)
	}

	return RepoResult{
		Path:      dir,
		Branches:  branches,
		Deletable: cmd.IsRepoDeletable(branches),
	}
}


func runRepos(directory string, state StateFlag, scan ScanFlag, dryRun bool, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if dryRun {
		fmt.Fprintf(color.Output, "%s\n", bold("== DRY RUN =="))
	}

	absDir, _ := filepath.Abs(directory)
	fmt.Fprintf(color.Output, "Scanning %s for git repositories...\n", absDir)

	repos, err := cmd.FindGitRepos(directory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if len(repos) == 0 {
		fmt.Fprintf(color.Output, "  %s\n", hiBlack("No git repositories found"))
		return
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	sp := spinner.New(spinner.CharSets[14], 40*time.Millisecond)
	defer sp.Stop()

	// Repos are processed sequentially here because os.Chdir is process-wide.
	// Parallelizing would require changing the Connection interface to pass working
	// directories per-command (e.g. using exec.Cmd.Dir or git -C).
	var results []RepoResult
	for i, repo := range repos {
		sp.Suffix = fmt.Sprintf(" Analyzing %d/%d repositories...", i+1, len(repos))
		if !debug {
			sp.Start()
		}

		os.Chdir(repo)
		result := analyzeRepo(ctx, repo, state.toModel(), scan.toModel(), dryRun, debug)
		results = append(results, result)

		os.Chdir(origDir)
	}

	sp.Stop()
	fmt.Fprintf(color.Output, "%s Analyzed %d repositories\n\n", green("✔"), len(repos))

	var deletableResults []RepoResult
	var errorResults []RepoResult
	for _, r := range results {
		if r.Error != nil {
			errorResults = append(errorResults, r)
		} else if r.Deletable {
			deletableResults = append(deletableResults, r)
		}
	}

	if len(errorResults) > 0 {
		fmt.Fprintf(color.Output, "%s\n", bold("Errors"))
		for _, r := range errorResults {
			fmt.Fprintf(color.Output, "  %s %s\n", red("✕"), r.Path)
			fmt.Fprintf(color.Output, "    %s\n", hiBlack(r.Error.Error()))
		}
		fmt.Println()
	}

	if len(deletableResults) == 0 {
		fmt.Fprintf(color.Output, "%s\n", hiBlack("No deletable repositories found"))
		return
	}

	fmt.Fprintf(color.Output, "%s\n", bold("Deletable repositories"))
	for _, r := range deletableResults {
		fmt.Fprintf(color.Output, "  %s\n", r.Path)
		fmt.Fprintf(color.Output, "    %s\n", green("✔ All branches merged into upstream"))
	}
	fmt.Println()

	if dryRun {
		return
	}

	drainStdin()

	reader := bufio.NewReader(os.Stdin)
	for _, r := range deletableResults {
		fmt.Fprintf(color.Output, "Delete %s? [y/N] ", r.Path)
		input, err := readInput(ctx, reader)
		if err != nil {
			return
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			if err := os.RemoveAll(r.Path); err != nil {
				fmt.Fprintf(color.Output, "  %s %s\n", red("✕"), err)
			} else {
				fmt.Fprintf(color.Output, "  %s Deleted\n", green("✔"))
			}
		} else {
			fmt.Fprintf(color.Output, "  %s\n", hiBlack("Skipped"))
		}
	}
}


type GitHubRepoResult struct {
	Owner     string
	Repo      string
	IsFork    bool
	Deletable bool
	LocalPath string
	Reason    string
	Error     error
}

func runForks(ctx context.Context, connection *conn.Connection, directory string, state StateFlag, dryRun bool, debug bool, concurrency int) {

	viewerLogin, err := connection.GetViewerLogin(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	viewerLogin = strings.TrimSpace(viewerLogin)

	authScopesOutput, _ := connection.GetAuthScopes(ctx)
	scopes := cmd.ParseAuthScopes(authScopesOutput)
	canDeleteForks := cmd.HasDeleteRepoScope(scopes)
	if !canDeleteForks {
		fmt.Fprintf(color.Output, "%s Fork deletion requires delete_repo scope. Run: gh auth refresh -s delete_repo\n\n", hiBlack("ℹ"))
	}

	userReposJSON, err := connection.GetUserRepos(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	allRepos, err := cmd.ParseUserReposDetailed(userReposJSON)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if len(allRepos) == 0 {
		fmt.Fprintf(color.Output, "%s\n", hiBlack("No GitHub repositories found"))
		return
	}

	// Filter to forks only; non-forks are not deletable by this command
	var forkRepos []cmd.UserRepoInfo
	for _, r := range allRepos {
		if r.IsFork {
			forkRepos = append(forkRepos, r)
		}
	}
	if len(forkRepos) == 0 {
		fmt.Fprintf(color.Output, "%s\n", hiBlack("No fork repositories found"))
		return
	}

	// Build a map of local clones found in the given directory
	localRepoMap := make(map[string]string)
	localRepos, err := cmd.FindGitRepos(directory)
	if err == nil {
		origDir, _ := os.Getwd()
		for _, repo := range localRepos {
			os.Chdir(repo)
			remotes, err := conn.GetRemoteNames(ctx, connection)
			os.Chdir(origDir)
			if err != nil {
				continue
			}
			for _, remote := range remotes {
				localRepoMap[remote.RepoName] = repo
			}
		}
	}

	sp := spinner.New(spinner.CharSets[14], 40*time.Millisecond)
	defer sp.Stop()

	type scanResult struct {
		Result GitHubRepoResult
	}

	resultCh := make(chan scanResult, len(forkRepos))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	sp.Suffix = fmt.Sprintf(" Checking 0/%d repositories...", len(forkRepos))
	if !debug {
		sp.Start()
	}

	for _, fork := range forkRepos {
		wg.Add(1)
		go func(fork cmd.UserRepoInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			parts := strings.SplitN(fork.NameWithOwner, "/", 2)
			forkOwner, forkRepo := parts[0], parts[1]
			localPath := localRepoMap[fork.NameWithOwner]

			// If a local clone exists, check it for unpushed work before
			// consulting the GitHub API. Local-only changes would be
			// invisible to the API check, so we catch them here.
			if localPath != "" {
				hasChanges, err := hasLocalChanges(ctx, localPath)
				if err == nil && hasChanges {
					resultCh <- scanResult{Result: GitHubRepoResult{
						Owner:     forkOwner,
						Repo:      forkRepo,
						IsFork:    fork.IsFork,
						Deletable: false,
						LocalPath: localPath,
						Reason:    "local clone has unpushed changes",
					}}
					return
				}
			}

			parentDefaultBranch := ""
			if fork.ParentOwner != "" && fork.ParentRepo != "" {
				_, parentDefault, _, err := cmd.GetRepo(ctx, connection, []shared.Remote{{Hostname: "github.com", RepoName: fork.ParentOwner + "/" + fork.ParentRepo}})
				if err != nil {
					resultCh <- scanResult{Result: GitHubRepoResult{Owner: forkOwner, Repo: forkRepo, IsFork: fork.IsFork, LocalPath: localPath, Error: err}}
					return
				}
				parentDefaultBranch = parentDefault
			}

			deletable, err := cmd.IsRepoDeletableAPI(ctx, connection, forkOwner, forkRepo, fork.DefaultBranch, parentDefaultBranch, fork.IsFork, fork.ParentOwner, fork.ParentRepo, viewerLogin)
			if err != nil {
				resultCh <- scanResult{Result: GitHubRepoResult{Owner: forkOwner, Repo: forkRepo, IsFork: fork.IsFork, LocalPath: localPath, Error: err}}
				return
			}

			resultCh <- scanResult{Result: GitHubRepoResult{
				Owner:     forkOwner,
				Repo:      forkRepo,
				IsFork:    fork.IsFork,
				Deletable: deletable,
				LocalPath: localPath,
			}}
		}(fork)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	var results []GitHubRepoResult
	completed := 0
	total := len(forkRepos)
	running := true

	for running {
		select {
		case rr, ok := <-resultCh:
			if !ok {
				running = false
				continue
			}
			completed++
			sp.Suffix = fmt.Sprintf(" Checking %d/%d repositories...", completed, total)
			if rr.Result.Owner != "" || rr.Result.Error != nil {
				results = append(results, rr.Result)
			}
		case <-ticker.C:
			sp.Suffix = fmt.Sprintf(" Checking %d/%d repositories... (still working)", completed, total)
		}
	}
	slices.SortFunc(results, func(a, b GitHubRepoResult) int {
		if a.Owner != b.Owner {
			if a.Owner < b.Owner {
				return -1
			}
			return 1
		}
		if a.Repo != b.Repo {
			if a.Repo < b.Repo {
				return -1
			}
			return 1
		}
		return 0
	})

	sp.Stop()
	fmt.Fprintf(color.Output, "%s Checked %d repositories\n\n", green("✔"), len(forkRepos))

	var deletableResults []GitHubRepoResult
	var notDeletableResults []GitHubRepoResult
	var errorResults []GitHubRepoResult
	for _, r := range results {
		if r.Error != nil {
			errorResults = append(errorResults, r)
		} else if r.Deletable {
			deletableResults = append(deletableResults, r)
		} else {
			notDeletableResults = append(notDeletableResults, r)
		}
	}

	if len(errorResults) > 0 {
		fmt.Fprintf(color.Output, "%s\n", bold("Errors"))
		for _, r := range errorResults {
			fmt.Fprintf(color.Output, "  %s %s/%s\n", red("✕"), r.Owner, r.Repo)
			fmt.Fprintf(color.Output, "    %s\n", hiBlack(r.Error.Error()))
		}
		fmt.Println()
	}

	if len(notDeletableResults) > 0 {
		fmt.Fprintf(color.Output, "%s\n", bold("Non-deletable repositories"))
		for _, r := range notDeletableResults {
			label := fmt.Sprintf("%s/%s", r.Owner, r.Repo)
			if r.LocalPath != "" {
				label = fmt.Sprintf("%s/%s (%s)", r.Owner, r.Repo, r.LocalPath)
			}
			reason := r.Reason
			if reason == "" {
				reason = "has open PRs or branches ahead of upstream"
			}
			fmt.Fprintf(color.Output, "  %s %s\n", hiBlack("•"), label)
			fmt.Fprintf(color.Output, "    %s\n", hiBlack(reason))
		}
		fmt.Println()
	}

	if len(deletableResults) == 0 {
		fmt.Fprintf(color.Output, "%s\n", hiBlack("No deletable repositories found"))
		return
	}

	fmt.Fprintf(color.Output, "%s\n", bold("Deletable repositories"))
	for _, r := range deletableResults {
		label := fmt.Sprintf("%s/%s", r.Owner, r.Repo)
		if r.LocalPath != "" {
			label = fmt.Sprintf("%s/%s (%s)", r.Owner, r.Repo, r.LocalPath)
		}
		fmt.Fprintf(color.Output, "  %s\n", label)
		fmt.Fprintf(color.Output, "    %s\n", green("✔ All branches merged into upstream"))
	}
	fmt.Println()

	if dryRun {
		return
	}

	drainStdin()

	reader := bufio.NewReader(os.Stdin)
	for _, r := range deletableResults {
		if !canDeleteForks {
			fmt.Fprintf(color.Output, "  %s/%s: %s\n", r.Owner, r.Repo, hiBlack("Skipping fork deletion: missing delete_repo scope. Run: gh auth refresh -s delete_repo"))
			continue
		}

		label := fmt.Sprintf("%s/%s", r.Owner, r.Repo)
		fmt.Fprintf(color.Output, "Delete GitHub fork https://github.com/%s ? [y/N] ", label)
		input, err := readInput(ctx, reader)
		if err != nil {
			return
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			_, err := connection.DeleteGitHubRepo(ctx, r.Owner, r.Repo)
			if err != nil {
				fmt.Fprintf(color.Output, "  %s %s\n", red("✕"), err)
			} else {
				fmt.Fprintf(color.Output, "  %s GitHub fork deleted\n", green("✔"))
				// If we know about a local clone, offer to clean that up too
				if r.LocalPath != "" {
					fmt.Fprintf(color.Output, "Delete local clone %s? [y/N] ", r.LocalPath)
					input, err := readInput(ctx, reader)
					if err != nil {
						return
					}
					input = strings.TrimSpace(strings.ToLower(input))
					if input == "y" || input == "yes" {
						if err := os.RemoveAll(r.LocalPath); err != nil {
							fmt.Fprintf(color.Output, "  %s %s\n", red("✕"), err)
						} else {
							fmt.Fprintf(color.Output, "  %s Local clone deleted\n", green("✔"))
						}
					} else {
						fmt.Fprintf(color.Output, "  %s\n", hiBlack("Skipped"))
					}
				}
			}
		} else {
			fmt.Fprintf(color.Output, "  %s\n", hiBlack("Skipped"))
		}
	}
}

// hasLocalChanges checks whether a git repo at repoPath has any unpushed
// commits, dirty worktree, or untracked files. Used by runForks to guard
// against deleting a fork whose local clone would lose work invisible to
// the GitHub API.
func hasLocalChanges(ctx context.Context, repoPath string) (bool, error) {
	// Check for dirty worktree or untracked files
	out, err := exec.CommandContext(ctx, "git", "-C", repoPath, "status", "--porcelain").Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return true, nil
	}

	// Check for commits on local branches not present on any remote tracking branch
	out, err = exec.CommandContext(ctx, "git", "-C", repoPath, "log", "--oneline", "--branches", "--not", "--remotes").Output()
	if err != nil {
		return false, fmt.Errorf("failed to check unpushed commits: %w", err)
	}
	if len(strings.TrimSpace(string(out))) > 0 {
		return true, nil
	}

	return false, nil
}
