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
echo "[2/4] Starting Station Bravo (HTTP :8081 | P2P :9001)..."
go run ./cmd/node \
    --http-port=8081 \
    --p2p-port=9001 \
    --node-name="Station-Bravo-Outpost" \
    --rendezvous="aapdasetu-local-sim" &
BRAVO_PID=$!

# Wait for both nodes to initialize their APIs
sleep 2

# Link nodes into local mesh
echo "[3/5] Linking Station Alpha and Station Bravo into local mesh..."
for attempt in {1..10}; do
    ALPHA_ADDR=$(python3 -c "import json, urllib.request; data=json.loads(urllib.request.urlopen('http://localhost:8080/api/status', timeout=2).read()); print([a for a in data.get('addresses', []) if '127.0.0.1' in a][0])" 2>/dev/null || true)
    if [ -n "$ALPHA_ADDR" ]; then
        RES=$(curl -s -X POST http://localhost:8081/api/peers/connect \
            -H "Content-Type: application/json" \
            -d "{\"address\":\"$ALPHA_ADDR\"}" 2>/dev/null || true)
        if echo "$RES" | grep -q "connected"; then
            echo "  ✓ Mesh connection established between Alpha and Bravo!"
            break
        fi
    fi
    sleep 1
done

# 4. Start Frontend Dev Servers (One for Alpha, One for Bravo)
echo "[4/5] Starting Station Alpha UI on http://localhost:3000..."
cd "$ROOT_DIR/frontend"
VITE_BACKEND_URL=http://localhost:8080 npm run dev -- --port 3000 &

sleep 1

echo "[5/5] Starting Station Bravo UI on http://localhost:3001..."
VITE_BACKEND_URL=http://localhost:8081 npm run dev -- --port 3001 &

echo ""
echo "=========================================================="
echo "   AapdaSetu Dual-Station Simulation is LIVE!             "
echo "=========================================================="
echo " Open these two URLs side by side in your browser:"
echo "   1. Station Alpha (HQ):      http://localhost:3000"
echo "   2. Station Bravo (Outpost): http://localhost:3001"
echo ""
echo " What to try:"
echo "   - Send a chat message from Tab 1 -> Tab 2 receives it instantly!"
echo "   - Click 'Emergency Broadcast' in Tab 2 -> Tab 1 beeps and shows red bulletin!"
echo "   - Check the sidebar -> each station shows the other as an active peer!"
echo ""
echo " Press [Ctrl+C] to stop all simulation processes."
echo "=========================================================="

wait
