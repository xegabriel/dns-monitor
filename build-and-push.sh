#!/bin/bash

# Exit immediately if a command fails
set -e

# Function to get the next semantic version
get_next_version() {
    local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "0")

    if [[ "$latest_tag" == "0" ]]; then
        echo "1.0.0"
        return
    fi

    # Remove optional leading 'v'
    local tag_body=${latest_tag#v}

    if [[ "$tag_body" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        local major="${BASH_REMATCH[1]}"
        local minor="${BASH_REMATCH[2]}"
        local patch="${BASH_REMATCH[3]}"
        # Increment patch version
        echo "$major.$minor.$((patch + 1))"
    else
        # Fallback if format doesn't match strict semver X.Y.Z
        echo "⚠️ Warning: Latest tag '$latest_tag' is not in strict X.Y.Z format. Defaulting to 1.0.0." >&2 # Send warning to stderr
        echo "1.0.0"
    fi
}

# Generate a new version tag
GHCR_HOSTNAME="ghcr.io"
NEW_TAG=$(get_next_version)
IMAGE_NAME="dns-monitor"

# --- Validate Environment Variables ---
if [[ -z "${NEW_TAG:-}" ]]; then
    echo "❌ Error: NEW_TAG environment variable is not set."
    exit 1
fi
if [[ -z "${IMAGE_NAME:-}" ]]; then
    echo "❌ Error: IMAGE_NAME environment variable is not set."
    exit 1
fi
if [[ -z "${GITHUB_USER_OR_ORG:-}" ]]; then
    echo "❌ Error: GITHUB_USER_OR_ORG environment variable is not set."
    exit 1
fi
if [[ -z "${GITHUB_USER:-}" ]]; then
    echo "❌ Error: GITHUB_USER environment variable is not set (usually your GitHub username)."
    exit 1
fi
if [[ -z "${GITHUB_PAT:-}" ]]; then
    echo "❌ Error: GITHUB_PAT environment variable is not set (Personal Access Token with package scopes)."
    exit 1
fi

# Confirm before proceeding
echo "🚀 New version to be tagged: $IMAGE_NAME:$NEW_TAG"
read -p "Do you want to proceed? (y/N): " CONFIRM
if [[ "$CONFIRM" != "y" ]]; then
    echo "❌ Aborting."
    exit 1
fi

# Define the full image paths for GHCR
GHCR_IMAGE_BASE="${GHCR_HOSTNAME}/${GITHUB_USER_OR_ORG}/${IMAGE_NAME}"
GHCR_IMAGE_TAGGED="${GHCR_IMAGE_BASE}:${NEW_TAG}"
GHCR_IMAGE_LATEST="${GHCR_IMAGE_BASE}:latest"

# --- Git Tagging ---
echo "🏷️ Creating and pushing Git tag: $NEW_TAG"
# Check if tag exists locally first
if git rev-parse "$NEW_TAG" >/dev/null 2>&1; then
    echo "⚠️ Warning: Git tag '$NEW_TAG' already exists locally. Skipping tag creation."
else
    git tag -a "$NEW_TAG" -m "Release $NEW_TAG" || { echo "❌ Failed to create Git tag."; exit 1; }
fi
git push origin "$NEW_TAG" || { echo "❌ Failed to push Git tag '$NEW_TAG'. It might already exist remotely or another issue occurred."; exit 1; }

# --- Build and Push Multi-Platform Docker Image ---
echo "🐳 Building and pushing multi-platform image to GHCR..."
docker buildx build --platform linux/amd64,linux/arm64 \
  -t "$GHCR_IMAGE_TAGGED" \
  -t "$GHCR_IMAGE_LATEST" \
  . --push || { echo "❌ Failed to build and push multi-platform image."; exit 1; }

echo "✅ Docker images pushed successfully to GHCR!"
echo "📌 Image tags available at ${GHCR_HOSTNAME}/${GITHUB_USER_OR_ORG}/${IMAGE_NAME}:"
echo "   - ${NEW_TAG}"
echo "   - latest"

exit 0