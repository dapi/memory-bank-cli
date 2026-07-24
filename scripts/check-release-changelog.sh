#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: $0 vX.Y.Z" >&2
  exit 2
fi

release_version="$1"
numeric_identifier='(0|[1-9][0-9]*)'
prerelease_identifier="(${numeric_identifier}|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)"
semver_regex="^v${numeric_identifier}\\.${numeric_identifier}\\.${numeric_identifier}(-${prerelease_identifier}(\\.${prerelease_identifier})*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"

if ! [[ "$release_version" =~ $semver_regex ]]; then
  echo "release version must be a complete, v-prefixed semantic version" >&2
  exit 2
fi

changelog_version="${release_version#v}"
escaped_changelog_version="$(printf '%s\n' "$changelog_version" | sed 's/[][\\.^$*+?{}|()]/\\&/g')"
release_heading_regex="^## \\[${escaped_changelog_version}\\] - [0-9]{4}-[0-9]{2}-[0-9]{2}$"

if ! grep -Eq "$release_heading_regex" CHANGELOG.md; then
  echo "CHANGELOG.md must contain a dated ## [$changelog_version] release section" >&2
  exit 1
fi

echo "CHANGELOG.md contains release section for $release_version"
