package shared

import "strings"

type Worktree struct {
	Path   string
	Branch string
	IsMain bool
}

// ParseWorktrees parses the output of `git worktree list --porcelain`
func ParseWorktrees(output string) []Worktree {
	worktrees := []Worktree{}
	lines := strings.FieldsFunc(strings.ReplaceAll(output, "\r\n", "\n"),
		func(c rune) bool { return c == '\n' })

	var current *Worktree
	isFirst := true

	for _, line := range lines {
		if after, ok := strings.CutPrefix(line, "worktree "); ok {
			if current != nil {
				worktrees = append(worktrees, *current)
			}
			current = &Worktree{
				Path:   after,
				IsMain: isFirst,
			}
			isFirst = false
		} else if after, ok := strings.CutPrefix(line, "branch refs/heads/"); ok {
			if current != nil {
				current.Branch = after
			}
		}
	}

	if current != nil {
		worktrees = append(worktrees, *current)
	}

	return worktrees
}
