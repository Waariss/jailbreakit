# MASTG-Style Positioning

## Purpose

`jailbreakit` is an iOS pentest lab readiness helper for authorized mobile application security testing. It helps testers prepare and document a host/device lab before dynamic analysis.

## What jailbreakit Is

`jailbreakit` is a local CLI that helps with:

- Host dependency checks through `jailbreakit doctor`.
- Connected device identification through `jailbreakit detect`.
- Route guidance through `jailbreakit recommend`.
- Lab readiness checks through `jailbreakit lab-check`.
- Frida and Objection readiness checks through `jailbreakit frida-check`.
- Evidence report generation through `jailbreakit evidence`.
- Burp CA profile generation and installation through `jailbreakit burp-ca`.
- Authorized local IPA installation through `jailbreakit install`.

## What jailbreakit Is Not

`jailbreakit` is not:

- An exploitation framework.
- An unauthorized jailbreak automation tool.
- A DRM bypass tool.
- A decrypted IPA distribution tool.
- A credential collection or storage mechanism.
- An official OWASP or MASTG tool.

## MASTG-Style Use Cases

`jailbreakit` is designed to support mobile security testing workflows such as:

- Preparing an iOS device for dynamic analysis.
- Checking libimobiledevice, iproxy, SSH, scp, ideviceinstaller, Frida, and Objection readiness.
- Validating whether SSH over USB can be checked safely after the tester starts `iproxy`.
- Preparing for runtime testing with Frida and Objection.
- Generating and installing a Burp CA profile while keeping full trust as an explicit user action.
- Installing authorized lab IPA files.
- Generating repeatable readiness evidence for working notes.
- Supporting jailbreak-detection validation workflows on authorized lab devices.

## Example Workflows

Check baseline host and device readiness:

```sh
jailbreakit lab-check
```

Check SSH after starting a local USB tunnel:

```sh
iproxy 2222 22
jailbreakit lab-check --ssh-host 127.0.0.1 --ssh-port 2222 --ssh-user root
```

Check Frida and Objection readiness:

```sh
jailbreakit frida-check
```

Generate evidence for a pentest note:

```sh
jailbreakit evidence --format markdown
jailbreakit evidence --format json --out lab-evidence.json
```

Generate and optionally install a Burp CA profile:

```sh
jailbreakit burp-ca --cert cacert.der --out burp-ca.mobileconfig
jailbreakit burp-ca --cert cacert.der --install
```

Install an authorized lab IPA:

```sh
jailbreakit install ./App.ipa
```

## Safety Boundaries

`jailbreakit` does not create a new jailbreak, distribute jailbreak binaries, bypass app DRM, provide decrypted IPAs, store Apple ID credentials, fetch Frida server binaries automatically, or auto-enable iOS certificate trust.

Users must only test devices and applications they are authorized to assess.

## Why This Is a Lab Readiness Helper

The tool focuses on repeatable setup checks, local orchestration, and evidence collection. It reports missing dependencies, suggests safe next steps, and avoids automatic exploit delivery or credential guessing. This makes it suitable as a testing/lab readiness helper rather than an exploitation tool.
