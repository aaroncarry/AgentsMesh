#!/usr/bin/env bash
set -euo pipefail

SERVER_URL="${AGENTSMESH_SERVER_URL:-https://agentsmesh.int.rclabenv.com}"
TOKEN_URL="${AGENTSMESH_TOKEN_URL:-https://agentsmesh.int.rclabenv.com/api/v1/orgs/default/runners/grpc/tokens}"
RUNNER_AGENTSMESH_API_KEY="${RUNNER_AGENTSMESH_API_KEY:?RUNNER_AGENTSMESH_API_KEY is required}"
CONFIG_FILE="${HOME}/.agentsmesh/config.yaml"

echo "========================================"
echo "  AgentsMesh Runner Init"
echo "========================================"
echo "Server URL: ${SERVER_URL}"
echo "Token URL:  ${TOKEN_URL}"

if [ -f "${CONFIG_FILE}" ]; then
    echo "Existing runner config found at ${CONFIG_FILE}, skipping registration."
else
    echo "Fetching registration token..."

    RESPONSE="$(curl -fsSL -X POST \
        "${TOKEN_URL}" \
        -H "Authorization: Bearer ${RUNNER_AGENTSMESH_API_KEY}" \
        -H "Content-Type: application/json")"

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
