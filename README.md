![logo_readme](https://user-images.githubusercontent.com/5178598/152155497-c06799b7-a95a-44e5-a8a0-a0a9c96ce646.png)

[![CI](https://github.com/seachicken/gh-poi/actions/workflows/ci.yml/badge.svg)](https://github.com/seachicken/gh-poi/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/seachicken/gh-poi/branch/main/graph/badge.svg?token=tcPxPgst2q)](https://codecov.io/gh/seachicken/gh-poi)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6380/badge)](https://bestpractices.coreinfrastructure.org/projects/6380)
[![Stake to support us](https://badge.devprotocol.xyz/0x9ca78E1ca8E49a0e9C8BfB59A8Ed58E1E4440615/descriptive)](https://stakes.social/0x9ca78E1ca8E49a0e9C8BfB59A8Ed58E1E4440615)

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
- `gh poi --dry-run` You can check the branch to be deleted without actually deleting it
