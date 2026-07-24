#!/usr/bin/env bash
# Hermetic CLI-boundary tests for issue #21.  Every scenario owns fresh local
# Git repositories; no remote URL other than its temporary bare repository is
# used after this script starts.
set -euo pipefail

work_root="${E2E_WORK_ROOT:-$(mktemp -d)}"
binary="${E2E_BINARY:-}"
scope="${E2E_SCOPE:-full}"
keep_work="${E2E_KEEP_WORK:-0}"

cleanup() {
  if [ "$keep_work" != 1 ]; then rm -rf -- "$work_root"; else printf 'E2E workspace: %s\n' "$work_root"; fi
}
trap cleanup EXIT

fail() { printf 'FAIL %s: %s\n' "${case_name:-setup}" "$*" >&2; exit 1; }
require() { "$@" || fail "command failed: $*"; }
expect_fail() { if "$@"; then fail "command unexpectedly succeeded: $*"; fi; }
file() { printf '%s/%s' "$downstream" "$1"; }
source_file() { printf '%s/template/memory-bank/%s' "$source" "$1"; }
lock_file() { file memory-bank/.lock; }

test -n "$binary" || { printf 'E2E_BINARY must name a pre-built memory-bank-cli executable\n' >&2; exit 2; }
test -x "$binary" || { printf 'E2E_BINARY is not executable: %s\n' "$binary" >&2; exit 2; }

write_template_v1() {
  mkdir -p "$template_work/template/memory-bank/flows" "$template_work/template/memory-bank/dna" "$template_work/template/memory-bank/domain"
  cat >"$template_work/template/memory-bank/README.md" <<'EOF'
---
title: Fixture Memory Bank
doc_kind: guide
doc_function: canonical
purpose: fixture
status: active
---
# Fixture
EOF
  printf 'managed v1\n' >"$template_work/template/memory-bank/dna/managed.md"
  printf 'local-only v1\n' >"$template_work/template/memory-bank/dna/local-only.md"
  printf 'unchanged v1\n' >"$template_work/template/memory-bank/domain/unchanged.md"
  printf 'delete v1\n' >"$template_work/template/memory-bank/dna/delete.md"
  printf 'rename v1\n' >"$template_work/template/memory-bank/dna/rename-old.md"
  printf 'mode payload\n' >"$template_work/template/memory-bank/dna/mode.md"
  printf 'unrelated v1\n' >"$template_work/template/memory-bank/flows/unrelated.md"
}

write_template_v2() {

  mkdir -p "$template_work/template/memory-bank/dna/new-dir"
  printf 'managed v2\n' >"$template_work/template/memory-bank/dna/managed.md"
  printf 'created v2\n' >"$template_work/template/memory-bank/dna/created.md"
  printf 'nested v2\n' >"$template_work/template/memory-bank/dna/new-dir/item.md"
  rm "$template_work/template/memory-bank/dna/delete.md"
  mv "$template_work/template/memory-bank/dna/rename-old.md" "$template_work/template/memory-bank/dna/rename-new.md"
  printf 'unrelated v2\n' >"$template_work/template/memory-bank/flows/unrelated.md"
}

setup_case() {
  case_name="$1"
  case_root="$work_root/$case_name"
  remote="$case_root/template.git"; template_work="$case_root/template-work"; source="$case_root/source"; downstream="$case_root/downstream"
  mkdir -p "$case_root" "$template_work" "$downstream"
  require git init --quiet "$template_work"
  require git -C "$template_work" config user.name 'E2E Fixture'
  require git -C "$template_work" config user.email 'fixture@example.invalid'
  write_template_v1
  require git -C "$template_work" add .
  require git -C "$template_work" commit --quiet -m v1
  require git -C "$template_work" tag v1.0.0
  v1_sha="$(git -C "$template_work" rev-parse v1.0.0^{commit})"
  write_template_v2
  require git -C "$template_work" add -A
  require git -C "$template_work" commit --quiet -m v2
  require git -C "$template_work" tag v1.1.0
  v2_sha="$(git -C "$template_work" rev-parse v1.1.0^{commit})"
  require git init --bare --quiet "$remote"
  require git -C "$template_work" remote add origin "$remote"
  require git -C "$template_work" push --quiet --tags origin HEAD
  require git clone --quiet "$remote" "$source"
  require git -C "$source" checkout --quiet --detach "$v1_sha"
  require git init --quiet "$downstream"
  require git -C "$downstream" config user.name 'E2E Downstream'
  require git -C "$downstream" config user.email 'downstream@example.invalid'
}

