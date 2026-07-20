# Homebrew Tap Preparation

The canonical Formula template is `packaging/homebrew/Formula/jailbreakit.rb`.
The tap is not published yet. Do not advertise `brew install waariss/tap/jailbreakit` as available until the external repository exists and the Formula has been tested there.

## First-time tap setup

Create a public GitHub repository named `Waariss/homebrew-tap`:

```text
homebrew-tap/
├── Formula/
│   └── jailbreakit.rb
└── README.md
```

Copy `jailbreakit/packaging/homebrew/Formula/jailbreakit.rb` to `homebrew-tap/Formula/jailbreakit.rb`, then replace its source archive checksum with the verified value.

For a released tag, update the template from this repository:

```sh
./scripts/update-homebrew-formula.sh v1.3.2
```

The helper downloads the tagged source archive, calculates SHA256 using `shasum` or `sha256sum`, and changes only the Formula URL and checksum.

## User installation

After the tap is public and tested:

```sh
brew install waariss/tap/jailbreakit
```

or:

```sh
brew tap waariss/tap
brew install jailbreakit
```

## Local maintainer testing

With the tap repository available:

```sh
brew uninstall jailbreakit 2>/dev/null || true
brew untap waariss/tap 2>/dev/null || true
brew tap waariss/tap
brew install --build-from-source waariss/tap/jailbreakit
jailbreakit version
jailbreakit doctor
brew test waariss/tap/jailbreakit
brew audit --strict waariss/tap/jailbreakit
```

Before the tap exists, test the Formula file directly:

```sh
brew install --build-from-source ./packaging/homebrew/Formula/jailbreakit.rb
brew test ./packaging/homebrew/Formula/jailbreakit.rb
brew audit --strict ./packaging/homebrew/Formula/jailbreakit.rb
```

Homebrew does not install optional iOS testing tools as Formula dependencies. Use `jailbreakit doctor` to discover `ideviceinfo`, `iproxy`, `ideviceinstaller`, Frida, Objection, and `palera1n`.
