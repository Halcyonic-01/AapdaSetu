#!/bin/bash
set -e

# AapdaSetu — Multi-Node Local Mesh Simulation
# Runs two independent P2P nodes on different ports + frontend

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "=========================================================="
echo "    AapdaSetu — Local Multi-Node Simulation Runner       "
echo "=========================================================="

cleanup() {
    echo ""
    echo "Shutting down AapdaSetu mesh nodes and frontend..."
    kill $(jobs -p) 2>/dev/null || true
    wait 2>/dev/null || true
    echo "All simulation processes stopped cleanly."
}
trap cleanup EXIT INT TERM

# 1. Start Node Alpha (Primary Station)
echo "[1/3] Starting Station Alpha (HTTP :8080 | P2P :9000)..."
cd "$ROOT_DIR/backend"
go run ./cmd/node \
    --http-port=8080 \
    --p2p-port=9000 \
    --node-name="Station-Alpha-HQ" \
    --rendezvous="aapdasetu-local-sim" &
ALPHA_PID=$!

# Wait briefly for Alpha to bind ports
sleep 2

# 2. Start Node Bravo (Secondary Station)
echo "[2/3] Starting Station Bravo (HTTP :8081 | P2P :9001)..."
go run ./cmd/node \
    --http-port=8081 \
    --p2p-port=9001 \
    --node-name="Station-Bravo-Outpost" \
    --rendezvous="aapdasetu-local-sim" &
BRAVO_PID=$!

sleep 2

# 3. Start Frontend Dev Server
echo "[3/3] Starting Frontend Dev Server on http://localhost:3000..."
cd "$ROOT_DIR/frontend"
npm run dev -- --port 3000 &
FRONTEND_PID=$!

echo ""
echo "=========================================================="
echo "   AapdaSetu Local Simulation Mesh is Live!               "
echo "=========================================================="
echo " - Station Alpha Web UI:  http://localhost:3000"
echo " - Station Alpha API:     http://localhost:8080/api/health"
echo " - Station Bravo API:     http://localhost:8081/api/health"
echo ""
echo " Quick Test: Send broadcast from Station Bravo to Alpha:"
echo "   curl -X POST http://localhost:8081/api/broadcast/send \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"body\":\"TEST DISASTER BULLETIN FROM OUTPOST\",\"severity\":\"critical\"}'"
echo ""
echo " Press [Ctrl+C] to stop all simulation nodes."
echo "=========================================================="

wait
