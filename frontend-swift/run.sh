#!/bin/zsh

set -euo pipefail

SCRIPT_DIR="${0:A:h}"
PROJECT_DIR="${SCRIPT_DIR:h}"
BACKEND_DIR="${SCRIPT_DIR}/.build/backend"
BACKEND_BIN="${BACKEND_DIR}/mrrss-server"
SERVER_URL="http://127.0.0.1:1234/api/version"
BACKEND_PID=""

cleanup() {
    if [[ -n "${BACKEND_PID}" ]]; then
        kill "${BACKEND_PID}" 2>/dev/null || true
        wait "${BACKEND_PID}" 2>/dev/null || true
    fi
}

trap cleanup EXIT INT TERM

if ! curl --silent --fail --max-time 1 "${SERVER_URL}" >/dev/null 2>&1; then
    mkdir -p "${BACKEND_DIR}"
    echo "Building the MrRSS backend..."
    (cd "${PROJECT_DIR}" && go build -o "${BACKEND_BIN}" .)

    echo "Starting the MrRSS backend on 127.0.0.1:1234..."
    "${BACKEND_BIN}" -host 127.0.0.1 -port 1234 &
    BACKEND_PID=$!

    for _ in {1..100}; do
        if curl --silent --fail --max-time 1 "${SERVER_URL}" >/dev/null 2>&1; then
            break
        fi
        if ! kill -0 "${BACKEND_PID}" 2>/dev/null; then
            echo "The MrRSS backend exited before it became ready." >&2
            exit 1
        fi
        sleep 0.1
    done

    if ! curl --silent --fail --max-time 1 "${SERVER_URL}" >/dev/null 2>&1; then
        echo "The MrRSS backend did not become ready within 10 seconds." >&2
        exit 1
    fi
fi

echo "Starting the MrRSS SwiftUI frontend..."
cd "${SCRIPT_DIR}"
swift run MrRSS
