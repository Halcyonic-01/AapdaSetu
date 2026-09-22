package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/api"
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

	// Initialize API server
	apiServer := api.NewServer(cfg)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[Shutdown] Error during shutdown: %v", err)
	}
	log.Println("[Shutdown] Node gracefully stopped.")
}
