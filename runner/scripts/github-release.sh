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

# Helper function for curl with retry and timeout
# Outputs response body to stdout, logs progress to stderr
curl_retry() {
  max_attempts=3
  attempt=1
  while [ $attempt -le $max_attempts ]; do
    echo "    [Attempt $attempt/$max_attempts] $(date +%H:%M:%S)" >&2

    # Create temp file for response body
    tmp_body=$(mktemp)
    tmp_headers=$(mktemp)

    http_code=$(curl --connect-timeout 30 --max-time 300 \
      -w "%{http_code}" \
      -o "$tmp_body" \
      -D "$tmp_headers" \
      "$@" 2>/dev/null) || true

    if [ -n "$http_code" ] && [ "$http_code" -ge 200 ] && [ "$http_code" -lt 400 ]; then
      echo "    [OK] HTTP $http_code" >&2
      cat "$tmp_body"
      rm -f "$tmp_body" "$tmp_headers"
      return 0
    fi

    echo "    [FAILED] HTTP $http_code" >&2
    if [ -f "$tmp_body" ]; then
      echo "    Response: $(cat "$tmp_body" | head -c 200)" >&2
    fi
    rm -f "$tmp_body" "$tmp_headers"

    if [ $attempt -lt $max_attempts ]; then
      echo "    Waiting 5s before retry..." >&2
      sleep 5
    fi
    attempt=$((attempt + 1))
  done
  echo "    [ERROR] All $max_attempts attempts failed" >&2
  return 1
}

# Helper function to safely parse JSON
json_get() {
  json="$1"
  field="$2"
  # Check if input looks like JSON
  if echo "$json" | grep -q '^{'; then
    echo "$json" | jq -r "$field // empty" 2>/dev/null || echo ""
  else
    echo ""
  fi
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
  "${GITHUB_API}/releases/tags/$CI_COMMIT_TAG") || true

EXISTING_ID=$(json_get "$EXISTING_RELEASE" '.id')

if [ -n "$EXISTING_ID" ]; then
  echo "  Found existing release (ID: $EXISTING_ID), deleting..."
  curl_retry -s -X DELETE \
    -H "Authorization: token $GITHUB_RELEASE_TOKEN" \
    "${GITHUB_API}/releases/$EXISTING_ID" > /dev/null || true
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

RELEASE_ID=$(json_get "$RELEASE_RESPONSE" '.id')
UPLOAD_URL=$(json_get "$RELEASE_RESPONSE" '.upload_url' | sed 's/{.*}//')

if [ -z "$UPLOAD_URL" ]; then
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
  file="$1"
  filename=$(basename "$file")
  filesize=$(ls -lh "$file" | awk '{print $5}')

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
