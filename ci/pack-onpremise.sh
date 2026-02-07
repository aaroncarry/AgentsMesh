#!/bin/bash
# =============================================================================
# Pack OnPremise Deployment Package
# =============================================================================
#
# Creates a complete deployment package for air-gapped installation.
#
# Usage:
#   ./pack-onpremise.sh [VERSION]
#
# Example:
#   ./pack-onpremise.sh v1.0.0
#
# Output:
#   agentsmesh-onpremise-{VERSION}.tar.gz
#
# Contents:
#   - deploy/onpremise/* (deployment files)
#   - images/*.tar (Docker images)
#   - runner binaries (optional)
#
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}/.."
VERSION="${1:-latest}"
OUTPUT_NAME="agentsmesh-onpremise-${VERSION}"
OUTPUT_DIR="${PROJECT_ROOT}/${OUTPUT_NAME}"
IMAGES_DIR="${PROJECT_ROOT}/images"

echo "=============================================="
echo "Packing AgentsMesh OnPremise Package"
echo "=============================================="
echo ""
echo "Version: ${VERSION}"
echo "Output:  ${OUTPUT_NAME}.tar.gz"
echo ""

# Check if images exist
if [ ! -d "${IMAGES_DIR}" ] || [ -z "$(ls -A ${IMAGES_DIR}/*.tar 2>/dev/null)" ]; then
    echo "Error: No images found in ${IMAGES_DIR}"
    echo "Please run build-onpremise.sh first."
    exit 1
fi

# Clean previous output
rm -rf "${OUTPUT_DIR}" "${OUTPUT_DIR}.tar.gz"

# Create output directory
mkdir -p "${OUTPUT_DIR}"

echo "[1/4] Copying deployment files..."
cp -r "${PROJECT_ROOT}/deploy/onpremise/"* "${OUTPUT_DIR}/"
echo "  Done."

echo "[2/4] Copying Docker images..."
mkdir -p "${OUTPUT_DIR}/images"
cp "${IMAGES_DIR}"/*.tar "${OUTPUT_DIR}/images/"
echo "  Done."

echo "[3/4] Setting permissions..."
chmod +x "${OUTPUT_DIR}/scripts/"*.sh
echo "  Done."

echo "[4/4] Creating archive..."
cd "${PROJECT_ROOT}"
tar -czvf "${OUTPUT_NAME}.tar.gz" "${OUTPUT_NAME}"
echo "  Done."

# Clean up extracted directory
rm -rf "${OUTPUT_DIR}"

# Summary
echo ""
echo "=============================================="
echo "Package Created Successfully!"
echo "=============================================="
echo ""
echo "Package: ${PROJECT_ROOT}/${OUTPUT_NAME}.tar.gz"
echo "Size:    $(du -h "${PROJECT_ROOT}/${OUTPUT_NAME}.tar.gz" | cut -f1)"
echo ""
echo "Contents:"
tar -tzf "${PROJECT_ROOT}/${OUTPUT_NAME}.tar.gz" | head -20
echo "  ..."
echo ""
echo "Installation:"
echo "  1. Copy ${OUTPUT_NAME}.tar.gz to target server"
echo "  2. tar -xzf ${OUTPUT_NAME}.tar.gz"
echo "  3. cd ${OUTPUT_NAME}"
echo "  4. ./scripts/install.sh --ip <SERVER_IP>"
echo ""
