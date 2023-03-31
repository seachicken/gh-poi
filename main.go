package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/seachicken/gh-poi/cmd"
	"github.com/seachicken/gh-poi/cmd/protect"
	"github.com/seachicken/gh-poi/conn"
	"github.com/seachicken/gh-poi/shared"
)

var (
	white     = color.New(color.FgWhite).SprintFunc()
	whiteBold = color.New(color.FgWhite, color.Bold).SprintFunc()
	hiBlack   = color.New(color.FgHiBlack).SprintFunc()
	green     = color.New(color.FgGreen).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
)

func main() {
	var dryRun bool
	var debug bool
	flag.BoolVar(&dryRun, "dry-run", false, "Show branches to delete")
	flag.BoolVar(&debug, "debug", false, "Enable debug logs")
	flag.Usage = func() {
		fmt.Fprintf(color.Output, "%s\n\n", white("Delete the merged local branches."))
		fmt.Fprintf(color.Output, "%s\n", whiteBold("USAGE"))
		fmt.Fprintf(color.Output, "  %s\n\n", white("gh poi <command> [flags]"))
		fmt.Fprintf(color.Output, "%s", whiteBold("COMMANDS"))
		fmt.Fprintf(color.Output, "%s\n", white(`
  protect:   Protect local branches from deletion
  unprotect: Unprotect local branches
  `))
		fmt.Fprintf(color.Output, "%s\n", whiteBold("FLAGS"))
		flag.PrintDefaults()
		fmt.Println()
	}
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		runMain(dryRun, debug)
	} else {
		subcmd, args := args[0], args[1:]
		switch subcmd {
		case "protect":
			protectCmd := flag.NewFlagSet("protect", flag.ExitOnError)
			protectCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", white("Protect local branches from deletion."))
				fmt.Fprintf(color.Output, "%s\n", whiteBold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", white("gh poi protect <branchname>..."))
			}
			protectCmd.Parse(args)

			runProtect(args, debug)
		case "unprotect":
			unprotectCmd := flag.NewFlagSet("unprotect", flag.ExitOnError)
			unprotectCmd.Usage = func() {
				fmt.Fprintf(color.Output, "%s\n\n", white("Unprotect local branches."))
				fmt.Fprintf(color.Output, "%s\n", whiteBold("USAGE"))
				fmt.Fprintf(color.Output, "  %s\n\n", white("gh poi unprotect <branchname>..."))
			}
			unprotectCmd.Parse(args)

			runUnprotect(args, debug)
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q for poi\n", subcmd)
		}
	}
}

func runMain(dryRun bool, debug bool) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if dryRun {
		fmt.Fprintf(color.Output, "%s\n", whiteBold("== DRY RUN =="))
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

	branches, fetchingErr := cmd.GetBranches(ctx, remote, connection, dryRun)

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

	fmt.Fprintf(color.Output, "%s\n", whiteBold("Deleted branches"))
	printBranches(getBranches(branches, deletedStates))
	fmt.Println()

	fmt.Fprintf(color.Output, "%s\n", whiteBold("Branches not deleted"))
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
			fmt.Fprintf(color.Output, "  %s", white(branch.Name))
		}
		reason := ""
		if branch.IsProtected {
			reason = "protected"
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
				white(pr.Url),
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
		if contains(branch.State, states) {
			results = append(results, branch)
		}
	}
	return results
}

func contains(state shared.BranchState, states []shared.BranchState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}
