package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/broadcast"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/storage"
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

func TestChatMessagesWithPaginationAndPersistence(t *testing.T) {
	dbStore, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("failed to init sqlite: %v", err)
	}
	defer dbStore.Close()

	ctx := context.Background()
	baseTime := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	for i := 1; i <= 60; i++ {
		msg := &broadcast.ChatMessage{
			ID:         fmt.Sprintf("msg-%03d", i),
			SenderID:   "peer-1",
			SenderName: "Alpha",
			Body:       fmt.Sprintf("Message %d", i),
			Timestamp:  baseTime.Add(time.Duration(i) * time.Minute),
			Emergency:  false,
		}
		if err := dbStore.SaveMessage(ctx, msg); err != nil {
			t.Fatalf("failed to save msg %d: %v", i, err)
		}
	}

	bMgr := broadcast.NewTestManager(dbStore)
	cfg := &config.Config{
		HTTPPort: 8080,
		P2PPort:  9000,
		NodeName: "TestNode",
	}
	server := NewServer(cfg, nil, nil, bMgr)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	// Test 1: Default limit (50) and offset (0) -> latest 50 messages: msg-011 through msg-060
	req1 := httptest.NewRequest("GET", "/api/chat/messages", nil)
	rr1 := httptest.NewRecorder()
	mux.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr1.Code)
	}

	var msgs1 []*broadcast.ChatMessage
	if err := json.Unmarshal(rr1.Body.Bytes(), &msgs1); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(msgs1) != 50 {
		t.Fatalf("expected 50 messages, got %d", len(msgs1))
	}
	if msgs1[0].ID != "msg-011" || msgs1[49].ID != "msg-060" {
		t.Fatalf("expected range msg-011..msg-060, got %s..%s", msgs1[0].ID, msgs1[49].ID)
	}

	// Test 2: Custom limit and offset (?limit=10&offset=50) -> oldest 10 messages: msg-001 through msg-010
	req2 := httptest.NewRequest("GET", "/api/chat/messages?limit=10&offset=50", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr2.Code)
	}

	var msgs2 []*broadcast.ChatMessage
	if err := json.Unmarshal(rr2.Body.Bytes(), &msgs2); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(msgs2) != 10 {
		t.Fatalf("expected 10 messages, got %d", len(msgs2))
	}
	if msgs2[0].ID != "msg-001" || msgs2[9].ID != "msg-010" {
		t.Fatalf("expected range msg-001..msg-010, got %s..%s", msgs2[0].ID, msgs2[9].ID)
	}
}

