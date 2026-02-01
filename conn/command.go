package conn

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cli/safeexec"
	"github.com/seachicken/gh-poi/shared"
)

type (
	Connection struct {
		Debug bool
	}

	DebugMask int
)

const (
	None DebugMask = iota
	Output
)

var (
	hasSchemePattern  = regexp.MustCompile("^[^:]+://")
	scpLikeURLPattern = regexp.MustCompile("^([^@]+@)?([^:]+):(/?.+)$")
)

func (conn *Connection) CheckRepos(ctx context.Context, hostname string, repoNames []string) error {
	for _, name := range repoNames {
		args := []string{
			"api",
			"--hostname", hostname,
			"repos/" + name,
			"--silent",
		}
		if _, err := conn.run(ctx, "gh", args, None); err != nil {
			return err
		}
	}
	return nil
}

func GetRemoteNames(ctx context.Context, conn shared.Connection) ([]shared.Remote, error) {
	output, err := conn.GetRemoteNames(ctx)
	if err != nil {
		return []shared.Remote{}, err
	}
	return parseRemotes(output), nil
}

func (conn *Connection) GetRemoteNames(ctx context.Context) (string, error) {
	args := []string{
		"remote", "-v",
	}
	return conn.run(ctx, "git", args, None)
}

// acceptable url formats:
//
//	ssh://[user@]host.xz[:port]/path/to/repo.git/
//	git://host.xz[:port]/path/to/repo.git/
//	http[s]://host.xz[:port]/path/to/repo.git/
//	ftp[s]://host.xz[:port]/path/to/repo.git/
//
// An alternative scp-like syntax may also be used with the ssh protocol:
//
//	[user@]host.xz:path/to/repo.git/
//
// ref. http://git-scm.com/docs/git-fetch#_git_urls
// the code is heavily inspired by https://github.com/x-motemen/ghq/blob/7163e61e2309a039241ad40b4a25bea35671ea6f/url.go
func parseRemotes(output string) []shared.Remote {
	results := []shared.Remote{}

	for _, remoteConfig := range splitLines(output) {
		splitConfig := strings.Fields(remoteConfig)
		if len(splitConfig) != 3 {
			return []shared.Remote{}
		}

		ref := splitConfig[1]
		if !hasSchemePattern.MatchString(ref) {
			if scpLikeURLPattern.MatchString(ref) {
				matched := scpLikeURLPattern.FindStringSubmatch(ref)
				user := matched[1]
				host := matched[2]
				path := matched[3]
				ref = fmt.Sprintf("ssh://%s%s/%s", user, host, strings.TrimPrefix(path, "/"))
			}
		}
		u, err := url.Parse(ref)
		if err != nil {
			return []shared.Remote{}
		}

		repo := u.Path
		repo = strings.TrimPrefix(repo, "/")
		repo = strings.TrimSuffix(repo, ".git")

		results = append(results, shared.Remote{
			Name:     splitConfig[0],
			Hostname: u.Host,
			RepoName: repo,
		})
	}

	return results
}

func (conn *Connection) GetSshConfig(ctx context.Context, name string) (string, error) {
	args := []string{
		"-T", "-G", name,
	}
	return conn.run(ctx, "ssh", args, Output)
}

func (conn *Connection) GetRepoNames(ctx context.Context, hostname string, repoName string) (string, error) {
	args := []string{
		"repo", "view", hostname + "/" + repoName,
		"--json", "owner,name,parent,defaultBranchRef",
	}
	return conn.run(ctx, "gh", args, None)
}

