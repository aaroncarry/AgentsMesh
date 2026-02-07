#!/bin/sh
#
# GitHub Release Script for AgentsMesh Runner
#
# This script creates a GitHub release and uploads artifacts.
# It's called by GitLab CI/CD pipeline.
#
# Required environment variables:
#   - CI_COMMIT_TAG: Git tag for the release (e.g., v0.3.1)
#   - GITHUB_RELEASE_TOKEN: GitHub PAT with repo access
#
# Expected artifact locations:
#   - runner/dist/*.tar.gz, *.zip - CLI artifacts from GoReleaser
#   - runner/dist/*.deb, *.rpm, *.apk - Linux packages
#   - runner/*.dmg, *.tar.gz, *.zip - Desktop artifacts (AgentsMesh* prefix)
#

set -e

GITHUB_REPO="AgentsMesh/AgentsMeshRunner"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}"

# Validate required environment variables
if [ -z "$CI_COMMIT_TAG" ]; then
  echo "[ERROR] CI_COMMIT_TAG is not set"
  exit 1
fi

if [ -z "$GITHUB_RELEASE_TOKEN" ]; then
  echo "[ERROR] GITHUB_RELEASE_TOKEN is not set"
  exit 1
fi

# Helper function for curl with retry, timeout, and progress
curl_retry() {
  local max_attempts=3
  local attempt=1
  local exit_code=0
  while [ $attempt -le $max_attempts ]; do
    echo "    [Attempt $attempt/$max_attempts] $(date +%H:%M:%S)"
    if curl --connect-timeout 30 --max-time 300 -w "\n    HTTP Status: %{http_code}, Time: %{time_total}s\n" "$@"; then
      return 0
    fi
    exit_code=$?
    echo "    [FAILED] curl exit code: $exit_code"
    if [ $attempt -lt $max_attempts ]; then
      echo "    Waiting 5s before retry..."
      sleep 5
    fi
    attempt=$((attempt + 1))
  done
  echo "    [ERROR] All $max_attempts attempts failed"
  return $exit_code
}

echo "=============================================="
echo "GitHub Release for $CI_COMMIT_TAG"
echo "=============================================="
echo "Start time: $(date)"
echo ""

# List available artifacts
echo "[Step 1/5] Listing artifacts..."
echo "  CLI artifacts (runner/dist/):"
ls -lh runner/dist/*.tar.gz runner/dist/*.zip 2>/dev/null || echo "    (none)"
echo "  Linux packages (runner/dist/):"
ls -lh runner/dist/*.deb runner/dist/*.rpm runner/dist/*.apk 2>/dev/null || echo "    (none)"
echo "  Desktop artifacts (runner/):"
ls -lh runner/*.dmg runner/AgentsMesh*.tar.gz runner/AgentsMesh*.zip 2>/dev/null || echo "    (none)"
echo ""

# Check if release already exists
echo "[Step 2/5] Checking existing release..."
EXISTING_RELEASE=$(curl_retry -s \
  -H "Authorization: token $GITHUB_RELEASE_TOKEN" \
  "${GITHUB_API}/releases/tags/$CI_COMMIT_TAG" 2>&1) || true

EXISTING_ID=$(echo "$EXISTING_RELEASE" | jq -r '.id // empty')

if [ -n "$EXISTING_ID" ] && [ "$EXISTING_ID" != "null" ]; then
  echo "  Found existing release (ID: $EXISTING_ID), deleting..."
  curl_retry -s -X DELETE \
    -H "Authorization: token $GITHUB_RELEASE_TOKEN" \
    "${GITHUB_API}/releases/$EXISTING_ID" || true
  echo "  Waiting 2s for GitHub to process deletion..."
  sleep 2
else
  echo "  No existing release found"
fi
echo ""

# Create release
echo "[Step 3/5] Creating new release..."
RELEASE_RESPONSE=$(curl_retry -s -X POST \
  -H "Authorization: token $GITHUB_RELEASE_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"tag_name\":\"$CI_COMMIT_TAG\",\"name\":\"AgentsMesh Runner $CI_COMMIT_TAG\",\"draft\":false,\"prerelease\":false}" \
  "${GITHUB_API}/releases")

RELEASE_ID=$(echo "$RELEASE_RESPONSE" | jq -r '.id // empty')
UPLOAD_URL=$(echo "$RELEASE_RESPONSE" | jq -r '.upload_url // empty' | sed 's/{.*}//')

if [ -z "$UPLOAD_URL" ] || [ "$UPLOAD_URL" = "null" ]; then
  echo "  [ERROR] Failed to create release!"
  echo "  Response: $RELEASE_RESPONSE"
  exit 1
fi

echo "  Release created successfully!"
echo "  Release ID: $RELEASE_ID"
echo "  Upload URL: $UPLOAD_URL"
echo ""

# Upload function
upload_file() {
  local file="$1"
  local filename=$(basename "$file")
  local filesize=$(ls -lh "$file" | awk '{print $5}')

  upload_count=$((upload_count + 1))
  echo "  [$upload_count] $filename ($filesize)"

  if curl_retry -s -X POST \
    -H "Authorization: token $GITHUB_RELEASE_TOKEN" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @"$file" \
    "$UPLOAD_URL?name=$filename" > /dev/null; then
    echo "    [OK] Uploaded successfully"
  else
    echo "    [FAILED] Upload failed"
    upload_failed=$((upload_failed + 1))
  fi
}

upload_count=0
upload_failed=0

# Upload CLI artifacts
echo "[Step 4/5] Uploading CLI artifacts..."
for file in runner/dist/*.tar.gz runner/dist/*.zip; do
  [ -f "$file" ] && upload_file "$file"
done

# Upload Linux packages
for file in runner/dist/*.deb runner/dist/*.rpm runner/dist/*.apk; do
  [ -f "$file" ] && upload_file "$file"
done
echo ""

# Upload Desktop artifacts
echo "[Step 5/5] Uploading Desktop artifacts..."
for file in runner/*.dmg runner/*.tar.gz runner/*.zip; do
  if [ -f "$file" ] && case "$(basename "$file")" in AgentsMesh*) true;; *) false;; esac; then
    upload_file "$file"
  fi
done
echo ""

# Summary
echo "=============================================="
echo "Release Summary"
echo "=============================================="
echo "End time: $(date)"
echo "Files uploaded: $upload_count"
echo "Files failed: $upload_failed"
echo "Release URL: https://github.com/${GITHUB_REPO}/releases/tag/$CI_COMMIT_TAG"
echo ""

if [ $upload_failed -gt 0 ]; then
  echo "[WARNING] Some uploads failed!"
  exit 1
fi

echo "[SUCCESS] Release complete!"
