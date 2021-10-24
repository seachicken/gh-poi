# gh poi

A [gh](https://github.com/cli/cli) extension for deleting merged local branches.

This extension checks the state of remote pull requests, so it works even when you "Squash and merge" pull requests.

## Installation

```
gh extension install seachicken/gh-poi
```

## Usage

- `gh poi` Delete the merged local branches
- `gh poi --check` You can check the branch to be deleted without actually deleting it
