#!/bin/sh
set -eu

formula_path="$(dirname "$0")/../packaging/homebrew/Formula/jailbreakit.rb"
version="${1:-}"

if ! printf '%s\n' "$version" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$'; then
  echo "usage: $0 vX.Y.Z" >&2
  exit 2
fi

archive_url="https://github.com/Waariss/jailbreakit/archive/refs/tags/${version}.tar.gz"
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT HUP INT TERM
archive_path="$temp_dir/jailbreakit.tar.gz"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$archive_url" -o "$archive_path"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$archive_url" -O "$archive_path"
else
  echo "curl or wget is required" >&2
  exit 1
fi

if command -v shasum >/dev/null 2>&1; then
  sha256="$(shasum -a 256 "$archive_path" | awk '{print $1}')"
elif command -v sha256sum >/dev/null 2>&1; then
  sha256="$(sha256sum "$archive_path" | awk '{print $1}')"
else
  echo "shasum or sha256sum is required" >&2
  exit 1
fi

temp_formula="$temp_dir/jailbreakit.rb"
awk -v url="$archive_url" -v sha="$sha256" '
  /^  url / { print "  url \"" url "\""; next }
  /^  sha256 / { print "  sha256 \"" sha "\""; next }
  { print }
' "$formula_path" > "$temp_formula"
mv "$temp_formula" "$formula_path"
echo "updated $formula_path for $version"
