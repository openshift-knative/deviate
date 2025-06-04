# Deviate

[![Mage](https://github.com/openshift-knative/deviate/actions/workflows/mage.yaml/badge.svg)](https://github.com/openshift-knative/deviate/actions/workflows/mage.yaml)
[![Lints](https://github.com/openshift-knative/deviate/actions/workflows/lint.yaml/badge.svg)](https://github.com/openshift-knative/deviate/actions/workflows/lint.yaml)
[![Go Report Card](https://goreportcard.com/badge/openshift-knative/deviate)](https://goreportcard.com/report/openshift-knative/deviate)
[![Releases](https://img.shields.io/github/release-pre/openshift-knative/deviate.svg?sort=semver)](https://github.com/openshift-knative/deviate/releases)
[![LICENSE](https://img.shields.io/github/license/openshift-knative/deviate.svg)](https://github.com/openshift-knative/deviate/blob/main/LICENSE)

`deviate` is a general-purpose tool designed to manage forks of upstream projects and automate the synchronization of changes. It helps maintain your fork by managing release branches, applying fork-specific patches, and creating well-defined pull requests, potentially replacing manual scripts like `update-to-head.sh` and `create-release.sh`.

It promotes an **upstream-first** contribution model, where fork-specific patches represent the minimal delta required, and most changes are ideally contributed back to the upstream project.

While initially developed within the OpenShift Serverless context, its core functionality is not specific to it. An optional feature for Dockerfile generation (`dockerfileGen`) uses a library from `openshift-knative/hack`, but the rest of the tool is broadly applicable.

## Features

*   **Upstream Synchronization**: Keeps your fork aligned with an upstream source.
*   **Configurable Behavior**: Uses a YAML configuration file (`.deviate.yaml`) to define upstream/downstream repositories, branch names, PR labels, commit/PR messages, and more.
*   **Automated PR Creation**: Generates pull requests for synchronization changes using the `gh` CLI.
*   **CI Integration**: Creates a special "CI trigger" commit to ensure your CI workflows run on the proposed changes.
*   **Release Management**:
    *   Supports patterns for release branch naming and synchronization from upstream release branches.
    *   Automates the creation of "resync PRs" to update fork release branches when the corresponding upstream release branch receives updates (e.g., cherry-picked commits).
*   **Tag Syncing**: Can synchronize Git tags from the upstream.
*   **Patch Application**: Manages fork-specific patches, applying them to the development line and at the creation of new release branches.

## Configuration (`.deviate.yaml`)

`deviate` requires a configuration file named `.deviate.yaml` in the root of your repository.

Here's an example structure:

```yaml
# Upstream repository URL (e.g., git@github.com:upstream/project.git)
upstream: "UPSTREAM_REPO_URL"
# Downstream repository URL (your fork, e.g., git@github.com:your-org/project-fork.git)
downstream: "DOWNSTREAM_REPO_URL"

# Set to true to simulate changes without pushing or creating PRs
dryRun: false

# Glob pattern for GitHub workflow files to remove (e.g., from upstream)
githubWorkflowsRemovalGlob: ".github/workflows/upstream-ci-*.yaml"

# Labels to apply to Pull Requests created by deviate
syncLabels:
  - "bot/sync"
  - "apply-patches" # Example label for patches

# Optional: Configuration for Dockerfile generation
# Uses github.com/openshift-knative/hack/pkg/dockerfilegen
dockerfileGen:
  enabled: false
  # ... other dockerfilegen params

# Configuration for re-syncing a certain number of past releases from upstream
resyncReleases:
  enabled: true # Set to true to enable resyncing past releases
  numberOf: 3   # Number of past releases to resync if enabled

branches:
  # Main/default branch in your fork. Patches are typically applied here continuously.
  # This branch is often used as the base for `releaseNext`.
  main: "main"
  # `releaseNext` defines the pattern for the rolling "next release" branch in your fork.
  # It usually tracks the main development line of the upstream (e.g., upstream/main).
  # Deviate will determine the actual version (e.g., "release-1.23") based on upstream tags/branches.
  # Fork-specific patches are continuously applied to this line.
  releaseNext: "release-"
  # Branch prefix for CI trigger branches (e.g., "sync-ci-release-1.23")
  synchCi: "sync-ci-"
  releaseTemplates:
    # Go template for identifying/naming upstream release branches. {{ .Version }} is available.
    upstream: "release-{{ .Version }}"
    # Go template for naming downstream (fork) release branches. {{ .Version }} is available.
    downstream: "release-{{ .Version }}"
  searches:
    # Regex to find upstream release branches. Needs a `Version` capture group.
    upstreamReleases: '^release-(?P<Version>\d+\.\d+)$'
    # Regex to find downstream release branches. Needs a `Version` capture group.
    downstreamReleases: '^release-(?P<Version>\d+\.\d+)$'

tags:
  synchronize: true
  refSpec: "v*" # Example: sync all tags starting with 'v'

messages:
  triggerCi: "chore(sync): Trigger CI for {{ .ReleaseBranch }} into {{ .MainBranch }}"
  triggerCiBody: "Automated PR to trigger CI for syncing `{{ .ReleaseBranch }}` into `{{ .MainBranch }}`."
  applyForkFiles: "chore: Apply fork-specific files and patches"
  imagesGenerated: "chore: Generate images"
```

## How `deviate sync` Works

The `deviate sync` command orchestrates the synchronization process:

1.  **Loads Configuration**: Reads `.deviate.yaml`.
2.  **Fetches Remotes**: Updates local refs for upstream and downstream repositories.
3.  **Manages `release-next` Branch (Fork's Rolling Development Line)**:
    *   Ensures the fork's branch corresponding to `branches.releaseNext` (e.g., `release-1.24` if upstream's latest is `1.24`) is up-to-date with the upstream's main development line (e.g., `upstream/main`).
    *   Applies fork-specific patches to this `release-next` branch.
4.  **Processes Existing Release Branches**:
    *   For each existing release branch in the fork (e.g., `release-1.23`) that also exists upstream:
        *   If `resyncReleases.enabled` is true, `deviate` checks for new commits on the *upstream* release branch (e.g., `upstream/release-1.23`) that are not yet on the fork's corresponding release branch.
        *   If new commits are found, `deviate` creates a "resync PR" in the fork to bring these commits from the upstream release branch into the fork's release branch. This helps automate the process of incorporating upstream bug fixes or backports into your fork's maintained releases.
5.  **Creates New Release Branches**:
    *   If `deviate` identifies a new release upstream (e.g., upstream creates `release-1.25`) for which the fork does not yet have a corresponding branch:
        *   It creates a new release branch in the fork (e.g., `downstream/release-1.25`) based on the upstream one.
        *   Fork-specific patches are applied *once* to this newly created release branch.
6.  **Tag Syncing**: If `tags.synchronize` is true, syncs tags matching `tags.refSpec`.
7.  **CI Trigger and Sync PR for `release-next`**:
    *   Creates a temporary CI trigger branch (e.g., `sync-ci-release-1.24`) from the fork's `release-next` branch.
    *   Adds a small, timestamped file (`ci`) to this temporary branch and commits it (using `messages.triggerCi`). This ensures CI workflows run.
    *   Pushes this CI trigger branch.
    *   Creates a Pull Request to merge the CI trigger branch (e.g., `sync-ci-release-1.24`) into the actual `release-next` branch (e.g., `release-1.24`) in the fork. Labels from `syncLabels` are applied.

**Note on Upstream Branching**: `deviate`'s release management features work best when the upstream project also maintains release branches (e.g., `release-1.23`, `release-1.24`). If the upstream only uses a `main` branch for all development and releases, then `deviate` will primarily sync that `upstream/main` to your fork's `releaseNext` / `main` branch.

## Quickstart: GitHub Actions Workflow

To automate `deviate sync`, you can use a GitHub Actions workflow. Add the following to `.github/workflows/sync-upstream.yaml` in your forked repository:

```yaml
name: Sync Upstream

on:
  schedule:
    # Example: Run at 3 AM UTC every day
    - cron: '0 3 * * *'
  workflow_dispatch: {} # Allows manual triggering

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Fork
        uses: actions/checkout@v4
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
          fetch-depth: 0 # Fetch all history for all tags and branches

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21' # Adjust to your project's Go version

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
```

**Before using this workflow:**

1.  **Create `.deviate.yaml`**: Ensure a valid `.deviate.yaml` configuration file is in your fork.
2.  **Git User**: Update the Git user email and name.
3.  **Permissions**: The `GITHUB_TOKEN` usually has permissions to create PRs. For direct pushes to protected branches (if ever needed, though not typical for this flow), a PAT might be required.

## Auto-merging Sync PRs with Mergify

Use the `syncLabels` in `.deviate.yaml` to configure Mergify for auto-merging PRs created by `deviate`.

Set up Mergify via its dashboard (app.mergify.com) for your repository:

1.  **Install Mergify** on your fork.
2.  **Define `syncLabels`** in `.deviate.yaml` (e.g., `bot/sync`, `automerge-sync`).
3.  **Create Mergify Rules** to match these labels and define merge conditions.

   Example Mergify rule concept:

   ```yaml
   # In Mergify UI or .mergify.yml in your fork
   pull_request_rules:
     - name: Auto-merge deviate sync PRs
       conditions:
         - "label=bot/sync"
         - "status-success=Your_CI_Check_Name"
         # Add other conditions (no conflicts, etc.)
       actions:
         merge:
           method: squash # or merge, rebase
   ```

