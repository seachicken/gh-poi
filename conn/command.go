package conn

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cli/safeexec"
)

type Connection struct{}

func (conn *Connection) CheckRepos(ctx context.Context, hostname string, repoNames []string) error {
	for _, name := range repoNames {
		args := []string{
			"api",
			"--hostname", hostname,
			"repos/" + name,
			"--silent",
		}
		if _, err := run(ctx, "gh", args); err != nil {
			return err
		}
	}
	return nil
}

func (conn *Connection) GetRemoteNames(ctx context.Context) (string, error) {
	args := []string{
		"remote", "-v",
	}
	return run(ctx, "git", args)
}

func (conn *Connection) GetSshConfig(ctx context.Context, name string) (string, error) {
	args := []string{
		"-T", "-G", name,
	}
	return run(ctx, "ssh", args)
}

func (conn *Connection) GetRepoNames(ctx context.Context, hostname string, repoName string) (string, error) {
	args := []string{
		"repo", "view", hostname + "/" + repoName,
		"--json", "owner",
		"--json", "name",
		"--json", "parent",
		"--json", "defaultBranchRef",
	}
	return run(ctx, "gh", args)
}

func (conn *Connection) GetBranchNames(ctx context.Context) (string, error) {
	args := []string{
		"branch", "-v", "--no-abbrev",
		"--format=%(HEAD):%(refname:lstrip=2):%(objectname)",
	}
	return run(ctx, "git", args)
}

func (conn *Connection) GetMergedBranchNames(ctx context.Context, remoteName string, branchName string) (string, error) {
	args := []string{
		"branch", "--merged", fmt.Sprintf("%s/%s", remoteName, branchName),
	}
	return run(ctx, "git", args)
}

func (conn *Connection) GetLog(ctx context.Context, branchName string) (string, error) {
	args := []string{
		"log", "--first-parent", "--max-count=30", "--format=%H", branchName, "--",
	}
	return run(ctx, "git", args)
}

func (conn *Connection) GetAssociatedRefNames(ctx context.Context, oid string) (string, error) {
	args := []string{
		"branch", "--all", "--format=%(refname)",
		"--contains", oid,
	}
	return run(ctx, "git", args)
}

func (conn *Connection) GetPullRequests(
	ctx context.Context,
	hostname string, repoNames []string, queryHashes string) (string, error) {
	args := []string{
		"api", "graphql",
		"--hostname", hostname,
		"-f", fmt.Sprintf(`query=query {
  search(type: ISSUE, query: "is:pr %s %s", last: 100) {
    issueCount
    edges {
      node {
        ... on PullRequest {
          number
          url
          state
          isDraft
          headRefName
          commits(last: 10) {
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
			getQueryRepos(repoNames),
			queryHashes,
		),
	}
	return run(ctx, "gh", args)
}

func (conn *Connection) GetUncommittedChanges(ctx context.Context) (string, error) {
	args := []string{
		"status", "--short",
	}
	return run(ctx, "git", args)
}

func (conn *Connection) GetConfig(ctx context.Context, key string) (string, error) {
	args := []string{
		"config", "--get", key,
	}
	return run(ctx, "git", args)
}

func (conn *Connection) CheckoutBranch(ctx context.Context, branchName string) (string, error) {
	args := []string{
		"checkout", "--quiet", branchName,
	}
	return run(ctx, "git", args)
}

func (conn *Connection) DeleteBranches(ctx context.Context, branchNames []string) (string, error) {
	args := append([]string{
		"branch", "-D"},
		branchNames...,
	)
	return run(ctx, "git", args)
}

func (conn *Connection) PruneRemoteBranches(ctx context.Context, remoteName string) (string, error) {
	args := []string{
		"remote", "prune", remoteName,
	}
	return run(ctx, "git", args)
}

func getQueryRepos(repoNames []string) string {
	var repos strings.Builder
	for _, name := range repoNames {
		repos.WriteString(fmt.Sprintf("repo:%s ", name))
	}
	return repos.String()
}

func run(ctx context.Context, name string, args []string) (string, error) {
	cmdPath, err := safeexec.LookPath(name)
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, cmdPath, args...)
	cmd.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to run external command: %s, args: %v\n%w", name, args, err)
	}

	return stdout.String(), err
}
