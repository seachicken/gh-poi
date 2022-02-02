![logo_readme](https://user-images.githubusercontent.com/5178598/152155497-c06799b7-a95a-44e5-a8a0-a0a9c96ce646.png)

[![CI](https://github.com/seachicken/gh-poi/actions/workflows/ci.yml/badge.svg)](https://github.com/seachicken/gh-poi/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/seachicken/gh-poi/branch/main/graph/badge.svg?token=tcPxPgst2q)](https://codecov.io/gh/seachicken/gh-poi)
[![Stake to support us](https://badge.devprotocol.xyz/0x9ca78E1ca8E49a0e9C8BfB59A8Ed58E1E4440615/descriptive)](https://stakes.social/0x9ca78E1ca8E49a0e9C8BfB59A8Ed58E1E4440615)

A [gh](https://github.com/cli/cli) extension for deleting merged local branches.

This extension checks the state of remote pull requests, so it works even when you "Squash and merge" pull requests.

![demo](https://user-images.githubusercontent.com/5178598/140624593-bf38ded3-388b-4a4b-a5c0-4053f8de51ad.gif)

## Installation

```
gh extension install seachicken/gh-poi
```

## Usage

- `gh poi` Delete the merged local branches
- `gh poi --check` You can check the branch to be deleted without actually deleting it

## ✨ Local branch to be deleted

- Branches merged in the origin repository
- Branches merged in the upstream repository
