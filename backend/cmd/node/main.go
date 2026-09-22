package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/api"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/broadcast"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/p2p"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/storage"
)

func main() {
	log.Println("==================================================")
	log.Println("   AapdaSetu - Offline Emergency Mesh Node       ")
	log.Println("==================================================")

	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("[Config] Node Name: %s", cfg.NodeName)
	log.Printf("[Config] HTTP Port: %d | P2P Port: %d", cfg.HTTPPort, cfg.P2PPort)
	log.Printf("[Config] Discovery Rendezvous: %s", cfg.Rendezvous)

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// 1. Initialize SQLite Message Store
	dbPath := filepath.Join(cfg.DataDir, "aapdasetu.db")
	sqliteStore, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("[Storage] Failed to initialize SQLite store: %v", err)
	}
	defer sqliteStore.Close()
	log.Printf("[Storage] Local SQLite database active: %s", dbPath)

	// 2. Initialize Peer Store
	peerStore := peers.NewStore()

	// 3. Initialize P2P Mesh Node (libp2p Host + mDNS Discovery + GossipSub)
	p2pNode, err := p2p.NewNode(rootCtx, cfg, peerStore)
	if err != nil {
		log.Fatalf("[P2P Mesh] Failed to initialize node: %v", err)
	}
	defer p2pNode.Close()

	// 4. Initialize Emergency Broadcast & Chat Engine
	broadcastMgr, err := broadcast.NewManager(rootCtx, p2pNode, cfg, nil)
	if err != nil {
		log.Fatalf("[Broadcast] Failed to initialize broadcast engine: %v", err)
	}
	defer broadcastMgr.Close()
	broadcastMgr.SetPersister(sqliteStore)

	// Connect to bootstrap peer if specified
	if cfg.BootstrapPeer != "" {
		go func() {
			time.Sleep(500 * time.Millisecond)
			if err := p2pNode.ConnectDirect(rootCtx, cfg.BootstrapPeer); err != nil {
				log.Printf("[P2P Mesh] Failed to connect to bootstrap peer %s: %v", cfg.BootstrapPeer, err)
			} else {
				log.Printf("[P2P Mesh] Connected to bootstrap peer: %s", cfg.BootstrapPeer)
			}
		}()
	}

	// 4. Initialize API Server
	apiServer := api.NewServer(cfg, p2pNode, peerStore, broadcastMgr)
	mux := http.NewServeMux()
	apiServer.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      apiServer.WithCORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start HTTP server in background
	go func() {
		log.Printf("[HTTP] API server listening at http://localhost:%d", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[HTTP] Server failed: %v", err)
		}
	}()

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("\n[Shutdown] Shutting down AapdaSetu node...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Shutdown] Error during HTTP shutdown: %v", err)
	}

	broadcastMgr.Close()

	if err := p2pNode.Close(); err != nil {
		log.Printf("[Shutdown] Error stopping P2P services: %v", err)
	}

	log.Println("[Shutdown] Node gracefully stopped.")
}
