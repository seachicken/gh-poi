package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"slices"
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
	if err := runCLI(os.Args[1:]); err != nil {
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if code := exitErr.ExitCode(); code != 0 {
			return code
		}
	}
	return 1
}

func runCLI(args []string) error {
	state := Merged
	scan := Quick
	var dryRun bool
	var debug bool

	flags := flag.NewFlagSet("gh-poi", flag.ContinueOnError)
	flags.SetOutput(color.Output)
	flags.Var(&state, "state", "Specify the PR state to delete by {closed|merged}")
	flags.Var(&scan, "scan", "Specify the scan mode by {quick|deep}")
	flags.BoolVar(&dryRun, "dry-run", false, "Show branches to delete without actually deleting it")
	flags.BoolVar(&debug, "debug", false, "Enable debug logs")
	flags.Usage = func() {
		fmt.Fprintf(color.Output, "%s\n\n", "Delete the merged local branches.")
		fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
		fmt.Fprintf(color.Output, "  %s\n\n", "gh poi <command> [flags]")
		fmt.Fprintf(color.Output, "%s", bold("COMMANDS"))
		fmt.Fprintf(color.Output, "%s\n", `
  lock:      Lock branches to prevent them from being deleted
  unlock:    Unlock branches to allow them to be deleted
  protect:   (Deprecated) use 'lock' instead
  unprotect: (Deprecated) use 'unlock' instead
  `)
		fmt.Fprintf(color.Output, "%s\n", bold("FLAGS"))
		maxLen := 0
		flags.VisitAll(func(f *flag.Flag) {
			if len(f.Name) > maxLen {
				maxLen = len(f.Name)
			}
		})
		flags.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(color.Output, "  --%-*s %s\n", maxLen+2, f.Name, f.Usage)
		})
		fmt.Println()
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	remaining := flags.Args()

	if len(remaining) == 0 {
		return runMain(state, scan, dryRun, debug)
	}

	subcmd, subArgs := remaining[0], remaining[1:]
	switch subcmd {
	case "lock", "protect":
		lockCmd := flag.NewFlagSet("lock", flag.ContinueOnError)
		lockCmd.SetOutput(color.Output)
		lockCmd.Usage = func() {
			fmt.Fprintf(color.Output, "%s\n\n", "Lock branches to prevent them from being deleted")
			fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
			fmt.Fprintf(color.Output, "  %s\n\n", "gh poi lock <branchname>...")
		}
		if err := lockCmd.Parse(subArgs); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}

		// TODO: Remove after deprecated commands are removed
		if subcmd == "protect" {
			fmt.Fprintln(os.Stderr, "warning: 'protect' is deprecated, please use 'lock' instead")
		}
		return runLock(lockCmd.Args(), debug)
	case "unlock", "unprotect":
		unlockCmd := flag.NewFlagSet("unlock", flag.ContinueOnError)
		unlockCmd.SetOutput(color.Output)
		unlockCmd.Usage = func() {
			fmt.Fprintf(color.Output, "%s\n\n", "Unlock branches to allow them to be deleted")
			fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
			fmt.Fprintf(color.Output, "  %s\n\n", "gh poi unlock <branchname>...")
		}
		if err := unlockCmd.Parse(subArgs); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}

		// TODO: Remove after deprecated commands are removed
		if subcmd == "unprotect" {
			fmt.Fprintln(os.Stderr, "warning: 'unprotect' is deprecated, please use 'unlock' instead")
		}
		return runUnlock(unlockCmd.Args(), debug)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q for poi\n", subcmd)
		return fmt.Errorf("unknown command %q for poi", subcmd)
	}
}

type mainConnection interface {
	shared.Connection
	PruneRemoteBranches(ctx context.Context, remoteName string) (string, error)
	PruneWorktrees(ctx context.Context) (string, error)
}

func runMain(state StateFlag, scan ScanFlag, dryRun bool, debug bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &conn.Connection{Debug: debug}
	return runMainWithConn(ctx, connection, state, scan, dryRun, debug)
}

func runMainWithConn(ctx context.Context, connection mainConnection, state StateFlag, scan ScanFlag, dryRun bool, debug bool) error {
	if dryRun {
		fmt.Fprintf(color.Output, "%s\n", bold("== DRY RUN =="))
	}

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
		sp.Stop()
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	branches, fetchingErr := cmd.GetBranches(ctx, remotes, connection, state.toModel(), scan.toModel(), dryRun)

	sp.Stop()

	if fetchingErr == nil {
		fmt.Fprintf(color.Output, "%s%s\n", green("✔"), fetchingMsg)
	} else {
		fmt.Fprintf(color.Output, "%s%s\n", red("✕"), fetchingMsg)
		fmt.Fprintln(os.Stderr, fetchingErr)
		return fetchingErr
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

		if deletingErr == nil {
			if _, err := connection.PruneRemoteBranches(ctx, remotes[0].Name); err != nil {
				deletingErr = err
			}
		}
		if deletingErr == nil {
			if _, err := connection.PruneWorktrees(ctx); err != nil {
				deletingErr = err
			}
		}

		sp.Stop()

		if deletingErr == nil {
			fmt.Fprintf(color.Output, "%s%s\n", green("✔"), deletingMsg)
		} else {
			fmt.Fprintf(color.Output, "%s%s\n", red("✕"), deletingMsg)
			fmt.Fprintln(os.Stderr, deletingErr)
			return deletingErr
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

	return nil
}

func runLock(branchNames []string, debug bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &conn.Connection{Debug: debug}
	return runLockWithConn(ctx, connection, branchNames)
}

func runLockWithConn(ctx context.Context, connection shared.Connection, branchNames []string) error {
	err := lock.LockBranches(ctx, branchNames, connection)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func runUnlock(branchNames []string, debug bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &conn.Connection{Debug: debug}
	return runUnlockWithConn(ctx, connection, branchNames)
}

func runUnlockWithConn(ctx context.Context, connection shared.Connection, branchNames []string) error {
	err := lock.UnlockBranches(ctx, branchNames, connection)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
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
