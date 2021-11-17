package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Yash-Handa/spinner"
	"github.com/fatih/color"
)

var (
	white     = color.New(color.FgWhite).SprintFunc()
	whiteBold = color.New(color.FgWhite, color.Bold).SprintFunc()
	hiBlack   = color.New(color.FgHiBlack).SprintFunc()
	green     = color.New(color.FgGreen).SprintFunc()
	red       = color.New(color.FgRed).SprintFunc()
)

func main() {
	var check bool
	flag.BoolVar(&check, "check", false, "Show branches to delete")
	flag.Parse()

	runMain(check)
}

func runMain(check bool) {
	if check {
		fmt.Fprintf(color.Output, "%s\n", whiteBold("== DRY RUN =="))
	}

	sp, _ := spinner.New(1020, 30*time.Millisecond, "", "")
	sp.Start()

	conn := &ConnectionImpl{}

	fetchingMsg := " Fetching pull requests..."
	var fetchingErr error
	sp.SetPostText(fetchingMsg)
	branches, fetchingErr := GetBranches(conn)

	deletingMsg := " Deleting branches..."
	var deletingErr error
	if !check && fetchingErr == nil {
		sp.SetPostText(deletingMsg)
		branches, deletingErr = DeleteBranches(branches, conn)
	}

	sp.Stop()

	if fetchingErr == nil {
		fmt.Fprintf(color.Output, "%s%s\n", green("✔"), fetchingMsg)
	} else {
		fmt.Fprintf(color.Output, "%s%s\n", red("✕"), fetchingMsg)
		fmt.Fprintln(os.Stderr, fetchingErr)
		return
	}

	if check {
		fmt.Fprintf(color.Output, "%s%s\n", hiBlack("-"), deletingMsg)
	} else {
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
	if check {
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
			issueNoColor := getIssueNoColor(pr.State)
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

func getIssueNoColor(state PullRequestState) color.Attribute {
	switch state {
	case Open:
		return color.FgGreen
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
