package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/seachicken/gh-poi/cmd"
	"github.com/seachicken/gh-poi/cmd/protect"
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
  protect:   Protect local branches from deletion
  unprotect: Unprotect local branches
  `)
		fmt.Fprintf(color.Output, "%s\n", bold("FLAGS"))
		maxLen := 0
		flag.VisitAll(func(f *flag.Flag) {
			if len(f.Name) > maxLen {
				maxLen = len(f.Name)
			}
		})
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(color.Output, "  --%-*s %s\n", maxLen, f.Name, f.Usage)
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
		case "protect":
			protectCmd := flag.NewFlagSet("protect", flag.ExitOnError)
			protectCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", "Protect local branches from deletion.")
				fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", "gh poi protect <branchname>...")
			}
			protectCmd.Parse(args)

			runProtect(args, debug)
		case "unprotect":
			unprotectCmd := flag.NewFlagSet("unprotect", flag.ExitOnError)
			unprotectCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", "Unprotect local branches.")
				fmt.Fprintf(color.Output, "%s\n", bold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", "gh poi unprotect <branchname>...")
			}
			unprotectCmd.Parse(args)

			runUnprotect(args, debug)
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q for poi\n", subcmd)
		}
	}
}

func runMain(state StateFlag, dryRun bool, debug bool) {
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

	remote, err := cmd.GetRemote(ctx, connection)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	branches, fetchingErr := cmd.GetBranches(ctx, remote, connection, state.toModel(), dryRun)

	sp.Stop()

	if fetchingErr == nil {
		fmt.Fprintf(color.Output, "%s%s\n", green("✔"), fetchingMsg)
	} else {
		fmt.Fprintf(color.Output, "%s%s\n", red("✕"), fetchingMsg)
		fmt.Fprintln(os.Stderr, fetchingErr)
		return
	}

	deletingMsg := " Deleting branches..."
	var deletingErr error

	if dryRun {
		fmt.Fprintf(color.Output, "%s%s\n", hiBlack("-"), deletingMsg)
	} else {
		sp.Suffix = deletingMsg
		if !debug {
			sp.Restart()
		}

		branches, deletingErr = cmd.DeleteBranches(ctx, branches, connection)
		connection.PruneRemoteBranches(ctx, remote.Name)

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
}

func runProtect(branchNames []string, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &conn.Connection{Debug: debug}

	err := protect.ProtectBranches(ctx, branchNames, connection)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func runUnprotect(branchNames []string, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	connection := &conn.Connection{Debug: debug}

	err := protect.UnprotectBranches(ctx, branchNames, connection)
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
		reason := ""
		if branch.IsProtected {
			reason = "protected"
		}
		if !branch.IsDefault && len(branch.PullRequests) > 0 && branch.HasTrackedChanges {
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
