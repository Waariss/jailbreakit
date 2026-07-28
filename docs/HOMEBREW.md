# Homebrew Tap Preparation

The canonical Formula template is `packaging/homebrew/Formula/jailbreakit.rb`.
The public tap is available at [Waariss/homebrew-tap](https://github.com/Waariss/homebrew-tap). Users can install the current Formula with `brew install waariss/tap/jailbreakit`.

## Tap maintenance

The tap repository uses this structure:

```text
homebrew-tap/
├── Formula/
│   └── jailbreakit.rb
└── README.md
```

For a new release, copy the tested Formula from `jailbreakit/packaging/homebrew/Formula/jailbreakit.rb` to `homebrew-tap/Formula/jailbreakit.rb`, then commit and push the Formula update to the tap repository.

For a released tag, update the template from this repository:

```sh
./scripts/update-homebrew-formula.sh v1.4.2
```

The helper downloads the tagged source archive, calculates SHA256 using `shasum` or `sha256sum`, and changes only the Formula URL and checksum.

## User installation

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

To test a Formula update locally before pushing it to the tap:

```sh
brew install --build-from-source ./packaging/homebrew/Formula/jailbreakit.rb
brew test ./packaging/homebrew/Formula/jailbreakit.rb
brew audit --strict ./packaging/homebrew/Formula/jailbreakit.rb
```

Homebrew does not install optional iOS testing tools as Formula dependencies. Use `jailbreakit doctor` to discover `ideviceinfo`, `iproxy`, `ideviceinstaller`, Frida, Objection, and `palera1n`.
