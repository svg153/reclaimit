#!/usr/bin/env bash
# install.sh - simple installer used by webinstall/web to fetch latest release
set -euo pipefail

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Error: required command '$1' was not found." >&2
    exit 1
  fi
}

need_cmd curl
need_cmd jq

OWNER="svg153"
REPO="reclaimit"
TAG=${1:-latest}
API_URL="https://api.github.com/repos/${OWNER}/${REPO}/releases"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$os" in
  linux) os="linux" ;;
  darwin) os="darwin" ;;
  msys*|mingw*|cygwin*) os="windows" ;;
  *)
    echo "Unsupported operating system: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$TAG" = "latest" ]; then
  release_url="$API_URL/latest"
else
  release_url="$API_URL/tags/$TAG"
fi

asset_pattern="reclaimit_.*_${os}_${arch}\\.(tar\\.gz|zip)$"
asset_url=$(curl -fsSL "$release_url" | jq -r --arg pattern "$asset_pattern" '.assets[] | select(.name | test($pattern)) | .browser_download_url' | head -n1)
if [ -z "$asset_url" ] || [ "$asset_url" = "null" ]; then
  echo "No release asset found for ${os}/${arch}. Please install from source or Homebrew." >&2
  exit 1
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT
archive_path="$tmp_dir/reclaimit.archive"

curl -fsSL "$asset_url" -o "$archive_path"
mkdir -p "$HOME/.local/bin"

case "$asset_url" in
  *.zip)
    need_cmd unzip
    unzip -q "$archive_path" -d "$tmp_dir"
    ;;
  *.tar.gz)
    tar -xzf "$archive_path" -C "$tmp_dir"
    ;;
  *)
    echo "Unsupported archive format: $asset_url" >&2
    exit 1
    ;;
esac

bin_path=$(find "$tmp_dir" -maxdepth 2 -type f \( -name reclaimit -o -name reclaimit.exe \) | head -n1)
if [ -z "$bin_path" ]; then
  echo "Unable to find reclaimit binary inside downloaded archive." >&2
  exit 1
fi

mv "$bin_path" "$HOME/.local/bin/reclaimit"
chmod +x "$HOME/.local/bin/reclaimit"
echo "reclaimit installed to $HOME/.local/bin/reclaimit"
