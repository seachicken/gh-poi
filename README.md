# gh poi

A [gh](https://github.com/cli/cli) extension for deleting merged local branches.

This extension checks the state of remote pull requests, so it works even when you "Squash and merge" pull requests.

![demo](https://user-images.githubusercontent.com/5178598/140025647-c8895ea4-72ec-4af3-969e-01896e15535c.gif)

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