func (conn *Connection) GetBranchNames(ctx context.Context) (string, error) {
	args := []string{
		"branch", "-v", "--no-abbrev",
		"--format=%(HEAD):%(refname:lstrip=2):%(objectname)",
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) GetMergedBranchNames(ctx context.Context, remoteName string, branchName string) (string, error) {
	args := []string{
		"branch", "--merged", fmt.Sprintf("%s/%s", remoteName, branchName),
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) GetRemoteHeadOid(ctx context.Context, remoteName string, branchName string) (string, error) {
	args := []string{
		"rev-parse", fmt.Sprintf("%s/%s", remoteName, branchName),
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) GetLsRemoteHeadOid(ctx context.Context, url string, branchName string) (string, error) {
	args := []string{
		"ls-remote", url, branchName,
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) GetLog(ctx context.Context, branchName string) (string, error) {
	args := []string{
		"log", "--first-parent", "--max-count=30", "--format=%H", branchName, "--",
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) GetAssociatedRefNames(ctx context.Context, oid string) (string, error) {
	args := []string{
		"branch", "--all", "--format=%(refname)",
		"--contains", oid,
	}
	return conn.run(ctx, "git", args, None)
}

// limitations:
// - https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests#search-within-a-users-or-organizations-repositories
// - https://docs.github.com/en/graphql/overview/resource-limitations
func (conn *Connection) GetPullRequests(
	ctx context.Context,
	hostname string, orgs string, repos string, queryHashes string) (string, error) {
	args := []string{
		"api", "graphql",
		"--hostname", hostname,
		"-f", fmt.Sprintf(`query=query {
  search(type: ISSUE, query: "is:pr %s %s %s", last: 100) {
    issueCount
    edges {
      node {
        ... on PullRequest {
          number
          url
          state
          isDraft
          headRefName
          commits(last: 100) {
            nodes {
              commit {
                oid
              }
            }
          }
          author { login }
        }
      }
    }
  }
}`,
			orgs, repos, queryHashes,
		),
	}
	return conn.run(ctx, "gh", args, None)
}

func (conn *Connection) GetUncommittedChanges(ctx context.Context) (string, error) {
	args := []string{
		"status", "--short",
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) GetConfig(ctx context.Context, key string) (string, error) {
	args := []string{
		"config", "--get", key,
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) AddConfig(ctx context.Context, key string, value string) (string, error) {
	args := []string{
		"config", "--add", key, value,
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) RemoveConfig(ctx context.Context, key string) (string, error) {
	args := []string{
		"config", "--unset", key,
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) CheckoutBranch(ctx context.Context, branchName string) (string, error) {
	args := []string{
		"checkout", "--quiet", branchName,
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) DeleteBranches(ctx context.Context, branchNames []string) (string, error) {
	args := append([]string{
		"branch", "-D"},
		branchNames...,
	)
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) PruneRemoteBranches(ctx context.Context, remoteName string) (string, error) {
	args := []string{
		"remote", "prune", remoteName,
	}
	return conn.run(ctx, "git", args, None)
}

func GetWorktrees(ctx context.Context, conn shared.Connection) ([]shared.Worktree, error) {
	output, err := conn.GetWorktrees(ctx)
	if err != nil {
		return []shared.Worktree{}, err
	}
	return parseWorktrees(output), nil
}

func (conn *Connection) GetWorktrees(ctx context.Context) (string, error) {
	args := []string{
		"worktree", "list", "--porcelain",
	}
	return conn.run(ctx, "git", args, None)
}

func parseWorktrees(output string) []shared.Worktree {
	results := []shared.Worktree{}
	var current *shared.Worktree
	isFirst := true

	for _, line := range splitLines(output) {
		if path, ok := strings.CutPrefix(line, "worktree "); ok {
			if current != nil {
				results = append(results, *current)
			}
			current = &shared.Worktree{
				Path:     path,
				IsMain:   isFirst,
				IsLocked: false,
			}
			isFirst = false
		} else if branch, ok := strings.CutPrefix(line, "branch refs/heads/"); ok {
			if current != nil {
				current.Branch = branch
			}
		} else if line == "locked" {
			current.IsLocked = true
		}
	}

	if current != nil {
		results = append(results, *current)
	}

	return results
}

func (conn *Connection) RemoveWorktree(ctx context.Context, path string) (string, error) {
	args := []string{
		"worktree", "remove", path,
	}
	return conn.run(ctx, "git", args, None)
}

func (conn *Connection) run(ctx context.Context, name string, args []string, mask DebugMask) (string, error) {
	cmdPath, err := safeexec.LookPath(name)
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, cmdPath, args...)
	cmd.Stdout = &stdout
	if name == "gh" {
		cmd.Env = append(os.Environ(), "CLICOLOR_FORCE=0")
	}

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)
	if err != nil {
		err = fmt.Errorf("failed to run external command: %s, args: %v\n %w", name, args, err)
		return "", err
	}

	if conn.Debug {
		switch mask {
		case None:
			log.Printf("[%v] run %s %v -> %q\n", duration, name, args, stdout.String())
		case Output:
			log.Printf("[%v] run %s %v -> *****\n", duration, name, args)
		}
	}

	return stdout.String(), err
}

func splitLines(text string) []string {
	return strings.FieldsFunc(strings.ReplaceAll(text, "\r\n", "\n"),
		func(c rune) bool { return c == '\n' })
}
