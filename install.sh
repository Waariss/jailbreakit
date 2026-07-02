#!/bin/sh
set -eu

REPO="${JAILBREAKIT_REPO:-Waariss/jailbreakit}"
VERSION="${JAILBREAKIT_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BIN_NAME="${BIN_NAME:-jailbreakit}"

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

detect_os() {
  case "$(uname -s)" in
    Darwin) printf 'darwin' ;;
    Linux) printf 'linux' ;;
    *) fail "unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    arm64|aarch64) printf 'arm64' ;;
    x86_64|amd64) printf 'amd64' ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac
}

download() {
  url="$1"
  out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$out"
    return
  fi
  if command -v wget >/dev/null 2>&1; then
    wget -q "$url" -O "$out"
    return
  fi
  fail "missing curl or wget"
}

sha256_file() {
  file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
    return
  fi
  fail "missing sha256sum or shasum"
}

install_binary() {
  src="$1"
  dst_dir="$2"
  dst="$dst_dir/$BIN_NAME"

  if mkdir -p "$dst_dir" 2>/dev/null && cp "$src" "$dst" 2>/dev/null; then
    chmod 0755 "$dst"
    return
  fi

  if command -v sudo >/dev/null 2>&1; then
    sudo mkdir -p "$dst_dir"
    sudo install -m 0755 "$src" "$dst"
    return
  fi

  fail "cannot write to $dst_dir and sudo is unavailable; set INSTALL_DIR to a writable directory"
}

need uname
need awk

os="$(detect_os)"
arch="$(detect_arch)"
asset="jailbreakit-${os}-${arch}"

if [ "$VERSION" = "latest" ]; then
  base_url="https://github.com/${REPO}/releases/latest/download"
else
  base_url="https://github.com/${REPO}/releases/download/${VERSION}"
fi

tmp_dir="${TMPDIR:-/tmp}/jailbreakit-install.$$"
mkdir -p "$tmp_dir"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

binary_path="$tmp_dir/$asset"
checksums_path="$tmp_dir/SHA256SUMS"

log "[*] downloading $asset from $REPO ($VERSION)"
download "$base_url/$asset" "$binary_path"
chmod 0755 "$binary_path"

if download "$base_url/SHA256SUMS" "$checksums_path" 2>/dev/null; then
  expected="$(awk -v asset="$asset" '$2 == asset {print $1}' "$checksums_path")"
  if [ -n "$expected" ]; then
    actual="$(sha256_file "$binary_path")"
    [ "$actual" = "$expected" ] || fail "checksum mismatch for $asset"
    log "[+] checksum verified"
  else
    log "[!] checksum entry not found for $asset; skipping verification"
  fi
else
  log "[!] SHA256SUMS not available; skipping verification"
fi

install_binary "$binary_path" "$INSTALL_DIR"
log "[+] installed $BIN_NAME to $INSTALL_DIR/$BIN_NAME"
log "[*] run: $BIN_NAME version"
