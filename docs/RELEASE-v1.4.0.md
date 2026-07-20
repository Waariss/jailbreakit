# jailbreakit v1.4.0

## Summary

This release adds device-aware Frida readiness checks, local Burp CA verification, and read-only IPA metadata inspection for authorized iOS security testing workflows.

## New commands

```sh
jailbreakit lab-check --device
jailbreakit frida-check --device
jailbreakit burp-ca verify --cert cacert.der --profile burp-ca.mobileconfig
jailbreakit install ./App.ipa --inspect
```

`frida-check --device` runs `frida-ps -U` with a short timeout. It does not install or download `frida-server`.

`burp-ca verify` validates the local certificate and checks whether a local mobileconfig contains the same certificate. It cannot prove that iOS installed the profile or enabled Full Trust.

`install --inspect` reads IPA metadata, including bundle identifier, versions, minimum iOS, executable architecture, provisioning profile presence, and available entitlement keys. It does not modify, decrypt, or install the IPA.

`lab-check --ssh-interactive` can use an SSH password prompt for the harmless readiness command. The password is handled by the system `ssh` process and is not read or stored by `jailbreakit`.

## Safety posture

`jailbreakit` remains an authorized iOS pentest lab setup helper. It does not provide jailbreak exploit code, bypass app DRM, distribute decrypted IPAs, store credentials, or bypass iOS certificate trust.

## Validation

```sh
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./cmd/jailbreakit
```

## Upgrade notes

The existing `frida-check`, `burp-ca`, and `install` commands keep their previous behavior. The new flags are opt-in. This release does not require a connected device for host-side checks or IPA inspection.
