#!/usr/bin/env bash
set -euo pipefail

SERVER_URL="${AGENTSMESH_SERVER_URL:-https://agentsmesh.int.rclabenv.com}"
LOGIN_URL="${AGENTSMESH_LOGIN_URL:-${SERVER_URL}/api/v1/auth/login}"
ORG_SLUG="${AGENTSMESH_ORG_SLUG:-default}"
TOKEN_URL="${AGENTSMESH_TOKEN_URL:-${SERVER_URL}/api/v1/orgs/${ORG_SLUG}/runners/grpc/tokens}"
LOGIN_EMAIL="${RUNNER_AGENTSMESH_LOGIN_EMAIL:-admin@localhost.local}"
LOGIN_PASSWORD="${RUNNER_AGENTSMESH_LOGIN_PASSWORD:-Admin@123}"
CONFIG_FILE="${HOME}/.agentsmesh/config.yaml"

echo "========================================"
echo "  AgentsMesh Runner Init"
echo "========================================"
echo "Server URL: ${SERVER_URL}"
echo "Org Slug:   ${ORG_SLUG}"
echo "Login URL:  ${LOGIN_URL}"
echo "Token URL:  ${TOKEN_URL}"

if [ -f "${CONFIG_FILE}" ]; then
    echo "Existing runner config found at ${CONFIG_FILE}, skipping registration."
else
    echo "Fetching JWT token..."

    LOGIN_RESPONSE="$(curl -fsSL -X POST \
        "${LOGIN_URL}" \
        -H "Content-Type: application/json" \
        --data-raw "{\"email\":\"${LOGIN_EMAIL}\",\"password\":\"${LOGIN_PASSWORD}\"}")"

    JWT_TOKEN="$(printf '%s' "${LOGIN_RESPONSE}" | python3 -c '
import json
import sys

data = json.load(sys.stdin)
print(data["token"])
')"

    if [ -z "${JWT_TOKEN}" ]; then
        echo "Failed to extract JWT token from login response."
        exit 1
    fi

    echo "Fetching registration token..."

    RESPONSE="$(curl -fsSL -X POST \
        "${TOKEN_URL}" \
        -H "Authorization: Bearer ${JWT_TOKEN}" \
        -H "Content-Type: application/json" \
        --data-raw '{}')"

    TOKEN="$(printf '%s' "${RESPONSE}" | python3 -c '
import json
import sys

data = json.load(sys.stdin)
print(data["token"])
')"

    if [ -z "${TOKEN}" ]; then
        echo "Failed to extract token from registration response."
        exit 1
    fi

    echo "Registering runner..."
    agentsmesh-runner register \
        --server "${SERVER_URL}" \
        --token "${TOKEN}"
fi

echo "Starting runner process..."
exec "$@"
