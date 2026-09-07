#!/usr/bin/env bash
# Real-binary source-format contract. No mocking of ownership or Git verification.
set -euo pipefail
: "${E2E_BINARY:?point E2E_BINARY at the built bridge}"
: "${LEGACY_SOURCE:?point LEGACY_SOURCE at the clean pinned legacy checkout}"
legacy_ref=f1f04de843aef45a2425d4a7351d577bbf89e940
work_root="$(mktemp -d)"
trap 'rm -rf "$work_root"' EXIT
mkdir "$work_root/downstream" "$work_root/unknown" "$work_root/component"
"$E2E_BINARY" capabilities --require source-format/v1 --require legacy/v1
"$E2E_BINARY" capabilities --require components/v1 --require adoption/v1
"$E2E_BINARY" init --repo-root "$work_root/downstream" --source "$LEGACY_SOURCE" --source-ref "$legacy_ref" --template-version legacy-f1f04de --json >"$work_root/init.json"
"$E2E_BINARY" pull --repo-root "$work_root/downstream" --source "$LEGACY_SOURCE" --source-ref "$legacy_ref" --template-version legacy-f1f04de --json >"$work_root/pull.json"
cp -R "$work_root/downstream" "$work_root/before"
for format in unknown component; do
  source_root="$work_root/$format"
  mkdir -p "$source_root/template/memory-bank"
  printf 'incoming\n' >"$source_root/template/memory-bank/README.md"
  if [ "$format" = component ]; then
    printf '%s\n' '{"schema_version":1,"payload_format":"components/v1","capabilities":["components/v1"]}' >"$source_root/memory-bank-source.json"
    printf '{}\n' >"$source_root/template/memory-bank/components.json"
  fi
  git -C "$source_root" init --quiet
  git -C "$source_root" add .
  git -C "$source_root" -c user.name=Fixture -c user.email=fixture@example.invalid commit --quiet -m "$format"
  source_ref="$(git -C "$source_root" rev-parse HEAD)"
  if "$E2E_BINARY" pull --repo-root "$work_root/downstream" --source "$source_root" --source-ref "$source_ref" --template-version fixture >"$work_root/$format.out" 2>"$work_root/$format.err"; then
    echo "bridge accepted $format source" >&2; exit 1
  fi
  diff -r "$work_root/before" "$work_root/downstream"
  mkdir "$work_root/new-$format"
  if "$E2E_BINARY" init --repo-root "$work_root/new-$format" --source "$source_root" --source-ref "$source_ref" --template-version fixture >"$work_root/$format-init.out" 2>"$work_root/$format-init.err"; then
    echo "bridge initialized $format source" >&2; exit 1
  fi
  test -z "$(ls -A "$work_root/new-$format")"
done
printf 'Real bridge source-format fixtures passed\n'
