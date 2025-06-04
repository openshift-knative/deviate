# Deviate

[![Mage](https://github.com/openshift-knative/deviate/actions/workflows/mage.yaml/badge.svg)](https://github.com/openshift-knative/deviate/actions/workflows/mage.yaml)
[![Lints](https://github.com/openshift-knative/deviate/actions/workflows/lint.yaml/badge.svg)](https://github.com/openshift-knative/deviate/actions/workflows/lint.yaml)
[![Go Report Card](https://goreportcard.com/badge/openshift-knative/deviate)](https://goreportcard.com/report/openshift-knative/deviate)
[![Releases](https://img.shields.io/github/release-pre/openshift-knative/deviate.svg?sort=semver)](https://github.com/openshift-knative/deviate/releases)
[![LICENSE](https://img.shields.io/github/license/openshift-knative/deviate.svg)](https://github.com/openshift-knative/deviate/blob/main/LICENSE)

A tool used to handle forks of upstream projects with workflow used in 
OpenShift Serverless.

Handles forks of upstream projects with a customizable workflow. `deviate` automates the synchronization of changes from an upstream repository to your fork, including managing release branches, applying patches, and creating well-defined pull requests.

## Features

*   **Upstream Synchronization**: Keeps your fork aligned with an upstream source.
*   **Configurable Behavior**: Uses a YAML configuration file (`.deviate.yaml`) to define upstream/downstream repositories, branch names, PR labels, commit/PR messages, and more.
*   **Automated PR Creation**: Generates pull requests for synchronization changes using the `gh` CLI.
*   **CI Integration**: Creates a special "CI trigger" commit to ensure your CI workflows run on the proposed changes.
*   **Release Management**: Supports patterns for release branch naming and synchronization.
*   **Tag Syncing**: Can synchronize Git tags from the upstream.

## Configuration (`.deviate.yaml`)

`deviate` requires a configuration file named `.deviate.yaml` in the root of your repository. This file defines how `deviate` should operate.

Here's an example structure with comments explaining key fields:

```yaml
# Upstream repository URL (e.g., git@github.com:upstream/project.git)
upstream: "UPSTREAM_REPO_URL"
# Downstream repository URL (your fork, e.g., git@github.com:your-org/project-fork.git)
downstream: "DOWNSTREAM_REPO_URL"

# Set to true to simulate changes without pushing or creating PRs
dryRun: false

# Glob pattern for GitHub workflow files to remove (e.g., from upstream, if they are not relevant to the fork)
githubWorkflowsRemovalGlob: ".github/workflows/upstream-ci-*.yaml"

# Labels to apply to Pull Requests created by deviate
# These can be used by tools like Mergify for auto-merging.
syncLabels:
  - "bot/sync"
  - "apply-patches"

# Configuration for Dockerfile generation (if applicable)
# See github.com/openshift-knative/hack/pkg/dockerfilegen for details
dockerfileGen:
  # enabled: true
  # ... other dockerfilegen params
  # Disabled by default, enable and configure if needed.
  enabled: false

# Configuration for re-syncing a certain number of past releases
resyncReleases:
  enabled: false # Set to true to enable resyncing past releases
  numberOf: 3    # Number of past releases to resync if enabled

branches:
  # Name of the main/default branch in your fork
  main: "main"
  # Prefix or name for the branch that tracks the next upstream release
  releaseNext: "release-" # This often includes a version, e.g., "release-1.23" which deviate will determine
  # Name of the branch used for triggering CI builds before creating the main sync PR
  # This will be combined with the release branch name, e.g., "sync-ci-release-1.23"
  synchCi: "sync-ci-"
  releaseTemplates:
    # Go template for naming upstream release branches. {{ .Version }} is available.
    upstream: "release-{{ .Version }}"
    # Go template for naming downstream release branches in your fork. {{ .Version }} is available.
    downstream: "release-{{ .Version }}"
  searches:
    # Regex to find upstream release branches
    upstreamReleases: '^release-(?P<Version>\d+\.\d+)$'
    # Regex to find downstream release branches (in your fork)
    downstreamReleases: '^release-(?P<Version>\d+\.\d+)$'

tags:
  # Whether to synchronize tags
  synchronize: true
  # RefSpec for tag synchronization (e.g., "v*")
  refSpec: "v*"

messages:
  # Commit message (and PR title) for the CI trigger PR.
  # {{ .ReleaseBranch }} and {{ .MainBranch }} are available placeholders.
  triggerCi: "chore(sync): Trigger CI for {{ .ReleaseBranch }} into {{ .MainBranch }}"
  # PR body for the CI trigger PR.
  # {{ .ReleaseBranch }} and {{ .MainBranch }} are available placeholders.
  triggerCiBody: "Automated PR to trigger CI for syncing `{{ .ReleaseBranch }}` into `{{ .MainBranch }}`."
  # Commit message for applying fork-specific files/patches.
  applyForkFiles: "chore: Apply fork-specific files and patches"
  # Commit message when images are generated (if dockerfileGen is used).
  imagesGenerated: "chore: Generate images"

## How `deviate sync` Works

The core command is `deviate sync`. When executed (typically in a GitHub Action):

1.  **Loads Configuration**: Reads `.deviate.yaml`.
2.  **Fetches Remotes**: Updates local refs for upstream and downstream repositories.
3.  **Determines Releases**: Identifies relevant release branches based on `branches.searches` and `branches.releaseTemplates`.
4.  **Mirrors Releases**: For each relevant release, it mirrors the upstream release branch to a local equivalent in your fork.
5.  **Applies Patches/Fork Files**: (Implied) Applies any patches or fork-specific modifications. This often involves a set of patch files or scripts managed within your fork.
6.  **Tag Syncing**: If `tags.synchronize` is true, it syncs tags matching `tags.refSpec`.
7.  **Prepares "Release Next" Sync**:
    *   Checks out the downstream `releaseNext` branch (e.g., `release-1.23`).
    *   Creates a temporary branch from it (e.g., `sync-ci-release-1.23`).
    *   Adds a small, timestamped file (`ci`) to this temporary branch and commits it. This ensures there's a change to trigger CI workflows. The commit message uses the `messages.triggerCi` template.
    *   Pushes this temporary CI trigger branch.
8.  **Creates Sync Pull Request**:
    *   Uses `gh pr create` to open a Pull Request.
    *   **Title**: From `messages.triggerCi` template.
    *   **Body**: From `messages.triggerCiBody` template.
    *   **Base Branch**: The target release branch (e.g., `release-1.23`).
    *   **Head Branch**: The temporary CI trigger branch (e.g., `sync-ci-release-1.23`).
    *   **Labels**: Applies labels defined in `syncLabels` from your config.

## Quickstart: GitHub Actions Workflow

There isn't a pre-configured GitHub Actions workflow for `deviate sync` in this repository, as its execution is specific to the fork using it. Here's an example workflow you can add to your forked repository (e.g., in `.github/workflows/sync-upstream.yaml`):

```yaml
name: Sync Upstream

on:
  schedule:
    # Run at 3 AM UTC every day
    - cron: '0 3 * * *'
  workflow_dispatch: {} # Allows manual triggering

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Fork
        uses: actions/checkout@v4
        with:
          # Ensure you have a token with write access to create PRs
          # This is usually the default GITHUB_TOKEN for actions in the same repo
          token: ${{ secrets.GITHUB_TOKEN }}
          # Fetch all history for all tags and branches
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21' # Or your project's Go version

      - name: Install gh CLI
        uses: dev-hanz-ops/install-gh-cli-action@v0.2.1

      - name: Install deviate
        run: go install github.com/openshift-knative/deviate/cmd/deviate@latest # Or pin to a specific version

      - name: Configure Git User
        run: |
          git config --global user.email "your-bot-email@example.com"
          git config --global user.name "Your Bot Name"

      - name: Run deviate sync
        run: deviate sync
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          # If your .deviate.yaml contains sensitive information or needs to be dynamically generated,
          # you might consider alternative ways to provide it (e.g., create it in a previous step).
          # For most use cases, committing .deviate.yaml to your fork is standard.
```

**Before using this workflow:**

1.  **Create `.deviate.yaml`**: Ensure a valid `.deviate.yaml` configuration file exists in the root of your forked repository.
2.  **Git User**: Update the Git user email and name in the workflow to attribute commits correctly.
3.  **Go Version**: Adjust the Go version if necessary.
4.  **Permissions**: The `GITHUB_TOKEN` usually has sufficient permissions to create PRs within the same repository. If `deviate` needs to push to protected branches directly (not typical for PR creation flow), you might need a PAT with more permissions.

## Auto-merging Sync PRs with Mergify

`deviate` applies labels (from `syncLabels` in `.deviate.yaml`) to the pull requests it creates. You can use these labels to configure Mergify for auto-merging.

Since there is no `.mergify.yml` file in this repository (as Mergify configuration is specific to the repository using it), you'll need to configure Mergify through its dashboard (app.mergify.com):

1.  **Ensure Mergify is installed** on your forked repository.
2.  **Define `syncLabels` in `.deviate.yaml`**: Choose one or more labels that signify a PR is a sync PR ready for potential auto-merge (e.g., `bot/sync`, `automerge-sync`).
3.  **Create Mergify Rules**: In the Mergify dashboard for your repository, create rules that match these labels and define the conditions for merging.

   Example Mergify rule concept (syntax for illustration):

   ```yaml
   # This is conceptual and needs to be configured in the Mergify UI or a .mergify.yml in your fork
   pull_request_rules:
     - name: Auto-merge deviate sync PRs
       conditions:
         - "label=bot/sync"        # Matches one of your syncLabels
         - "status-success=CI_Check_1" # Ensure your CI checks pass
         - "status-success=CI_Check_2"
         # Add other conditions like no conflicts, approved reviews (if needed), etc.
       actions:
         merge:
           method: squash # or merge, rebase
   ```

By configuring `deviate` to apply specific labels and then instructing Mergify to act on those labels, you can achieve a fully automated sync and merge process.
