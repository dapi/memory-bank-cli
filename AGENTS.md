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
4. Run the Go tests, vet, GoReleaser configuration check, snapshot build, and
   release E2E validation before publishing.
5. Publish through `.github/workflows/release.yml`, wait for it to complete,
   install the exact released tag locally, and verify `memory-bank-cli
   --version`.

The release workflow enforces step 3 with
`scripts/check-release-changelog.sh`.
