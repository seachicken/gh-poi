![logo_readme](https://user-images.githubusercontent.com/5178598/152155497-c06799b7-a95a-44e5-a8a0-a0a9c96ce646.png)

[![CI](https://github.com/seachicken/gh-poi/actions/workflows/ci.yml/badge.svg)](https://github.com/seachicken/gh-poi/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/seachicken/gh-poi/branch/main/graph/badge.svg?token=tcPxPgst2q)](https://codecov.io/gh/seachicken/gh-poi)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6380/badge)](https://bestpractices.coreinfrastructure.org/projects/6380)

This [gh](https://github.com/cli/cli) extension determines which local branches have been merged and safely deletes them.

![demo](https://user-images.githubusercontent.com/5178598/140624593-bf38ded3-388b-4a4b-a5c0-4053f8de51ad.gif)

## Motivation

Daily development makes it difficult to know which branch is active when there are many unnecessary branches left locally, which causes a small amount of stress. If you squash merge a pull request, there is no history of the merge to the default branch, so you have to force delete the branch to clean it up, and you have to be careful not to accidentally delete the active branch.

We have made it possible to automatically determine which branches have been merged and clean up the local environment without worry.

## Installation

```
gh extension install seachicken/gh-poi
```

## Usage

- `gh poi` Delete the merged local branches
- `gh poi --state (closed|merged)` Specify the PR state to delete (default merged)
- `gh poi --dry-run` Show branches to delete without actually deleting it
- `gh poi --debug` Enable debug logs
- `gh poi protect <branchname>...` Protect local branches from deletion
- `gh poi unprotect <branchname>...` Unprotect local branches

## FAQ

### Why the name "poi"?

"poi" means "feel free to throw it away" in Japanese.  
If you prefer an alias, you can change it with [gh alias set](https://cli.github.com/manual/gh_alias_set). (e.g. `gh alias set clean-branches poi`)
