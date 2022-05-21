package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/seachicken/gh-poi/conn"
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
	flag.BoolVar(&dryRun, "dry-run", false, "Show branches to delete")
	flag.BoolVar(&dryRun, "check", false, "[Deprecated] Show branches to delete")
	flag.Parse()

	runMain(dryRun)
}

func runMain(dryRun bool) {
	if dryRun {
		fmt.Fprintf(color.Output, "%s\n", whiteBold("== DRY RUN =="))
	}

	conn := &conn.Connection{}
	sp := spinner.New(spinner.CharSets[14], 40*time.Millisecond)

	fetchingMsg := " Fetching pull requests..."
	sp.Suffix = fetchingMsg
	sp.Start()
	var fetchingErr error

	branches, fetchingErr := GetBranches(conn, dryRun)

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
		sp.Restart()

		branches, deletingErr = DeleteBranches(branches, conn)

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

	var deletedStates []BranchState
	var notDeletedStates []BranchState
	if dryRun {
		deletedStates = []BranchState{Deletable}
		notDeletedStates = []BranchState{NotDeletable}
	} else {
		deletedStates = []BranchState{Deleted}
		notDeletedStates = []BranchState{Deletable, NotDeletable}
	}

	fmt.Fprintf(color.Output, "%s\n", whiteBold("Deleted branches"))
	printBranches(getBranches(branches, deletedStates))
	fmt.Println()

	fmt.Fprintf(color.Output, "%s\n", whiteBold("Not deleted branches"))
	printBranches(getBranches(branches, notDeletedStates))
	fmt.Println()
}

func printBranches(branches []Branch) {
	if len(branches) == 0 {
		fmt.Fprintf(color.Output, "%s\n",
			hiBlack("  There are no branches in the current directory"))
	}

	for _, branch := range branches {
		branchName := fmt.Sprintf("%s", branch.Name)
		if branch.Head {
			fmt.Fprintf(color.Output, "* %s\n", green(branchName))
		} else {
			fmt.Fprintf(color.Output, "  %s\n", white(branchName))
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

func getIssueNoColor(state PullRequestState, isDraft bool) color.Attribute {
	switch state {
	case Open:
		if isDraft {
			return color.FgHiBlack
		} else {
			return color.FgGreen
		}
	case Merged:
		return color.FgMagenta
	case Closed:
		return color.FgRed
	default:
		return color.FgHiBlack
	}
}

func getBranches(branches []Branch, states []BranchState) []Branch {
	results := []Branch{}
	for _, branch := range branches {
		if contains(branch.State, states) {
			results = append(results, branch)
		}
	}
	return results
}

func contains(state BranchState, states []BranchState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}
