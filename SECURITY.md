# Security Policy

## Authorized-Use Scope

`jailbreakit` is intended for authorized iOS security testing, iOS pentest lab setup, and dynamic analysis readiness checks only. Use it only on devices, applications, and lab environments that you own or are explicitly authorized to assess.

The project helps operators verify host tooling, connected-device metadata, SSH/iproxy readiness, IPA installation readiness, and Frida/Objection readiness. It does not create or distribute jailbreaks.

## Responsible Disclosure

If you believe you found a security issue in `jailbreakit`, please open a GitHub Issue with enough detail to reproduce the behavior. If public disclosure would expose sensitive information, open a minimal issue first and request a private coordination channel.

Contact placeholder: security contact to be added by the maintainer.

## Non-Goals

`jailbreakit` does not support or accept requests for:

- Unauthorized device access.
- App DRM or FairPlay bypass.
- Decrypted IPA distribution or handling.
- Apple ID credential storage.
- Exploit development assistance.
- iCloud, Activation Lock, MDM, or passcode bypass.
- Telemetry or remote command execution outside the user's explicit local workflow.

## Safe Usage

Run `jailbreakit` only in controlled lab environments. Generated evidence reports are intended for internal pentest notes and should be reviewed before sharing because local tool paths or device metadata may be included.
