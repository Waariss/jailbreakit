# Distribution

Supported installation channels:

## Go module

```sh
go install github.com/Waariss/jailbreakit/cmd/jailbreakit@latest
```

## GitHub Releases

Tagged releases provide macOS and Linux binaries for amd64 and arm64, plus `SHA256SUMS`. Raw binary names remain compatible with `install.sh`.

## Installer

```sh
curl -fsSL https://raw.githubusercontent.com/Waariss/jailbreakit/main/install.sh | sh
```

The installer verifies a release checksum when `SHA256SUMS` is available.

## Homebrew

The public Formula is available from `Waariss/homebrew-tap`:

```sh
brew install waariss/tap/jailbreakit
```

The tap is maintained separately from this repository and is updated after tagged releases.

## Not currently supported

- Windows
- Homebrew Core
- apt repositories
- RPM repositories
- macOS notarization
- Homebrew bottles
