#!/usr/bin/env bash

# AapdaSetu - One-Click Launcher
# Starts Go Backend Node + React Frontend with unified logs and graceful shutdown.

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${RED}${BOLD}"
echo "========================================================="
echo "   AapdaSetu (आपदा सेतु) - Emergency Mesh Network       "
echo "========================================================="
echo -e "${NC}"

# Check prerequisites
if ! command -v go &> /dev/null; then
    echo -e "${RED}[ERROR] Go is not installed or not in PATH.${NC}"
    exit 1
fi

if ! command -v node &> /dev/null || ! command -v npm &> /dev/null; then
    echo -e "${RED}[ERROR] Node.js / npm is not installed or not in PATH.${NC}"
    exit 1
fi

# Ensure frontend dependencies are installed
if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
    echo -e "${YELLOW}[SETUP] Frontend node_modules missing. Running npm install...${NC}"
    (cd "$FRONTEND_DIR" && npm install)
fi

# PIDs tracking
BACKEND_PID=""
FRONTEND_PID=""

cleanup() {
    echo -e "\n${YELLOW}[SHUTDOWN] Stopping all AapdaSetu services...${NC}"
    if [ -n "$BACKEND_PID" ]; then
        kill "$BACKEND_PID" 2>/dev/null || true
    fi
    if [ -n "$FRONTEND_PID" ]; then
        kill "$FRONTEND_PID" 2>/dev/null || true
    fi
    # Also kill any child jobs in this process group
    kill $(jobs -p) 2>/dev/null || true
    echo -e "${GREEN}[SHUTDOWN] All services stopped cleanly.${NC}"
    exit 0
}

# Trap signals for graceful exit
trap cleanup SIGINT SIGTERM EXIT

# 1. Start Go Backend
echo -e "${CYAN}[1/2] Starting Go Mesh Backend Node...${NC}"
cd "$BACKEND_DIR"
go run ./cmd/node --http-port=8080 --p2p-port=9000 --node-name="Rescue-Node-1" &
BACKEND_PID=$!

# Wait for backend to become healthy
echo -e "${CYAN}      Waiting for backend to initialize on port 8080...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        echo -e "${GREEN}      Backend is UP and healthy!${NC}"
        break
    fi
    sleep 0.5
done

# 2. Start Frontend
echo -e "${CYAN}[2/2] Starting React Frontend...${NC}"
cd "$FRONTEND_DIR"
npm run dev -- --host &
FRONTEND_PID=$!

# Wait briefly for Vite to spin up
sleep 1.5

echo -e "\n${GREEN}${BOLD}=========================================================${NC}"
echo -e "${GREEN}${BOLD}   AapdaSetu Emergency System is running!                ${NC}"
echo -e "${GREEN}${BOLD}=========================================================${NC}"
echo -e "   ${BOLD}Web Interface:${NC} ${BLUE}http://localhost:3000${NC}"
echo -e "   ${BOLD}Backend API:${NC}   ${BLUE}http://localhost:8080/api/status${NC}"
echo -e "   ${BOLD}P2P Mesh Port:${NC} ${BLUE}9000${NC}"
echo -e "---------------------------------------------------------"
echo -e "   ${YELLOW}Press Ctrl+C to terminate all services.${NC}"
echo -e "${GREEN}${BOLD}=========================================================${NC}\n"

# Open browser if on macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    open "http://localhost:3000" 2>/dev/null || true
fi

# Wait on background processes
wait
