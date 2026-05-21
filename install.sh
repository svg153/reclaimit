#!/usr/bin/env bash
# install.sh - simple installer used by webinstall/web to fetch latest release
set -euo pipefail
OWNER="svg153"
REPO="reclaimit"
TAG=${1:-latest}
API_URL="https://api.github.com/repos/${OWNER}/${REPO}/releases"
if [ "$TAG" = "latest" ]; then
  release_url="$API_URL/latest"
else
  release_url="$API_URL/tags/$TAG"
fi
asset_url=$(curl -s "$release_url" | jq -r '.assets[] | select(.name | test("reclaimit_")) | .browser_download_url' | head -n1)
if [ -z "$asset_url" ] || [ "$asset_url" = "null" ]; then
  echo "No release asset found. Please install from source or Homebrew."
  exit 1
fi
curl -L "$asset_url" -o /tmp/reclaimit.tar.gz
mkdir -p "$HOME/.local/bin"
rm -f /tmp/reclaimit && tar -xzf /tmp/reclaimit.tar.gz -C /tmp
mv /tmp/reclaimit "$HOME/.local/bin/reclaimit"
chmod +x "$HOME/.local/bin/reclaimit"
echo "reclaimit installed to $HOME/.local/bin/reclaimit"
