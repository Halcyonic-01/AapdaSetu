# AapdaSetu (आपदा सेतु)
### Offline Emergency Communication Mesh Network

AapdaSetu is a decentralized, offline-first peer-to-peer (P2P) emergency communication system designed for disaster scenarios where cellular networks and internet connectivity are severed.

---

## 🌟 Key Features

- **Zero-Internet Communication**: Operates entirely over local Wi-Fi or mobile hotspots without requiring internet access or central cloud servers.
- **mDNS Auto-Discovery**: Nodes automatically detect and connect to nearby peers on the local network.
- **libp2p Mesh Networking**: Uses GossipSub for multi-hop P2P message propagation.
- **Emergency Broadcasts**: Instant high-priority SOS emergency alert distribution.
- **Modern Responsive UI**: Built with React, TypeScript, and Tailwind CSS.
- **Containerized**: Ready to deploy with Docker and Docker Compose.

---

## 🏗️ Architecture

```
Nearby Devices  ──(mDNS Auto-Discovery)──►  libp2p Host Mesh
                                                   │
                                            (GossipSub PubSub)
                                                   │
                                          Go Backend Node (:8080)
                                                   │
                                          React Frontend (:3000)
```

---

## 🚀 One-Click Start (Recommended)

Run both the Go backend and React frontend simultaneously with a single command:

```bash
./start.sh
```

This will automatically check dependencies, start the backend node, launch the frontend dev server, and open `http://localhost:3000` in your browser. Press `Ctrl+C` anytime to stop all services cleanly.

---

## 🛠️ Manual Development Start

### Prerequisites
- Go 1.24+
- Node.js 20+ & npm

### 1. Run Backend Node
```bash
cd backend
go run ./cmd/node --http-port=8080 --p2p-port=9000 --node-name="Rescue-Alpha"
```

Verify backend health:
```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/status
```

### 2. Run Frontend Web Interface
```bash
cd frontend
npm install
npm run dev
```
Open `http://localhost:3000` in your browser.

---

## 🐳 Running with Docker

```bash
docker compose -f deployments/docker/docker-compose.yml up --build
```

---

## 📁 Repository Structure

```
AapdaSetu/
├── backend/
│   ├── cmd/node/             # Main node entry point
│   ├── config/               # Configuration & CLI flag parser
│   ├── internal/
│   │   ├── api/              # HTTP REST handlers & CORS middleware
│   │   ├── p2p/              # libp2p host & discovery (Phase 2)
│   │   ├── broadcast/        # Emergency broadcast logic (Phase 3)
│   │   └── peers/            # Peer state management (Phase 2 & 3)
│   ├── go.mod & go.sum       # Go module with libp2p dependencies
├── frontend/
│   ├── src/
│   │   ├── components/       # UI components (chat, peers, broadcast)
│   │   ├── services/         # API client service
│   │   ├── types/            # TypeScript interfaces
│   │   └── App.tsx           # Main application shell
│   ├── vite.config.ts        # Vite dev server + backend proxy
├── deployments/
│   └── docker/               # Dockerfile.backend, Dockerfile.frontend, docker-compose.yml
└── scripts/                  # Development and test helper scripts
```

---

## 🗺️ Implementation Roadmap

- [x] **Phase 1: Project Bootstrap & Skeleton** (Go HTTP node, libp2p dependencies, React Vite UI shell, Docker configs)
- [ ] **Phase 2: P2P Mesh & mDNS Discovery** (libp2p Host, mDNS service, Peer tracker)
- [ ] **Phase 3: Emergency Chat & Broadcast** (GossipSub topic handlers, REST dispatchers)
- [ ] **Phase 4: Full React Dashboard** (Real-time message feed, SOS modal, peer list)
- [ ] **Phase 5: Docker Packaging & CI/CD** (GitHub Actions, multi-node testing)
