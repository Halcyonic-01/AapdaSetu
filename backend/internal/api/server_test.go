package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 8080,
		P2PPort:  9000,
		NodeName: "TestNode",
	}
	server := NewServer(cfg)
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
	server := NewServer(cfg)
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
