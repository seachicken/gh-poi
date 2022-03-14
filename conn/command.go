package conn

import (
	"bytes"
	"fmt"
	"strings"

	exec "golang.org/x/sys/execabs"
)

type Connection struct{}

func (conn *Connection) CheckRepos(hostname string, repoNames []string) error {
	for _, name := range repoNames {
		args := []string{
			"api",
			"--hostname", hostname,
			"repos/" + name,
			"--silent",
		}
		if _, err := run("gh", args); err != nil {
			return err
		}
	}
	return nil
}

func (conn *Connection) GetRemoteNames() (string, error) {
	args := []string{
		"remote", "-v",
	}
	return run("git", args)
}

func (conn *Connection) GetRepoNames(hostname string, repoName string) (string, error) {
	args := []string{
		"repo", "view", hostname + "/" + repoName,
		"--json", "owner",
		"--json", "name",
		"--json", "parent",
		"--json", "defaultBranchRef",
	}
	return run("gh", args)
}

func (conn *Connection) GetBranchNames() (string, error) {
	args := []string{
		"branch", "-v", "--no-abbrev",
		"--format=%(HEAD),%(refname:lstrip=2),%(objectname)",
	}
	return run("git", args)
}

func (conn *Connection) GetLog(branchName string) (string, error) {
	args := []string{
		"log", "--first-parent", "--max-count=30", "--format=%H", branchName,
	}
	return run("git", args)
}

func (conn *Connection) GetAssociatedRefNames(oid string) (string, error) {
	args := []string{
		"branch", "--all", "--format=%(refname)",
		"--contains", oid,
	}
	return run("git", args)
}

func (conn *Connection) GetPullRequests(
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
	return run("gh", args)
}

func (conn *Connection) GetUncommittedChanges() (string, error) {
	args := append([]string{
		"status", "--short"})
	return run("git", args)
}

func (conn *Connection) GetConfig(key string) (string, error) {
	args := []string{
		"config", "--get", key,
	}
	return run("git", args)
}

func (conn *Connection) CheckoutBranch(branchName string) (string, error) {
	args := append([]string{
		"checkout", "--quiet", branchName})
	return run("git", args)
}

func (conn *Connection) DeleteBranches(branchNames []string) (string, error) {
	args := append([]string{
		"branch", "-D"},
		branchNames...)
	return run("git", args)
}

func getQueryRepos(repoNames []string) string {
	var repos strings.Builder
	for _, name := range repoNames {
		repos.WriteString(fmt.Sprintf("repo:%s ", name))
	}
	return repos.String()
}

func run(name string, args []string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to run external command: %s", name)
	}
	cmd.Wait()

	if stderr.Len() > 0 {
		return "", fmt.Errorf("failed to run external command: %s\n%s", name, stderr.String())
	}

	return stdout.String(), nil
}
