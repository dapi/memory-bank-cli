#!/usr/bin/env bash
set -euo pipefail

repository_url="${REPOSITORY_URL:-https://github.com/dapi/memory-bank-cli.git}"
cli_ref="${CLI_REF:?CLI_REF is required}"
template_ref="${TEMPLATE_REF:?TEMPLATE_REF is required}"
release_tag="${RELEASE_TAG:-}"
report_dir="${REPORT_DIR:-$PWD/artifacts/downstream-smoke}"
phase="external-tooling"
step="setup"
report="$report_dir/report.txt"
workspace=""

write_report() {
  local result="$1"
  {
    printf 'result=%s\n' "$result"
    printf 'phase=%s\n' "$phase"
    printf 'step=%s\n' "$step"
    printf 'requested_cli_ref=%s\n' "$cli_ref"
    printf 'requested_template_ref=%s\n' "$template_ref"
    printf 'release_tag=%s\n' "$release_tag"
    if [ -n "${cli_sha:-}" ]; then printf 'resolved_cli_sha=%s\n' "$cli_sha"; fi
    if [ -n "${cli_install_ref:-}" ]; then printf 'installed_cli_ref=%s\n' "$cli_install_ref"; fi
    if [ -n "${template_sha:-}" ]; then printf 'resolved_template_sha=%s\n' "$template_sha"; fi
    go version 2>&1 || true
    git --version 2>&1 || true
  } >"$report"
}

on_error() {
  local status=$?
  set +e
  write_report "failed"
  exit "$status"
}
trap on_error ERR
trap 'if [ -n "$workspace" ] && [ -d "$workspace" ]; then rm -rf -- "$workspace"; fi' EXIT

run_step() {
  step="$1"
  shift
  "$@" > >(tee "$report_dir/${step}.log") 2> >(tee -a "$report_dir/${step}.log" >&2)
}

assert_contains() {
  local path="$1"
  local expected="$2"
  if ! grep -Fq "$expected" "$path"; then
    printf 'expected %s to contain: %s\n' "$path" "$expected" >&2
    return 1
  fi
}

assert_digest() {
  local path="$1"
  local expected="$2"
  local actual
  actual="$(sha256sum "$path" | awk '{ print $1 }')"
  if [ "$actual" != "$expected" ]; then
    printf 'content digest changed for %s: expected %s, got %s\n' "$path" "$expected" "$actual" >&2
    return 1
  fi
}

assert_clean_repository() {
  local root="$1"
  local status_output
  status_output="$(git -C "$root" status --porcelain=v1)"
  if [ -n "$status_output" ]; then
    printf 'repository is not clean:\n%s\n' "$status_output" >&2
    return 1
  fi
}

# Keep setup and input resolution in the external-tooling boundary: neither a
# runner/tool failure nor an unresolved remote ref is evidence about the
# template, packaging, or CLI lifecycle itself.
mkdir -p "$report_dir"
workspace="$(mktemp -d)"

command -v git >/dev/null
command -v go >/dev/null
command -v curl >/dev/null
command -v jq >/dev/null
command -v sha256sum >/dev/null
go version >/dev/null
git --version >/dev/null

resolve_ref() {
  local ref="$1"
  local sha
  if [[ "$ref" =~ ^[0-9a-f]{40}$ ]]; then
    local resolver="$workspace/ref-resolution"
    if [ ! -d "$resolver/.git" ]; then
      git init --quiet "$resolver"
    fi
    git -C "$resolver" fetch --quiet --depth=1 "$repository_url" "$ref"
    sha="$(git -C "$resolver" rev-parse --verify 'FETCH_HEAD^{commit}')"
    test "$sha" = "$ref" || return 1
  else
    sha="$(git ls-remote "$repository_url" "${ref}^{}" | awk 'NR == 1 { print $1 }')"
    if [ -z "$sha" ]; then
      sha="$(git ls-remote "$repository_url" "$ref" | awk 'NR == 1 { print $1 }')"
    fi
  fi
  test -n "$sha" || return 1
  printf '%s\n' "$sha"
}

