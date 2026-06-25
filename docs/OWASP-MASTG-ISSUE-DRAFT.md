# Proposal: Add jailbreakit as an iOS pentest lab readiness helper

## Description

I would like to propose adding `jailbreakit` as an iOS pentest lab readiness helper for MASTG-style testing workflows.

Repository: https://github.com/Waariss/jailbreakit

Disclosure: I am the author/maintainer of this tool.

## Why It Is Useful for MASTG-Style Testing

Mobile application security testing often requires repeatable lab setup before runtime analysis can begin. `jailbreakit` focuses on checking whether the host and authorized iOS device are ready for dynamic analysis workflows without positioning itself as an exploitation framework.

It helps testers verify local dependencies, connected-device metadata, SSH/iproxy readiness, IPA installation readiness, and Frida/Objection readiness. It can also generate a sanitized readiness evidence report for working notes.

## Supported Workflows

- Host dependency checks with `jailbreakit doctor`.
- Connected device identification with `jailbreakit detect`.
- Jailbreak/testing route guidance with `jailbreakit recommend`.
- Lab readiness checks with `jailbreakit lab-check`.
- Frida and Objection readiness checks with `jailbreakit frida-check`.
- Evidence report generation with `jailbreakit evidence`.
- Authorized local IPA installation with `jailbreakit install`.

## Safety Boundaries

`jailbreakit` is intended only for authorized iOS security testing and lab setup. It does not:

- Create a new jailbreak.
- Exploit third-party devices.
- Bypass app DRM or FairPlay.
- Provide decrypted IPAs.
- Store Apple ID credentials.
- Download or install Frida server binaries automatically.
- Attempt iCloud, Activation Lock, MDM, or passcode bypass.

## Maintenance Status

The project is maintained as a Go CLI with unit tests and GitHub Actions for formatting, tests, and release builds. v1.3.0 adds lab readiness checks and evidence generation.

## Commands and Examples

```sh
jailbreakit lab-check
jailbreakit lab-check --ssh-host 127.0.0.1 --ssh-port 2222 --ssh-user root
jailbreakit frida-check
jailbreakit evidence --format markdown
jailbreakit evidence --format json --out lab-evidence.json
jailbreakit install ./App.ipa
```

Example outputs are available in the repository under `examples/`.

## Question for Maintainers

Would maintainers prefer this to start as a tool listing, a technique reference, or a small demo contribution showing how lab readiness evidence can support MASTG-style iOS dynamic analysis workflows?
