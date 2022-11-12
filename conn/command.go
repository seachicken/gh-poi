package conn

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/cli/safeexec"
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

func (conn *Connection) GetRemoteNames(ctx context.Context) (string, error) {
	args := []string{
		"remote", "-v",
	}
	return conn.run(ctx, "git", args, None)
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
		// disable colored outputs (https://github.com/seachicken/gh-poi/issues/79)
		"--jq", ".",
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

func (conn *Connection) GetPullRequests(
	ctx context.Context,
	hostname string, repoNames []string, queryHashes string) (string, error) {
	args := []string{
		"api", "graphql",
		"--hostname", hostname,
		"--jq", ".",
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

func (conn *Connection) run(ctx context.Context, name string, args []string, mask DebugMask) (string, error) {
	cmdPath, err := safeexec.LookPath(name)
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, cmdPath, args...)
	cmd.Stdout = &stdout

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)
	if err != nil {
		err = fmt.Errorf("failed to run external command: %s, args: %v\n%w", name, args, err)
	}

	if conn.Debug {
		switch mask {
		case None:
			log.Printf("[%v] run %s %v -> %s\n", duration, name, args, stdout.String())
		case Output:
			log.Printf("[%v] run %s %v -> *****\n", duration, name, args)
		}
	}

	return stdout.String(), err
}

func getQueryRepos(repoNames []string) string {
	var repos strings.Builder
	for _, name := range repoNames {
		repos.WriteString(fmt.Sprintf("repo:%s ", name))
	}
	return repos.String()
}