step="resolve-cli-ref"
cli_sha="$(resolve_ref "$cli_ref")"
cli_install_ref="$cli_sha"
if [ -n "$release_tag" ] && [ "$cli_ref" = "$release_tag" ] &&
  [[ "$cli_ref" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  # Stable release lanes must exercise the documented semantic-version
  # consumer command. Canary inputs remain bound to their resolved SHA.
  cli_install_ref="$cli_ref"
fi
step="resolve-template-ref"
template_sha="$(resolve_ref "$template_ref")"

source_root="$workspace/template"
downstream_root="$workspace/downstream"
bin_root="$workspace/bin"

phase="template"
step="clone-template"
git clone --quiet "$repository_url" "$source_root"
step="checkout-template"
git -C "$source_root" checkout --quiet --detach "$template_sha"
step="verify-template-clean"
test -z "$(git -C "$source_root" status --porcelain=v1 --untracked-files=all)"

phase="packaging"
step="install-cli"
GOBIN="$bin_root" go install "github.com/dapi/memory-bank-cli/cmd/memory-bank-cli@${cli_install_ref}"
cli="$bin_root/memory-bank-cli"
step="verify-cli-executable"
test -x "$cli"

if [ -n "$release_tag" ]; then
  step="fetch-release-metadata"
  release_api="https://api.github.com/repos/dapi/memory-bank-cli/releases/tags/${release_tag}"
  release_json="$workspace/release.json"
  curl --fail --silent --show-error --location "$release_api" -o "$release_json"
  step="discover-release-assets"
  checksums_url="$(jq -r '[.assets[]? | select(.name == "checksums.txt") | .browser_download_url] | first // empty' "$release_json")"
  if [ -n "$checksums_url" ] && jq -e 'any(.assets[]?; .name != "checksums.txt" and (.name | startswith("memory-bank-cli-")))' "$release_json" >/dev/null; then
    step="verify-release-assets"
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
step="prepare-downstream"
mkdir -p "$downstream_root"
git -C "$downstream_root" init --quiet
printf '# Downstream fixture\n' >"$downstream_root/README.md"
run_step "init" "$cli" init --repo-root "$downstream_root" --source "$source_root" --template-version "$template_ref" --source-ref "$template_sha"

adapted_file="$downstream_root/memory-bank/domain/model.md"
user_owned_file="$downstream_root/memory-bank/features/downstream-owned.txt"
printf '\nDownstream adaptation.\n' >>"$adapted_file"
printf 'Downstream-owned file.\n' >"$user_owned_file"
user_owned_digest="$(sha256sum "$user_owned_file" | awk '{ print $1 }')"
run_step "update-preservation" "$cli" update --repo-root "$downstream_root" --source "$source_root" --template-version "$template_ref" --source-ref "$template_sha"
run_step "verify-adaptation" assert_contains "$adapted_file" "Downstream adaptation."
run_step "verify-user-owned-content" assert_digest "$user_owned_file" "$user_owned_digest"
step="commit-baseline"
git -C "$downstream_root" add --all
git -C "$downstream_root" -c user.name='Downstream smoke' -c user.email='smoke@example.invalid' commit --quiet -m 'fixture baseline'
run_step "update-idempotence" "$cli" update --repo-root "$downstream_root" --source "$source_root" --template-version "$template_ref" --source-ref "$template_sha"
run_step "verify-no-diff" git -C "$downstream_root" diff --exit-code
run_step "verify-clean-repository" assert_clean_repository "$downstream_root"
run_step "lint" "$cli" lint --repo-root "$downstream_root"
run_step "doctor" "$cli" doctor --repo-root "$downstream_root" --profile downstream

step="complete"
write_report "passed"
