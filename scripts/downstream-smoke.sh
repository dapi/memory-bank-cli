#!/usr/bin/env bash
set -euo pipefail

repository_url="${REPOSITORY_URL:-https://github.com/dapi/memory-bank-cli.git}"
cli_ref="${CLI_REF:?CLI_REF is required}"
template_ref="${TEMPLATE_REF:?TEMPLATE_REF is required}"
release_tag="${RELEASE_TAG:-}"
report_dir="${REPORT_DIR:-$PWD/artifacts/downstream-smoke}"
phase="external-tooling"

mkdir -p "$report_dir"
report="$report_dir/report.txt"
workspace="$(mktemp -d)"

write_report() {
  local result="$1"
  {
    printf 'result=%s\n' "$result"
    printf 'phase=%s\n' "$phase"
    printf 'requested_cli_ref=%s\n' "$cli_ref"
    printf 'requested_template_ref=%s\n' "$template_ref"
    printf 'release_tag=%s\n' "$release_tag"
    if [ -n "${cli_sha:-}" ]; then printf 'resolved_cli_sha=%s\n' "$cli_sha"; fi
    if [ -n "${template_sha:-}" ]; then printf 'resolved_template_sha=%s\n' "$template_sha"; fi
    go version
    git --version
  } >"$report"
}

on_error() {
  local status=$?
  write_report "failed"
  exit "$status"
}
trap on_error ERR
trap 'rm -rf "$workspace"' EXIT

source_root="$workspace/template"
downstream_root="$workspace/downstream"
bin_root="$workspace/bin"

phase="template"
git clone --quiet "$repository_url" "$source_root"
git -C "$source_root" checkout --quiet --detach "$template_ref"
template_sha="$(git -C "$source_root" rev-parse HEAD)"
test -z "$(git -C "$source_root" status --porcelain=v1 --untracked-files=all)"

phase="packaging"
GOBIN="$bin_root" go install "github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@${cli_ref}"
cli="$bin_root/memory-bank-cli"
test -x "$cli"
cli_sha="$(git ls-remote "$repository_url" "${cli_ref}^{}" | awk 'NR == 1 { print $1 }')"
if [ -z "$cli_sha" ]; then
  cli_sha="$(git ls-remote "$repository_url" "$cli_ref" | awk 'NR == 1 { print $1 }')"
fi
if [ -z "$cli_sha" ]; then
  cli_sha="$cli_ref"
fi

if [ -n "$release_tag" ]; then
  release_api="https://api.github.com/repos/dapi/memory-bank-cli/releases/tags/${release_tag}"
  release_json="$workspace/release.json"
  curl --fail --silent --show-error --location "$release_api" -o "$release_json"
  checksums_url="$(jq -r '.assets[] | select(.name == "checksums.txt") | .browser_download_url' "$release_json")"
  if [ "$checksums_url" != "null" ]; then
    curl --fail --silent --show-error --location "$checksums_url" -o "$workspace/checksums.txt"
    while IFS=$'\t' read -r asset_name asset_url; do
      asset_path="$workspace/$asset_name"
      curl --fail --silent --show-error --location "$asset_url" -o "$asset_path"
      expected="$(awk -v name="$asset_name" '$2 == name { print $1 }' "$workspace/checksums.txt")"
      actual="$(sha256sum "$asset_path" | awk '{ print $1 }')"
      test -n "$expected"
      test "$actual" = "$expected"
    done < <(jq -r '.assets[] | select(.name != "checksums.txt" and (.name | startswith("memory-bank-cli-"))) | [.name, .browser_download_url] | @tsv' "$release_json")
  fi
fi

phase="cli"
mkdir -p "$downstream_root"
git -C "$downstream_root" init --quiet
"$cli" init --repo-root "$downstream_root" --source "$source_root" --template-version "$template_ref" --source-ref "$template_sha"

adapted_file="$downstream_root/memory-bank/domain/model.md"
user_owned_file="$downstream_root/memory-bank/features/downstream-owned.md"
printf '\nDownstream adaptation.\n' >>"$adapted_file"
printf '# Downstream-owned file\n' >"$user_owned_file"
"$cli" update --repo-root "$downstream_root" --source "$source_root" --template-version "$template_ref" --source-ref "$template_sha"
grep -q 'Downstream adaptation.' "$adapted_file"
test -f "$user_owned_file"
git -C "$downstream_root" add --all
git -C "$downstream_root" -c user.name='Downstream smoke' -c user.email='smoke@example.invalid' commit --quiet -m 'fixture baseline'
"$cli" update --repo-root "$downstream_root" --source "$source_root" --template-version "$template_ref" --source-ref "$template_sha"
git -C "$downstream_root" diff --exit-code
test -z "$(git -C "$downstream_root" status --porcelain=v1)"
"$cli" lint --repo-root "$downstream_root"
"$cli" doctor --repo-root "$downstream_root" --profile downstream

write_report "passed"
