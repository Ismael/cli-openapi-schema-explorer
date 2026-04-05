#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_DIR"

# Get version from argument or prompt
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  CURRENT=$(grep -o '"version": *"[^"]*"' .claude-plugin/plugin.json | head -1 | sed 's/.*"\([^"]*\)"/\1/')
  read -rp "Version (current: ${CURRENT}): " VERSION
fi

# Strip leading 'v' if provided
VERSION="${VERSION#v}"

if [ -z "$VERSION" ]; then
  echo "Error: version required" >&2
  exit 1
fi

TAG="v${VERSION}"

# Check for uncommitted changes
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "Error: uncommitted changes. Commit or stash first." >&2
  exit 1
fi

# Check tag doesn't already exist
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "Error: tag ${TAG} already exists" >&2
  exit 1
fi

# Update version in .claude-plugin files
sed -i "s/\"version\": *\"[^\"]*\"/\"version\": \"${VERSION}\"/g" \
  .claude-plugin/plugin.json \
  .claude-plugin/marketplace.json

# Commit and tag
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json
git commit -m "release: v${VERSION}"
git tag "$TAG"

echo ""
echo "Created commit and tag ${TAG}"
echo ""
read -rp "Push to origin? [y/N] " PUSH
if [[ "$PUSH" =~ ^[Yy]$ ]]; then
  git push origin main
  git push origin "$TAG"
  echo "Pushed. GitHub Actions will build and create the release."
else
  echo "Run manually:"
  echo "  git push origin main && git push origin ${TAG}"
fi
