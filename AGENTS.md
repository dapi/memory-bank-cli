# Repository instructions

## Releases

When asked to make a release, update `CHANGELOG.md` as part of the release
without waiting for a separate request:

1. Choose the next semantic version from the user-visible changes since the
   latest release.
2. Move the relevant entries from `Unreleased` into a
   `## [X.Y.Z] - YYYY-MM-DD` section and leave an empty `Unreleased` section.
3. Keep the changelog entry in the same commit that will be tagged. Do not
   publish a release from a commit that lacks its versioned changelog section.
4. Treat the required GitHub Actions checks as the authoritative release
   validation gate. Local Go tests, vet, GoReleaser checks, snapshot builds,
   and release E2E runs are optional diagnostics; do not repeat them or block
   publication when the equivalent CI checks are green.
5. From the release commit on `main`, create and push the annotated
   `vX.Y.Z` tag. The tag-triggered `.github/workflows/release.yml` validates,
   builds, and publishes the release. Wait for it to complete, install the
   exact released tag locally, and verify `memory-bank-cli --version`.

The release workflow enforces step 3 with
`scripts/check-release-changelog.sh`.
