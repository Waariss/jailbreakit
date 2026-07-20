# Releasing

Releases are tag-driven. Pushing to `main` does not create a public release.

1. Ensure CI passes.
2. Add or update `docs/RELEASE-vX.Y.Z.md`.
3. Confirm `jailbreakit version` behavior locally and review the version resolver.
4. Create an annotated tag: `git tag -a vX.Y.Z -m "jailbreakit vX.Y.Z"`.
5. Push the tag: `git push origin vX.Y.Z`.
6. Let `.github/workflows/release.yml` run its tests and builds.
7. Verify the four raw binary assets and `SHA256SUMS` in the GitHub Release.
8. Test `install.sh` against the release.
9. Test `go install github.com/Waariss/jailbreakit/cmd/jailbreakit@vX.Y.Z`.
10. Update the Homebrew Formula with `./scripts/update-homebrew-formula.sh vX.Y.Z`.
11. Copy the tested Formula to `Waariss/homebrew-tap`.
12. Test `brew install waariss/tap/jailbreakit` after the tap update.

The release workflow injects tag, commit, date, and author metadata into release binaries. `go install package@version` gets the module version from Go build metadata; local builds use `dev` when no release metadata is available.

Do not create a release until the release notes, tests, and Formula update are reviewed. The project is for authorized iOS security testing and lab setup only.
