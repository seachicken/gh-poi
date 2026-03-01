package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
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

func main() {
	state := Merged
	var dryRun bool
	var debug bool
	flag.Var(&state, "state", "Specify the PR state to delete by {closed|merged}")
	flag.BoolVar(&dryRun, "dry-run", false, "Show branches to delete without actually deleting it")
	flag.BoolVar(&debug, "debug", false, "Enable debug logs")
	flag.Usage = func() {
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
		runMain(state, dryRun, debug)
	} else {
		subcmd, args := args[0], args[1:]
		switch subcmd {
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
				fmt.Fprintln(os.Stderr, shared.ProtectDeprecationMsg)
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
				fmt.Fprintln(os.Stderr, shared.UnprotectDeprecationMsg)
			}
			runUnlock(args, debug)
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q for poi\n", subcmd)
		}
	}
}

func runWithRepoCheck(ctx context.Context, debug bool, fn func(context.Context, shared.Connection) error) {
	connection := &conn.Connection{Debug: debug}

	// Check repository presence upfront before touching git
	if isLocal, err := connection.IsLocalRepo(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	} else if !isLocal {
		cmd.HandleRepoError(conn.ErrNotAGitRepository)
		return
	}

	// Execute the command
	if err := fn(ctx, connection); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func runMainLogic(ctx context.Context, connection *conn.Connection, state StateFlag, dryRun bool, debug bool) error {
	sp := spinner.New(spinner.CharSets[14], 40*time.Millisecond)
	defer sp.Stop()

	if dryRun {
		fmt.Fprintf(color.Output, "%s\n", bold("== DRY RUN =="))
	}

	fetchingMsg := " Fetching pull requests..."
	sp.Suffix = fetchingMsg
	if !debug {
		sp.Start()
	}

	remote, err := cmd.GetRemote(ctx, connection)
	if err != nil {
		return err
	}

	branches, fetchingErr := cmd.GetBranches(ctx, remote, connection, state.toModel(), dryRun)

	sp.Stop()

	if fetchingErr == nil {
		fmt.Fprintf(color.Output, "%s%s\n", green("✔"), fetchingMsg)
	} else {
		fmt.Fprintf(color.Output, "%s%s\n", red("✕"), fetchingMsg)
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
		connection.PruneRemoteBranches(ctx, remote.Name)

		sp.Stop()

		if deletingErr == nil {
			fmt.Fprintf(color.Output, "%s%s\n", green("✔"), deletingMsg)
		} else {
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

func runMain(state StateFlag, dryRun bool, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	runWithRepoCheck(ctx, debug, func(ctx context.Context, c shared.Connection) error {
		return runMainLogic(ctx, c.(*conn.Connection), state, dryRun, debug)
	})
}

func runLock(branchNames []string, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	runWithRepoCheck(ctx, debug, func(ctx context.Context, conn shared.Connection) error {
		return lock.LockBranches(ctx, branchNames, conn)
	})
}

func runUnlock(branchNames []string, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	runWithRepoCheck(ctx, debug, func(ctx context.Context, conn shared.Connection) error {
		return lock.UnlockBranches(ctx, branchNames, conn)
	})
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
		if branch.IsLocked {
			reason = "locked"
		} else if branch.Worktree != nil && branch.Worktree.IsLocked {
			reason = "worktree locked"
		} else if branch.Worktree != nil && !branch.Worktree.IsMain && branch.Head {
			reason = "worktree here"
		} else if !branch.IsDefault && len(branch.PullRequests) > 0 && branch.HasTrackedChanges {
			reason = "uncommitted changes"
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
