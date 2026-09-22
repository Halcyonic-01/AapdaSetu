#!/bin/bash
set -e

echo "=========================================="
echo " Starting AapdaSetu in Development Mode   "
echo "=========================================="

cleanup() {
    echo ""
    echo "Stopping AapdaSetu processes..."
    kill $(jobs -p) 2>/dev/null || true
}
trap cleanup EXIT

echo "Starting Backend Node on :8080..."
cd "$(dirname "$0")/../backend"
go run ./cmd/node --http-port=8080 --p2p-port=9000 --node-name="Local-Dev-Node" &

echo "Starting Frontend Dev Server on :3000..."
cd "../frontend"
npm run dev &

wait
