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

The prepared Formula is intended for the future `Waariss/homebrew-tap` repository:

```sh
brew install waariss/tap/jailbreakit
```

This is not publicly supported until that tap exists and the Formula has passed maintainer testing.

## Not currently supported

- Windows
- Homebrew Core
- apt repositories
- RPM repositories
- macOS notarization
- Homebrew bottles
