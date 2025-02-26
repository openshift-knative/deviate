#!/usr/bin/env bash

set -Eeuo pipefail

branch_name=autobump
deps=()

function usage {
  cat 1>&2 <<-EOL

Usage: $0 [options]

Bumps given Golang deps within the project.

Options:

 -b,--branch <branch-name>  Which branch to use for bumping

 -d,--dep <dep-name>        List of dependencies to bump. Repeat the option to
                            pass multiple IP addresses.

EOL
}

while [[ $# -gt 0 ]]; do
  case $1 in
    -b|--branch)
      branch_name="$2"
      shift
      shift
      ;;
    -d|--dep)
      deps+=("$2")
      shift
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      error "Unknown option $1"
      usage
      exit 1
      ;;
  esac
done

prexists="$(gh pr list --label 'ci/autobump' --author '@me' | wc -l)"

set -x

if (( prexists )); then
  gh pr checkout
else
  git checkout -b "$branch_name"
fi

for dep in "${deps[@]}"; do
  go get -u "$dep"
done

go mod tidy
go work sync
git add .

if [ "$(git status --porcelain | wc -l)" -eq 0 ]; then
  echo 'no changes'
  exit 0
fi

git commit -m 'Autobump of deps'

if (( prexists )); then
  git push
else
  gh pr create \
    --title ':robot: Autobump of deps' \
    --body 'This is automated PR' \
    --label 'skip-review' \
    --label 'kind/cleanup' \
    --label 'ci/autobump'
fi
