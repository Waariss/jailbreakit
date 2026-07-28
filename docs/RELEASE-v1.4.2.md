# jailbreakit v1.4.2

## Summary

This release exposes the existing external-signer capability as a general local IPA sideload command for authorized iOS lab workflows.

## New command

```sh
jailbreakit sideload ./App.ipa
jailbreakit sideload ./App.ipa --login --apple-id tester@example.com
jailbreakit sideload ./App.ipa --dry-run
```

`sideload` detects a supported external signer or accepts a custom command template containing `{ipa}`. Apple ID passwords and 2FA responses remain handled by the external signer; `jailbreakit` does not read or store them.

Host-side `ApplicationVerificationFailed` errors now point users to either the explicit sideload workflow or the device-side AppSync and `appinst` route.

## Safety posture

`jailbreakit` remains an authorized iOS pentest lab setup helper. It does not provide jailbreak exploit code, bypass app DRM, distribute decrypted IPAs, store credentials, or bypass iOS trust controls.

Sideloading does not guarantee support for private entitlements used by specialized jailbreak applications. Those packages must follow their upstream-supported installation route.

## Validation

```sh
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
go build ./cmd/jailbreakit
```

## Upgrade notes

Existing `install`, lab readiness, Frida, Burp CA, evidence, and recommendation commands keep their previous behavior. The new `sideload` command is additive.