init_v1() { require "$binary" init --repo-root "$downstream" --source "$source" --template-version v1.0.0 --source-ref "$v1_sha"; }
checkout_v2() { require git -C "$source" checkout --quiet --detach "$v2_sha"; }
update_v2() { "$binary" update --repo-root "$downstream" --source "$source" --template-version v1.1.0 --source-ref "$v2_sha"; }
snapshot() { rm -rf "$case_root/snapshot"; cp -R "$downstream" "$case_root/snapshot"; }
assert_unchanged() { diff -r -q --exclude .git "$case_root/snapshot" "$downstream" >/dev/null || fail 'downstream changed despite expected refusal'; }
assert_contains() { grep -Fq "$2" "$1" || fail "expected $1 to contain $2"; }
assert_absent() { test ! -e "$1" || fail "expected absent: $1"; }
assert_lock_v2() { assert_contains "$(lock_file)" '"version": "v1.1.0"'; assert_contains "$(lock_file)" "$v2_sha"; }
doctor_profile() {
  local root="$1" expected="$2" output status
  output="$case_root/doctor-$expected.json"
  set +e
  "$binary" doctor --repo-root "$root" --profile auto --json >"$output" 2>&1
  status=$?
  set -e
  test "$status" -eq 0 || test "$status" -eq 1 || fail "unexpected doctor status: $status"
  grep -Eq '"profile"[[:space:]]*:[[:space:]]*"'"$expected"'"' "$output" || fail "doctor did not select $expected"
}

