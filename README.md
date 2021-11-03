# gh poi

A [gh](https://github.com/cli/cli) extension for deleting merged local branches.

This extension checks the state of remote pull requests, so it works even when you "Squash and merge" pull requests.

![demo](https://user-images.githubusercontent.com/5178598/140021435-bcc01ce0-a7d1-488b-a0d9-c46351c57229.gif)

## Installation

```
gh extension install seachicken/gh-poi
```

## Usage

- `gh poi` Delete the merged local branches
- `gh poi --check` You can check the branch to be deleted without actually deleting it

## 🧹 Local branch to be deleted

- 🗑 Branches merged in the origin repository
- 🗑 Branches merged in the upstream repository
