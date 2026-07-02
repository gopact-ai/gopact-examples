# Repository Governance

<!-- gopact:doc-language: en -->

Chinese documentation: [repository-governance_zh.md](repository-governance_zh.md)

After `gopact-examples` is public, `main` is PR-only. This repository is a user entry point, so each change should carry CI evidence, review state, auto-merge state, and secret-scanning evidence.

## Pull Request Flow

All example, workflow, and documentation changes land through pull requests. Direct pushes to `main` are reserved for repository recovery.

Required checks include formatting, module tidiness, tests, race tests, static analysis, coverage, vulnerability scanning, and the public-readiness script.

## Admin Auto-Merge

Admin-authored PRs may be squash-merged automatically after required checks pass. The automation still uses a pull request so branch protection, CI logs, and merge history remain visible.

## Public Release Checks

Before a repository is made public, maintainers run the public-readiness script and inspect commit messages for provider keys, endpoint IDs, model IDs, tokens, and private-only notes.

## Author Policy

The `author-policy` job implements conditional review rules:

- Admin-authored PRs can pass without a separate approval after CI succeeds.
- Non-admin-authored PRs require at least one admin approval on the latest commit.
- Do not configure a global required review count, because that would block single-maintainer admin PRs.

Require status checks to pass before merge, including `author-policy`, so both one-person maintenance and external contribution review use the same branch protection model.
