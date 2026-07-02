# Release Notes: v1.3.2

## Summary

`jailbreakit` v1.3.2 is a distribution polish release. It makes the project easier to install from GitHub Releases without requiring users to clone and build from source.

## Changes

- Added `install.sh` for installing the latest or a pinned GitHub Release binary.
- Added checksum verification against release `SHA256SUMS` when available.
- Updated README install instructions for release binaries.
- Bumped the local version to `v1.3.2`.

## Install

Install latest release binary:

```sh
curl -fsSL https://raw.githubusercontent.com/Waariss/jailbreakit/main/install.sh | sh
```

Install a pinned release:

```sh
curl -fsSL https://raw.githubusercontent.com/Waariss/jailbreakit/main/install.sh | JAILBREAKIT_VERSION=v1.3.2 sh
```

## Safety Posture

This release only changes distribution and installation ergonomics. `jailbreakit` remains an authorized iOS security testing and lab setup helper. It does not create a new jailbreak, exploit third-party devices, bypass app DRM, distribute decrypted IPAs, store credentials, or provide jailbreak exploit code.

## Test Commands Run

```sh
gofmt -w cmd internal
sh -n install.sh
go test ./...
go build -o /private/tmp/jailbreakit-v1.3.2 ./cmd/jailbreakit
```

## Known Limitations

- Homebrew support is not included in this release; a clean Homebrew tap can be added later if a separate tap repository is acceptable.
- `apt` repository support is not included in this release.
- macOS release binaries are still not notarized.

## Upgrade Notes

Tag and release as:

```sh
git tag v1.3.2
git push origin v1.3.2
```

The auto-release workflow can also create the missing `v1.3.2` tag and release from `internal/version/version.go` when changes are pushed to `main`.
