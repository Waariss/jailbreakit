# Release Notes: v1.3.0

## Summary

`jailbreakit` v1.3.0 is the initial public lab-readiness release for authorized iOS pentest environment preparation and MASTG-style dynamic analysis workflows.

## New Commands

- `jailbreakit lab-check` checks host dependencies, connected device metadata, IPA install readiness, Frida/Objection readiness, and optional SSH readiness.
- `jailbreakit frida-check` checks host-side Frida and Objection tooling and prints safe next steps.
- `jailbreakit evidence --format markdown|json` generates a lab readiness evidence report for pentest notes.

## Safety Posture

This release keeps the project scoped to authorized iOS security testing and lab setup. It does not create a new jailbreak, exploit third-party devices, bypass app DRM, distribute decrypted IPAs, store Apple ID credentials, or download Frida server binaries.

## Test Commands Run

```sh
gofmt -w cmd internal
go test ./...
go build -o /private/tmp/jailbreakit-v1.3.0 ./cmd/jailbreakit
```

Smoke checks:

```sh
./jailbreakit help
./jailbreakit lab-check
./jailbreakit frida-check
./jailbreakit evidence --format markdown
./jailbreakit evidence --format json
```

## Known Limitations

- `frida-check` does not install or download `frida-server`.
- SSH checks are skipped unless explicit SSH flags are provided.
- SSH checks use `BatchMode=yes` and do not prompt for or guess credentials.
- Device-side installer checks for `appinst` and `ipainstaller` are informational unless available in the active path.
- Evidence reports may include local tool paths and should be reviewed before sharing.

## Upgrade Notes

Existing commands remain available. The local version now reports `v1.3.0`.

For release:

```sh
git tag v1.3.0
git push origin v1.3.0
```

Release automation is now tag-driven through `.github/workflows/release.yml`; pushes to `main` do not create a public release.
