package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/broadcast"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/p2p"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
)

// Server handles HTTP API requests for the AapdaSetu frontend.
type Server struct {
	cfg          *config.Config
	p2pNode      *p2p.Node
	peerStore    *peers.Store
	broadcastMgr *broadcast.Manager
	startTime    time.Time
}

// NodeStatus represents the operational status of the AapdaSetu node.
type NodeStatus struct {
	NodeName   string    `json:"node_name"`
	Status     string    `json:"status"`
	PeerID     string    `json:"peer_id"`
	Addresses  []string  `json:"addresses"`
	P2PPort    int       `json:"p2p_port"`
	HTTPPort   int       `json:"http_port"`
	Uptime     string    `json:"uptime"`
	PeerCount  int       `json:"peer_count"`
	Rendezvous string    `json:"rendezvous"`
	Timestamp  time.Time `json:"timestamp"`
}

// SendChatRequest represents an inbound message creation request.
type SendChatRequest struct {
	Body string `json:"body"`
}

// SendBroadcastRequest represents an inbound emergency alert broadcast.
type SendBroadcastRequest struct {
	Body     string `json:"body"`
	Severity string `json:"severity,omitempty"`
}

// ConnectRequest defines the payload for dialing a peer directly.
type ConnectRequest struct {
	Address string `json:"address"`
}

// NewServer creates a new HTTP API Server instance.
func NewServer(cfg *config.Config, p2pNode *p2p.Node, peerStore *peers.Store, broadcastMgr *broadcast.Manager) *Server {
	return &Server{
		cfg:          cfg,
		p2pNode:      p2pNode,
		peerStore:    peerStore,
		broadcastMgr: broadcastMgr,
		startTime:    time.Now(),
	}
}

// RegisterRoutes registers HTTP routes with CORS support.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/peers", s.handlePeers)
	mux.HandleFunc("/api/peers/connect", s.handleConnectPeer)
	mux.HandleFunc("/api/chat/send", s.handleChatSend)
	mux.HandleFunc("/api/chat/messages", s.handleChatMessages)
	mux.HandleFunc("/api/broadcast/send", s.handleBroadcastSend)
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

	peerID := ""
	var addrs []string
	peerCount := 0

	if s.p2pNode != nil {
		peerID = s.p2pNode.ID().String()
		addrs = s.p2pNode.Addrs()
	}

	if s.peerStore != nil {
		peerCount = s.peerStore.Count()
	}

	status := NodeStatus{
		NodeName:   s.cfg.NodeName,
		Status:     "online",
		PeerID:     peerID,
		Addresses:  addrs,
		P2PPort:    s.cfg.P2PPort,
		HTTPPort:   s.cfg.HTTPPort,
		Uptime:     time.Since(s.startTime).Round(time.Second).String(),
		PeerCount:  peerCount,
		Rendezvous: s.cfg.Rendezvous,
		Timestamp:  time.Now().UTC(),
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding status: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.peerStore == nil {
		json.NewEncoder(w).Encode([]*peers.PeerInfo{})
		return
	}

	peerList := s.peerStore.List()
	if err := json.NewEncoder(w).Encode(peerList); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding peers: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) handleConnectPeer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, "Missing peer address", http.StatusBadRequest)
		return
	}

	if s.p2pNode == nil {
		http.Error(w, "P2P node not initialized", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := s.p2pNode.ConnectDirect(ctx, req.Address); err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "connected",
		"address": req.Address,
	})
}

func (s *Server) handleChatSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.broadcastMgr == nil {
		http.Error(w, "Broadcast engine not available", http.StatusServiceUnavailable)
		return
	}

	var req SendChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	msg, err := s.broadcastMgr.SendChat(r.Context(), req.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to dispatch message: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func (s *Server) handleChatMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.broadcastMgr == nil {
		json.NewEncoder(w).Encode([]*broadcast.ChatMessage{})
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages := s.broadcastMgr.Store().ListRecent(limit)
	json.NewEncoder(w).Encode(messages)
}

func (s *Server) handleBroadcastSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.broadcastMgr == nil {
		http.Error(w, "Broadcast engine not available", http.StatusServiceUnavailable)
		return
	}

	var req SendBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	msg, err := s.broadcastMgr.SendBroadcast(r.Context(), req.Body, req.Severity)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to dispatch broadcast: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}
