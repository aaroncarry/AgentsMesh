#!/bin/bash
# Sign and notarize macOS binary with Developer ID certificate using rcodesign.
# Called by GoReleaser post-build hook for darwin targets.
#
# Required for signing:
#   MACOS_CERTIFICATE              — base64-encoded .p12 certificate
#   MACOS_CERTIFICATE_PASSWORD     — .p12 password
#
# Optional for notarization (skipped if not set):
#   APPLE_API_KEY_ID               — App Store Connect API Key ID
#   APPLE_API_KEY_ISSUER_ID        — App Store Connect Issuer ID
#   APPLE_API_KEY                  — base64-encoded .p8 private key
set -euo pipefail

BINARY="$1"

if [ -z "${MACOS_CERTIFICATE:-}" ]; then
  if [ -n "${CI:-}" ] || [ -n "${GITHUB_ACTIONS:-}" ]; then
    echo "ERROR: MACOS_CERTIFICATE not set in CI — binary will be unsigned!" >&2
    exit 1
  fi
  echo "MACOS_CERTIFICATE not set, skipping sign (local build)"
  exit 0
fi

RCODESIGN="${RCODESIGN:-rcodesign}"
ENTITLEMENTS="${ENTITLEMENTS:-./build/darwin/entitlements.plist}"

echo "$MACOS_CERTIFICATE" | base64 -d > /tmp/cert.p12
trap 'rm -f /tmp/cert.p12 /tmp/notary-key.json /tmp/notary-submit.zip' EXIT

$RCODESIGN sign \
  --p12-file /tmp/cert.p12 \
  --p12-password "$MACOS_CERTIFICATE_PASSWORD" \
  --code-signature-flags runtime \
  --entitlements-xml-path "$ENTITLEMENTS" \
  "$BINARY"

echo "Signed: $BINARY"

# --- Notarization (optional) ---
# Apple Notary API requires signed binaries to be wrapped in a zip archive.
# Bare Mach-O files cannot be stapled, but Gatekeeper checks notarization
# status online after the binary is notarized.

if [ -z "${APPLE_API_KEY:-}" ]; then
  echo "APPLE_API_KEY not set, skipping notarization"
  exit 0
fi

echo "Preparing notarization credentials..."
$RCODESIGN encode-app-store-connect-api-key \
  -o /tmp/notary-key.json \
  "$APPLE_API_KEY_ID" \
  "$APPLE_API_KEY_ISSUER_ID" \
  <(echo "$APPLE_API_KEY" | base64 -d)

echo "Creating zip for notarization..."
zip -j /tmp/notary-submit.zip "$BINARY"

echo "Submitting to Apple Notary Service (this may take 2-10 minutes)..."
$RCODESIGN notary-submit \
  --api-key-file /tmp/notary-key.json \
  --wait \
  /tmp/notary-submit.zip

echo "Notarized: $BINARY"
