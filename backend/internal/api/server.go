package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
)

// Server handles HTTP API requests for the AapdaSetu frontend.
type Server struct {
	cfg       *config.Config
	startTime time.Time
}

// NodeStatus represents the operational status of the AapdaSetu node.
type NodeStatus struct {
	NodeName  string    `json:"node_name"`
	Status    string    `json:"status"`
	P2PPort   int       `json:"p2p_port"`
	HTTPPort  int       `json:"http_port"`
	Uptime    string    `json:"uptime"`
	PeerCount int       `json:"peer_count"`
	Timestamp time.Time `json:"timestamp"`
}

// NewServer creates a new HTTP API Server instance.
func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

// RegisterRoutes registers HTTP routes with CORS support.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/status", s.handleStatus)
}

// CORS middleware wrapper.
func (s *Server) WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := NodeStatus{
		NodeName:  s.cfg.NodeName,
		Status:    "online",
		P2PPort:   s.cfg.P2PPort,
		HTTPPort:  s.cfg.HTTPPort,
		Uptime:    time.Since(s.startTime).Round(time.Second).String(),
		PeerCount: 0,
		Timestamp: time.Now().UTC(),
	}
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding status: %v", err), http.StatusInternalServerError)
	}
}
