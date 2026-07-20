# Release Notes: v1.3.1

## Summary

`jailbreakit` v1.3.1 is a patch release for the first public lab-readiness line. It adds Burp CA profile generation/installation support and keeps the project scoped to authorized iOS pentest lab preparation.

## Changes

- Added `jailbreakit burp-ca --cert <path>` to generate an iOS Burp CA `.mobileconfig` profile from a local certificate.
- Added `burp-ca --install` installer detection for `ideviceprofile`, `pymobiledevice3`, and macOS `cfgutil`.
- Improved the missing-installer message with `pymobiledevice3`, `cfgutil`, and manual profile transfer hints.
- Kept generated `.mobileconfig` files usable even when no USB profile installer is installed.
- Bumped the local version to `v1.3.1`.

## Safety Posture

This release does not auto-enable iOS certificate trust. After installing the Burp CA profile, the user must still manually enable full trust in iOS Settings.

`jailbreakit` remains an authorized iOS security testing and lab setup helper. It does not create a new jailbreak, exploit third-party devices, bypass app DRM, distribute decrypted IPAs, store credentials, or provide jailbreak exploit code.

## Test Commands Run

```sh
gofmt -w cmd internal
go test ./...
go build -o /private/tmp/jailbreakit-v1.3.1 ./cmd/jailbreakit
```

## Known Limitations

- USB profile installation requires one supported host tool in `PATH`: `pymobiledevice3`, `ideviceprofile`, or `cfgutil` on macOS.
- iOS full certificate trust must be enabled manually by the device owner.
- The Burp CA certificate must come from the tester's own Burp instance; `jailbreakit` does not download it.

## Upgrade Notes

Tag and release as:

```sh
git tag v1.3.1
git push origin v1.3.1
```

Release automation is now tag-driven through `.github/workflows/release.yml`; pushes to `main` do not create a public release.