scenario_01() { setup_case E2E-01; init_v1; test -f "$(file memory-bank/dna/managed.md)"; assert_contains "$(lock_file)" '"version": "v1.0.0"'; assert_contains "$(lock_file)" "$v1_sha"; assert_contains "$(lock_file)" '"payload_digest"'; }
scenario_02() { setup_case E2E-02; init_v1; snapshot; expect_fail "$binary" init --repo-root "$downstream" --source "$source" --template-version v1.0.0 --source-ref "$v1_sha"; assert_unchanged; }
scenario_03() { setup_case E2E-03; init_v1; checkout_v2; require update_v2; assert_contains "$(file memory-bank/dna/managed.md)" 'managed v2'; test -f "$(file memory-bank/dna/created.md)"; assert_absent "$(file memory-bank/dna/delete.md)"; assert_lock_v2; }
scenario_04() { setup_case E2E-04; init_v1; printf 'local edit\n' >"$(file memory-bank/dna/managed.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_05() { setup_case E2E-05; init_v1; printf 'user data\n' >"$(file memory-bank/user-owned.md)"; checkout_v2; require update_v2; assert_contains "$(file memory-bank/user-owned.md)" 'user data'; }
scenario_06() { setup_case E2E-06; init_v1; checkout_v2; snapshot; require "$binary" update --dry-run --repo-root "$downstream" --source "$source" --template-version v1.1.0 --source-ref "$v2_sha" >"$case_root/dry-run.txt"; assert_contains "$case_root/dry-run.txt" 'create'; assert_contains "$case_root/dry-run.txt" 'delete'; assert_unchanged; }
scenario_07() { setup_case E2E-07; init_v1; mkdir -p "$downstream/.memory-bank-template"; doctor_profile "$downstream" downstream; }
scenario_08() { setup_case E2E-08; doctor_profile "$source" template; }
scenario_09() { setup_case E2E-09; snapshot; expect_fail "$binary" init --repo-root "$downstream" --source "$source" --template-version bad --source-ref 0000000000000000000000000000000000000000; assert_unchanged; }
scenario_11() { setup_case E2E-11; init_v1; printf 'local unchanged edit\n' >"$(file memory-bank/dna/local-only.md)"; checkout_v2; require update_v2; assert_contains "$(file memory-bank/dna/local-only.md)" 'local unchanged edit'; assert_lock_v2; }
scenario_12() { setup_case E2E-12; init_v1; printf 'local edit\n' >"$(file memory-bank/dna/managed.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_13() { setup_case E2E-13; init_v1; printf 'managed v2\n' >"$(file memory-bank/dna/managed.md)"; checkout_v2; require update_v2; assert_lock_v2; }
scenario_14() { setup_case E2E-14; init_v1; chmod +x "$(file memory-bank/dna/mode.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_15() { setup_case E2E-15; init_v1; local lock_before; lock_before="$(cat "$(lock_file)")"; printf 'outside remains untouched\n' >"$case_root/outside"; rm "$(file memory-bank/dna/managed.md)"; ln -s "$case_root/outside" "$(file memory-bank/dna/managed.md)"; checkout_v2; expect_fail update_v2; test -L "$(file memory-bank/dna/managed.md)"; test "$(readlink "$(file memory-bank/dna/managed.md)")" = "$case_root/outside"; assert_contains "$case_root/outside" 'outside remains untouched'; test "$(cat "$(lock_file)")" = "$lock_before"; assert_absent "$(file memory-bank/dna/created.md)"; }
scenario_16() { setup_case E2E-16; init_v1; rm "$(file memory-bank/dna/managed.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_17() { setup_case E2E-17; init_v1; checkout_v2; require update_v2; assert_absent "$(file memory-bank/dna/delete.md)"; assert_lock_v2; }
scenario_18() { setup_case E2E-18; init_v1; printf 'local delete edit\n' >"$(file memory-bank/dna/delete.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_19() { setup_case E2E-19; init_v1; rm "$(file memory-bank/dna/delete.md)"; checkout_v2; require update_v2; assert_absent "$(file memory-bank/dna/delete.md)"; assert_lock_v2; }
scenario_20() { setup_case E2E-20; init_v1; checkout_v2; require update_v2; assert_absent "$(file memory-bank/dna/rename-old.md)"; test -f "$(file memory-bank/dna/rename-new.md)"; setup_case E2E-20-conflict; init_v1; printf 'local rename edit\n' >"$(file memory-bank/dna/rename-old.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_21() { setup_case E2E-21; init_v1; printf 'user collision\n' >"$(file memory-bank/dna/created.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_22() { setup_case E2E-22; init_v1; printf 'file blocks dir\n' >"$(file memory-bank/dna/new-dir)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_23() { setup_case E2E-23; init_v1; mkdir -p "$(file memory-bank/user-dir)"; printf 'keep\n' >"$(file memory-bank/user-dir/keep.md)"; checkout_v2; require update_v2; assert_contains "$(file memory-bank/user-dir/keep.md)" keep; }
scenario_24() { setup_case E2E-24; init_v1; printf 'conflict\n' >"$(file memory-bank/dna/managed.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_25() { setup_case E2E-25; init_v1; snapshot; expect_fail "$binary" update --repo-root "$downstream" --source "$source" --template-version bad --source-ref 0000000000000000000000000000000000000000; assert_unchanged; }
scenario_26() { setup_case E2E-26; init_v1; printf 'conflict\n' >"$(file memory-bank/dna/managed.md)"; checkout_v2; snapshot; expect_fail update_v2; assert_unchanged; snapshot; expect_fail update_v2; assert_unchanged; }
scenario_27() { setup_case E2E-27; init_v1; printf 'local edit\n' >"$(file memory-bank/dna/managed.md)"; checkout_v2; expect_fail update_v2; printf 'managed v2\n' >"$(file memory-bank/dna/managed.md)"; require update_v2; assert_lock_v2; }

run_case() { local requested="$1"; case_name="$requested"; printf 'RUN %s\n' "$requested"; "scenario_${requested#E2E-}"; printf 'PASS %s\n' "$requested"; }

if [ -n "${E2E_CASES:-}" ]; then
  for number in $E2E_CASES; do run_case "E2E-$number"; done
elif [ "$scope" = release ]; then
  run_case E2E-01
  run_case E2E-03
else
  for number in 01 02 03 04 05 06 07 08 09 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27; do run_case "E2E-$number"; done
fi

printf 'All requested local E2E scenarios passed.\n'
