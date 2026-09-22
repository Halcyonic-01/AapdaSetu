package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/broadcast"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		P2PPort:  9000,
		NodeName: "TestNode",
	}
	server := NewServer(cfg, nil, nil, nil)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp["status"])
	}
}

func TestStatusEndpoint(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		P2PPort:  9000,
		NodeName: "Emergency-Unit-1",
	}
	peerStore := peers.NewStore()
	server := NewServer(cfg, nil, peerStore, nil)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/status", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var status NodeStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if status.NodeName != "Emergency-Unit-1" {
		t.Errorf("expected NodeName 'Emergency-Unit-1', got '%s'", status.NodeName)
	}
	if status.Status != "online" {
		t.Errorf("expected status 'online', got '%s'", status.Status)
	}
}

func TestPeersEndpoint(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		P2PPort:  9000,
		NodeName: "Emergency-Unit-1",
	}
	peerStore := peers.NewStore()
	server := NewServer(cfg, nil, peerStore, nil)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/peers", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var peerList []*peers.PeerInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &peerList); err != nil {
		t.Fatalf("failed to decode peers response: %v", err)
	}

	if len(peerList) != 0 {
		t.Errorf("expected 0 peers initially, got %d", len(peerList))
	}
}

func TestChatMessagesEmptyEndpoint(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		P2PPort:  9000,
		NodeName: "Emergency-Unit-1",
	}
	server := NewServer(cfg, nil, nil, nil)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/chat/messages", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var msgs []*broadcast.ChatMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &msgs); err != nil {
		t.Fatalf("failed to decode messages response: %v", err)
	}

	if len(msgs) != 0 {
		t.Errorf("expected 0 messages initially, got %d", len(msgs))
	}
}
