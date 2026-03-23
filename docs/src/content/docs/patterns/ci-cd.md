---
title: CI/CD Integration
sidebar:
  order: 2
---

gh-infra is designed to run in CI. The two most common patterns are **auto-apply on merge** and **scheduled drift detection**. Both work with either [central](../central/) or [self-managed](../self-managed/) setups.

## Auto-Apply on Merge

Automatically apply infrastructure changes when YAML files are merged to `main`. This ensures the live GitHub state always matches the declared state.

```yaml
# .github/workflows/infra-apply.yaml
on:
  push:
    branches: [main]
    paths:
      - "repos/**"
      - "files/**"
jobs:
  apply:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: gh infra apply ./repos/ --auto-approve
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

The `--auto-approve` flag skips the interactive confirmation prompt, which is required in non-interactive environments like CI.

## Authentication

The token you use depends on which pattern you follow.

### Self-Managed (single repo)

The default `GITHUB_TOKEN` provided by GitHub Actions is sufficient. It has permissions to modify the repository where the workflow runs — no extra setup needed.

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Central Management (multiple repos)

The default `GITHUB_TOKEN` **cannot access other repositories**. You need a token with cross-repo permissions:

- **Fine-grained personal access token** — Select the target repositories and grant `Administration: Read and write`, `Contents: Read and write`, and any other permissions you need (e.g., `Secrets: Read and write` for managing secrets).
- **GitHub App installation token** — Recommended for organizations. Create a GitHub App with the required permissions, install it on the target repos, and generate an installation token in your workflow.

Store the token as a repository secret (e.g., `GH_INFRA_TOKEN`) and pass it to the workflow:

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GH_INFRA_TOKEN }}
```

:::caution
Never use a classic personal access token with broad scopes. Use a fine-grained token or GitHub App with the minimum permissions required.
:::

## Drift Detection

Run `plan` on a schedule to detect when the live GitHub state has drifted from your YAML — for example, when someone changes a setting through the GitHub UI.

```yaml
# .github/workflows/infra-drift.yaml
on:
  schedule:
    - cron: "0 9 * * 1"   # Every Monday at 9am UTC
jobs:
  drift:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: gh infra plan ./repos/ --ci
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

The `--ci` flag makes `plan` exit with code 1 if any changes are detected. This turns the workflow into a pass/fail check that you can wire up to notifications (Slack, email, etc.) or use as a required status check.

## Combining Both

For a complete setup, use both workflows together:

- **Auto-apply** keeps the live state in sync after intentional changes
- **Drift detection** catches unintentional changes made outside of gh-infra

This gives you confidence that the YAML files are the true source of truth — not just a snapshot that may have drifted.
