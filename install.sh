#!/bin/sh
# hreysi installer — fetches the latest released binary from GitHub Releases.
#
#   curl -sSL https://raw.githubusercontent.com/Peleke/hreysi/main/install.sh | sh
#
# Override the install location with PREFIX (default /usr/local/bin).
set -e

REPO="Peleke/hreysi"
BIN="hreysi"
PREFIX="${PREFIX:-/usr/local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) echo "hreysi: unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux | darwin) ;;
  *) echo "hreysi: unsupported OS: $os" >&2; exit 1 ;;
esac

tag="$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep -m1 '"tag_name"' | cut -d'"' -f4)"
if [ -z "$tag" ]; then
  echo "hreysi: could not determine latest release" >&2
  exit 1
fi
ver="${tag#v}"

url="https://github.com/$REPO/releases/download/$tag/${BIN}_${ver}_${os}_${arch}.tar.gz"
echo "hreysi: downloading $url"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
curl -sSL "$url" | tar -xz -C "$tmp"

if [ -w "$PREFIX" ]; then
  mv "$tmp/$BIN" "$PREFIX/$BIN"
else
  echo "hreysi: $PREFIX not writable, using sudo"
  sudo mv "$tmp/$BIN" "$PREFIX/$BIN"
fi
chmod +x "$PREFIX/$BIN"

echo "hreysi: installed to $PREFIX/$BIN"
"$PREFIX/$BIN" version
