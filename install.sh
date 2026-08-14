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
INSTALL_DIR=${RECLAIMIT_INSTALL_DIR:-"${HOME}/.local/bin"}

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
release_json=$(curl -fsSL "$release_url")
release_assets=$(printf '%s' "$release_json" | jq -r --arg pattern "$asset_pattern" '
  (first(.assets[] | select(.name | test($pattern))) // {}) as $archive |
  (first(.assets[] | select(.name | test("^reclaimit_.*_checksums\\.txt$"))) // {}) as $checksums |
  [$archive.name // "", $archive.browser_download_url // "", $checksums.browser_download_url // ""] |
  @tsv
')
IFS=$'\t' read -r asset_name asset_url checksum_url <<< "$release_assets"
if [ -z "$asset_name" ] || [ -z "$asset_url" ]; then
  echo "No release asset found for ${os}/${arch}. Please install from source or the release page." >&2
  exit 1
fi
if [ -z "$checksum_url" ]; then
  echo "No checksum manifest found for release asset $asset_name; refusing to install." >&2
  exit 1
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT
archive_path="$tmp_dir/reclaimit.archive"
checksum_path="$tmp_dir/reclaimit_checksums.txt"

curl -fsSL "$asset_url" -o "$archive_path"
curl -fsSL "$checksum_url" -o "$checksum_path"

expected_checksum=$(awk -v asset="$asset_name" '
  {
    name = $2
    sub(/^\*/, "", name)
    if (name == asset) {
      print tolower($1)
      exit
    }
  }
' "$checksum_path")
if [[ ! "$expected_checksum" =~ ^[0-9a-f]{64}$ ]]; then
  echo "Checksum manifest has no valid SHA-256 entry for $asset_name; refusing to install." >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual_checksum=$(sha256sum "$archive_path" | awk '{print tolower($1)}')
elif command -v shasum >/dev/null 2>&1; then
  actual_checksum=$(shasum -a 256 "$archive_path" | awk '{print tolower($1)}')
else
  echo "Error: sha256sum or shasum is required to verify release integrity." >&2
  exit 1
fi

if [ "$actual_checksum" != "$expected_checksum" ]; then
  echo "Checksum mismatch for $asset_name; refusing to install." >&2
  exit 1
fi
echo "Verified SHA-256 checksum for $asset_name"

case "$asset_url" in
  *.zip)
    need_cmd unzip
    unzip -q "$archive_path" -d "$tmp_dir"
    ;;
  *.tar.gz)
    tar --no-same-owner -xzf "$archive_path" -C "$tmp_dir"
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

target_name="reclaimit"
if [ "$os" = "windows" ]; then
  target_name="reclaimit.exe"
fi
mkdir -p "$INSTALL_DIR"
install -m 0755 "$bin_path" "$INSTALL_DIR/$target_name"
echo "reclaimit installed to $INSTALL_DIR/$target_name"
