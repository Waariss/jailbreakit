# jailbreakit

[![ci](https://github.com/Waariss/jailbreakit/actions/workflows/ci.yml/badge.svg)](https://github.com/Waariss/jailbreakit/actions/workflows/ci.yml)
[![release](https://github.com/Waariss/jailbreakit/actions/workflows/release.yml/badge.svg)](https://github.com/Waariss/jailbreakit/actions/workflows/release.yml)
[![version](https://img.shields.io/github/v/release/Waariss/jailbreakit?label=version)](https://github.com/Waariss/jailbreakit/releases)

<div align="center">
  <img src="media/jailbreakit.png" width="45%" alt="jailbreakit logo">
</div>

`jailbreakit` is a CLI helper for authorized iOS pentesting workflows.

It detects a connected iPhone, checks jailbreak compatibility, recommends a route, and guides the user through `palera1n` or Dopamine setup. The goal is not to create a new jailbreak. The goal is to make the repetitive setup work easier for pentesters and security researchers working on devices they are allowed to test.

## Why

iOS jailbreak setup for pentesting is repetitive and easy to get wrong:

- identify the device model, chip, and iOS version
- check which jailbreak supports that version
- choose between `palera1n` and Dopamine
- run the correct first-time/rootful/rootless commands
- download the matching Dopamine IPA instead of the wrong latest release
- install a CLI signer and handle Apple ID login/2FA
- remember the iPhone trust-profile step after sideloading

`jailbreakit` turns that into one guided flow:

```sh
jailbreakit
```

## Platform Support

Current target platforms:

- macOS
- Linux

Windows is not supported yet. The project is currently iPhone-first; iPad, iPod, Apple TV, and T2 metadata exist as initial compatibility data, but the main tested workflow is iPhone jailbreak setup for authorized security testing.

## Install

Clone the repository:

```sh
git clone https://github.com/Waariss/jailbreakit.git
cd jailbreakit
```

Build the binary:

```sh
go build -o jailbreakit ./cmd/jailbreakit
```

Run it:

```sh
./jailbreakit
```

Install into your Go bin path:

```sh
go install ./cmd/jailbreakit
jailbreakit
```

### macOS Release Binary

Release binaries are not notarized yet. If macOS Gatekeeper blocks a downloaded binary with a message like "Apple could not verify ... is free of malware", remove the quarantine attribute and make it executable:

```sh
chmod +x jailbreakit-darwin-arm64
xattr -d com.apple.quarantine jailbreakit-darwin-arm64
./jailbreakit-darwin-arm64
```

For Intel Macs, replace `jailbreakit-darwin-arm64` with `jailbreakit-darwin-amd64`.

Long term, signed and notarized macOS releases are planned.

## Requirements

Core tools:

- Go 1.24+ to build from source
- `palera1n` for checkm8/palera1n flows
- `libimobiledevice` for `ideviceinfo` device detection
- `curl` or network access for downloads

macOS:

```sh
brew install libimobiledevice curl
```

Linux package names vary by distribution. Debian/Ubuntu-style systems usually need:

```sh
sudo apt install -y libimobiledevice-utils curl
```

Check your machine:

```sh
./jailbreakit doctor
```

Install package-manager dependencies where supported:

```sh
./jailbreakit doctor --install
```

`palera1n` should be installed from the official project instructions. `jailbreakit` does not guess unofficial package sources for it.

## Usage

The normal user flow is intentionally short:

```sh
./jailbreakit
```

Common utility commands:

```sh
./jailbreakit doctor
./jailbreakit detect
./jailbreakit recommend --ios 15.8.8 --product iPhone8,1
./jailbreakit version
```

Advanced commands are hidden from the default help:

```sh
./jailbreakit help advanced
```

## Development

Run tests:

```sh
go test ./...
```

Check formatting:

```sh
gofmt -w cmd internal
```

GitHub Actions runs `gofmt` and `go test ./...` on pushes and pull requests. Tagged releases build macOS and Linux binaries.

## What It Does

Guided mode:

1. Detects the connected device.
2. Maps `ProductType` to model and chip.
3. Checks the iOS jailbreak matrix from The Apple Wiki, with versioned embedded fallback data when the site is unavailable.
4. Shows compatible jailbreak routes.
5. Runs `palera1n`, or downloads and sideloads the matching Dopamine IPA.
6. Prints the next iPhone-side steps, including developer-profile trust instructions.

Example recommendation:

```text
ProductType:  iPhone8,1
Model:        iPhone 6s
Chip:         A9
iOS:          15.8.8

Options:
[1] palera1n 2.2.1 - rootless or rootful fakefs, semi-tethered
[2] Dopamine 2.5 Beta 3 - rootless, semi-untethered
```

For Dopamine `2.5 Beta 3`, `jailbreakit` resolves the release tag to `2.5b3` and downloads:

```text
https://github.com/opa334/Dopamine/releases/download/2.5b3/Dopamine.ipa
```

Routes outside the currently automated runners are still shown as recommendations when compatibility data is available. For example, iOS 12/13/14 may show tools such as Chimera, unc0ver, checkra1n, Odyssey, or Taurine as `recommend-only`. Those routes require the upstream tool or guide until runner automation is implemented.

## Dopamine Sideloading

`jailbreakit` is CLI-first. If no signer is available, guided mode can download the `plumesign` CLI, mark it executable, and continue.

Manual signer install:

```sh
./jailbreakit signer install --platform macos
./jailbreakit signer install --platform linux-aarch64
./jailbreakit signer install --platform linux-x86_64
```

The signer is saved to:

```text
./bin/plumesign
```

Apple ID handling:

- `jailbreakit` does not store Apple ID credentials.
- Signing and authentication are delegated to the selected local signer.
- If credentials are required, the local signer handles password and 2FA prompts interactively.
- The selected signer may maintain its own local authentication or session state.
- Users are encouraged to use an app-specific password or a dedicated lab Apple ID.

After Dopamine is installed, trust the developer profile on the iPhone:

```text
Settings > General > VPN & Device Management
```

Then open Dopamine and tap Jailbreak.

## palera1n Notes

Rootless flow:

```sh
./jailbreakit run palera1n --rootless
```

Rootful first-time BindFS creation:

```sh
./jailbreakit run palera1n --rootful
```

After the device returns to recovery mode, boot the existing rootful BindFS:

```sh
./jailbreakit run palera1n --rootful-boot
```

Guided mode handles this flow and tells the user when the second step is needed.

## Privacy

`jailbreakit` is local-first.

- No telemetry
- No remote command execution
- No Apple ID credential storage
- No device identifiers are uploaded by default
- Downloads are performed only from configured upstream project URLs

## Security & Third-Party Notices

This repository does not bundle third-party jailbreak binaries by default. It only references upstream tools, local installations, and upstream release downloads.

Third-party projects referenced or orchestrated by this tool have their own licenses, terms, and safety guidance:

- `palera1n` — follow the official project documentation and license
- `Dopamine` — follow the upstream release notes, license, and distribution terms
- `plumesign` — follow the upstream project terms and license
- The Apple Wiki — follow the site’s terms and attribution guidance

When available, verify downloaded artifacts against upstream checksums or signatures before use.

## Data

Compatibility and device metadata are versioned with the binary:

- `internal/matrix/compatibility.json` contains embedded fallback jailbreak data.
- `internal/device/product-map.json` contains embedded product metadata for iPhone-first detection, with initial iPad, iPod, Apple TV, and T2 coverage.

The Apple Wiki remains the preferred live source when reachable.

## Disclaimer

`jailbreakit` is intended strictly for authorized assessments, iOS pentesting, and responsible security research. Use it only on devices and environments where you have explicit permission.

Do not use this tool on devices you do not own or do not have explicit written permission to assess.

This project is a helper/orchestrator. It is not affiliated with Apple, `palera1n`, Dopamine, Impactor, or The Apple Wiki.

Jailbreaking can cause data loss, boot issues, restore requirements, or device instability. We are not responsible for any data loss, device damage, bricked devices, account issues, failed jailbreak attempts, or any other outcome caused by using this tool or the underlying third-party tools. When using `palera1n`, Dopamine, `plumesign`, or any related tooling, the user accepts full responsibility for anything that happens to their device during the process.

This tool does not bypass iCloud, Activation Lock, MDM, passcodes, DRM/FairPlay, or device ownership protections.

## Credits

This project orchestrates and references work from:

- [claration/Impactor](https://github.com/claration/Impactor)
- [opa334/Dopamine](https://github.com/opa334/Dopamine)
- [palera1n/palera1n](https://github.com/palera1n/palera1n)
- [The Apple Wiki](https://theapplewiki.com/wiki/Main_Page)

Respect the licenses, documentation, and safety guidance of the upstream projects.

## Status

Early MVP. Expect changes to commands, compatibility data, and sideloading flow as the project matures.
