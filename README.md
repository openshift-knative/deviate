# Deviate

[![Mage](https://github.com/openshift-knative/deviate/actions/workflows/mage.yaml/badge.svg)](https://github.com/openshift-knative/deviate/actions/workflows/mage.yaml)
[![Lints](https://github.com/openshift-knative/deviate/actions/workflows/lint.yaml/badge.svg)](https://github.com/openshift-knative/deviate/actions/workflows/lint.yaml)
[![Go Report Card](https://goreportcard.com/badge/openshift-knative/deviate)](https://goreportcard.com/report/openshift-knative/deviate)
[![Releases](https://img.shields.io/github/release-pre/openshift-knative/deviate.svg?sort=semver)](https://github.com/openshift-knative/deviate/releases)
[![LICENSE](https://img.shields.io/github/license/openshift-knative/deviate.svg)](https://github.com/openshift-knative/deviate/blob/main/LICENSE)

Deviate automates the synchronization of forked (downstream/midstream) repositories
with their upstream counterparts. It mirrors upstream release branches, overlays
fork-specific files, applies carried patches, optionally generates Dockerfiles,
and opens pull requests, all in a single `deviate sync` invocation.

Originally built for the OpenShift Serverless workflow, deviate is designed to
work with any project that maintains a fork of an upstream repository.

## How it works

When you run `deviate sync`, the tool executes these steps in order:

1. **Mirror releases** - Detects upstream release branches that don't exist downstream yet and creates them.
2. **Sync tags** - Synchronizes version tags from upstream to downstream.
3. **Sync release-next** - Resets the `release-next` branch to upstream's `main`, then applies fork customizations.
4. **Apply fork files** - For each synced branch:
   - Removes unwanted upstream files (configured via `deleteFromUpstream` filters).
   - Overlays midstream-specific files from the downstream repo's `main` branch (configured via `copyFromMidstream` filters).
   - Applies any `.patch` files found in `openshift/patches/`.
5. **Generate Dockerfiles** - Optionally rewrites Dockerfile image references for downstream registries.
6. **Create PR** - Opens a GitHub pull request for the `release-next` branch so CI can validate the sync.
7. **Re-sync recent releases** - Optionally re-syncs the N most recent release branches to pick up upstream hotfixes.

```
upstream/main ──► mirror to downstream ──► delete unwanted files
                                               │
                                               ▼
                                    overlay fork-specific files
                                               │
                                               ▼
                                      apply carried patches
                                               │
                                               ▼
                                     generate Dockerfiles (opt)
                                               │
                                               ▼
                                        push & create PR
```

## Installation

### From releases

Download the latest binary from the [releases page](https://github.com/openshift-knative/deviate/releases):

```bash
# Linux (amd64)
curl -Lo deviate https://github.com/openshift-knative/deviate/releases/latest/download/deviate-linux-amd64
chmod +x deviate
sudo mv deviate /usr/local/bin/

# macOS (arm64)
curl -Lo deviate https://github.com/openshift-knative/deviate/releases/latest/download/deviate-darwin-arm64
chmod +x deviate
sudo mv deviate /usr/local/bin/
```

### From source

```bash
git clone https://github.com/openshift-knative/deviate.git
cd deviate
go build -o deviate ./cmd/deviate
```

### Prerequisites

- **Git** with SSH key or agent configured (for push access to your downstream repo)
- **GitHub CLI** (`gh`) authenticated (for PR creation). Deviate embeds the `gh` client library, but relies on `gh`'s authentication configuration.

## Quickstart

### 1. Set up your fork repository

Your downstream repository should have two git remotes:

```bash
cd my-downstream-repo
git remote -v
# origin    git@github.com:my-org/my-downstream.git (fetch/push)  ← or "downstream"
# upstream  https://github.com/upstream-org/upstream-project.git (fetch/push)
```

Deviate auto-detects remote URLs from git config. If your downstream remote
is named `origin`, it will be used automatically. If it's named `downstream`,
that takes precedence.

### 2. Create a configuration file

Create `.deviate.yaml` in your repository root:

```yaml
# Required: upstream and downstream repo URLs
# These can also be auto-detected from git remotes (see "Configuration" below).
upstream: https://github.com/upstream-org/upstream-project.git
downstream: git@github.com:my-org/my-downstream.git

# Files to delete from upstream after mirroring
deleteFromUpstream:
  include:
    - ".github/workflows/upstream-*.yaml"   # Remove upstream CI that won't work downstream
    - "OWNERS"                              # Remove upstream ownership file

# Files to copy from downstream's main branch onto each synced branch
copyFromMidstream:
  include:
    - "**"            # Copy everything from downstream main
  exclude:
    - ".git/**"       # Never copy .git internals

# Labels applied to sync PRs (for filtering in GitHub)
syncLabels:
  - "kind/sync-fork-to-upstream"

# Branch configuration
branches:
  main: "main"                                     # Downstream main branch
  releaseNext: "release-next"                      # Branch tracking upstream HEAD
  checkPrPrefix: "ci/"                             # Prefix for sync PR branches
  releaseTemplates:
    upstream: "release-{{ .Major }}.{{ .Minor }}"   # Upstream release branch pattern
    downstream: "release-{{ .Major }}.{{ .Minor }}" # Downstream release branch pattern
  searches:
    upstreamReleases: '^release-(\d+)\.(\d+)$'      # Regex to find upstream releases
    downstreamReleases: '^release-(\d+)\.(\d+)$'     # Regex to find downstream releases

# Tag synchronization
tags:
  synchronize: true
  refSpec: "v*"       # Tag glob pattern to sync

# Re-sync recent releases to pick up upstream hotfixes
resyncReleases:
  enabled: true
  numberOfReleases: 6    # Re-sync the 6 most recent releases

# Commit messages for sync operations
messages:
  triggerCi: ":robot: Synchronize branch `%s` to `upstream/%s`"
  triggerCiBody: "This automated PR is to make sure the forked project's `%s` branch (forked upstream's `%s` branch) passes a CI."
  applyForkFiles: ":open_file_folder: Apply fork specific files"
  imagesGenerated: ":vhs: Images generated"

# Dockerfile generation (optional, skip if not needed)
dockerfileGen:
  skip: true
```

### 3. Add fork-specific files

Place any downstream-specific files (Dockerfiles, CI workflows, Makefiles,
OWNERS files) in your downstream repo's `main` branch. The `copyFromMidstream`
filters control which files are overlaid onto synced branches.

### 4. Add carried patches (optional)

If you need to modify upstream source code (e.g., adding build tags, patching
dependencies), create patch files:

```bash
mkdir -p openshift/patches

# Create a patch from a local change
git diff > openshift/patches/001-add-build-tags.patch

# Or generate from a commit
git format-patch -1 HEAD --stdout > openshift/patches/002-fix-dependency.patch
```

Patches are applied in alphabetical order after fork files are overlaid.
Prefix with numbers (001-, 002-) to control ordering.

### 5. Run the sync

```bash
# Sync all releases and release-next
deviate sync

# Sync a specific project directory (if not running from repo root)
deviate sync /path/to/my-downstream-repo

# Dry run (skip push and PR creation)
# Set via config: dryRun: true
```

## Configuration reference

### Configuration file

Deviate reads `.deviate.yaml` from the repository root by default. Override
with `--config`:

```bash
deviate sync --config path/to/config.yaml
```

### Auto-detection from git remotes

The `upstream` and `downstream` URLs can be omitted from the config file.
Deviate will auto-detect them from your git remotes:

- **upstream**: Read from the `upstream` remote
- **downstream**: Read from the `downstream` remote, falling back to `origin`

This means a minimal config can be as short as:

```yaml
dockerfileGen:
  skip: true
```

If your git remotes are properly named.

### Environment variable overrides

All configuration fields can be overridden via environment variables prefixed
with `DEVIATE_`. Nested fields use underscores:

```bash
export DEVIATE_UPSTREAM=https://github.com/other/repo.git
export DEVIATE_DRYRUN=true
deviate sync
```

### Configuration fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `upstream` | string | yes* | git remote `upstream` | Upstream repository URL |
| `downstream` | string | yes* | git remote `downstream` or `origin` | Downstream repository URL |
| `dryRun` | bool | no | `false` | Skip push and PR operations |
| `deleteFromUpstream` | filters | yes | see defaults | Files to remove from upstream |
| `copyFromMidstream` | filters | yes | `include: ["**"]` | Files to overlay from downstream `main` |
| `syncLabels` | []string | yes | `["kind/sync-fork-to-upstream"]` | Labels for sync PRs |
| `branches.main` | string | yes | `"main"` | Downstream main branch name |
| `branches.releaseNext` | string | yes | `"release-next"` | Branch tracking upstream HEAD |
| `branches.checkPrPrefix` | string | no | `"ci/"` | Prefix for sync PR branches |
| `branches.skipCheckPr` | bool | no | `false` | Skip PR creation |
| `branches.releaseTemplates.upstream` | string | yes | `"release-{{ .Major }}.{{ .Minor }}"` | Upstream release branch template |
| `branches.releaseTemplates.downstream` | string | yes | `"release-{{ .Major }}.{{ .Minor }}"` | Downstream release branch template |
| `branches.searches.upstreamReleases` | string | yes | `'^release-(\d+)\.(\d+)$'` | Regex for upstream release branches |
| `branches.searches.downstreamReleases` | string | yes | `'^release-(\d+)\.(\d+)$'` | Regex for downstream release branches |
| `tags.synchronize` | bool | no | `false` | Whether to sync tags |
| `tags.refSpec` | string | yes | `"v*"` | Tag glob pattern to sync |
| `resyncReleases.enabled` | bool | no | `false` | Re-sync recent releases |
| `resyncReleases.numberOfReleases` | int | no | `6` | How many recent releases to re-sync |
| `dockerfileGen.skip` | bool | no | `false` | Skip Dockerfile generation |
| `messages.triggerCi` | string | yes | (see defaults) | PR title template |
| `messages.triggerCiBody` | string | yes | (see defaults) | PR body template |
| `messages.applyForkFiles` | string | yes | (see defaults) | Commit message for fork files |
| `messages.imagesGenerated` | string | yes | (see defaults) | Commit message for image generation |

*Auto-detected from git remotes if not specified.

### File filters

The `deleteFromUpstream` and `copyFromMidstream` fields use glob-based filters
with `include` and `exclude` lists. Globs use `/` as the path separator:

```yaml
deleteFromUpstream:
  include:
    - ".github/workflows/knative-*.yaml"    # Match specific workflows
    - "vendor/**"                           # Match entire directory trees
  exclude:
    - ".github/workflows/knative-release.yaml"  # Keep this one
```

A file is included if it matches any `include` pattern and does not match any
`exclude` pattern.

### Carried patches

Patch files are read from `openshift/patches/` within the project directory.
Only files with a `.patch` extension are processed. They are applied using
`git apply` in alphabetical order.

Typical use cases:
- Adding downstream build tags or compile-time flags
- Patching dependencies that haven't been updated upstream yet
- Modifying default configuration values

If a patch fails to apply (e.g., due to upstream code changes), the sync
operation fails. Update the patch to match the current upstream before
re-running.

## Examples

### Minimal configuration (auto-detect remotes)

```yaml
# .deviate.yaml
dockerfileGen:
  skip: true
```

Requires git remotes `upstream` and `origin` (or `downstream`) to be configured.

### Knative Serverless component

```yaml
# .deviate.yaml
upstream: https://github.com/knative/eventing.git

deleteFromUpstream:
  include:
    - ".github/workflows/knative-*.yaml"

copyFromMidstream:
  include:
    - "**"

syncLabels:
  - "kind/sync-fork-to-upstream"

dockerfileGen:
  skip: false
  images-from:
    - eventing
```

### Non-Knative project with custom release naming

```yaml
# .deviate.yaml
upstream: https://github.com/example/project.git
downstream: git@github.com:my-org/project-fork.git

deleteFromUpstream:
  include:
    - ".github/**"
    - "CODEOWNERS"

copyFromMidstream:
  include:
    - ".tekton/**"
    - "Dockerfile*"
    - "Makefile"
    - "openshift/**"

branches:
  main: "main"
  releaseNext: "next"
  releaseTemplates:
    upstream: "v{{ .Major }}.{{ .Minor }}"
    downstream: "rh-{{ .Major }}.{{ .Minor }}"
  searches:
    upstreamReleases: '^v(\d+)\.(\d+)$'
    downstreamReleases: '^rh-(\d+)\.(\d+)$'

tags:
  synchronize: true
  refSpec: "v*"

resyncReleases:
  enabled: true
  numberOfReleases: 3

dockerfileGen:
  skip: true
```

### CI automation (GitHub Actions)

```yaml
# .github/workflows/sync-upstream.yaml
name: Sync upstream
on:
  schedule:
    - cron: '0 6 * * *'  # Daily at 6:00 UTC
  workflow_dispatch: {}

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Install deviate
        run: |
          curl -Lo deviate https://github.com/openshift-knative/deviate/releases/latest/download/deviate-linux-amd64
          chmod +x deviate
          sudo mv deviate /usr/local/bin/
      - name: Configure git
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git remote add upstream https://github.com/upstream-org/project.git
      - name: Run sync
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: deviate sync
```

## Architecture

```
cmd/deviate/          CLI entrypoint (cobra)
internal/cmd/         Command definitions (sync)
pkg/
  cli/                High-level Sync() entry point
  config/             Configuration loading, defaults, validation
    git/              Git remote/checkout abstractions
  files/              Glob-based file filters and directory utilities
  git/                Git operations (clone, fetch, push, merge, commit)
  github/             GitHub CLI client wrapper (PR creation)
  sync/               Core sync pipeline (mirror, fork files, patches, PRs)
  state/              Runtime state (config + git repo + context)
  log/                Structured logging with color support
  metadata/           Tool name and version
```

## Development

Build and test using [Mage](https://magefile.org/):

```bash
# Install mage
go install github.com/magefile/mage@latest

# Build
mage build

# Run tests
mage test

# Lint
mage lint
```

Or directly with Go:

```bash
go build -o deviate ./cmd/deviate
go test ./...
```

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
