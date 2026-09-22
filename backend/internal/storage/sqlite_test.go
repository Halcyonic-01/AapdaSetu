package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/internal/broadcast"
)

func TestSQLiteStore_Basic(t *testing.T) {
	ctx := context.Background()
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory sqlite: %v", err)
	}
	defer store.Close()

	// 1. Verify empty
	msgs, err := store.GetMessages(ctx, 50, 0)
	if err != nil {
		t.Fatalf("unexpected error getting empty messages: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}

	// 2. Save a chat message
	m1 := &broadcast.ChatMessage{
		ID:         "msg-1",
		SenderID:   "peer-alpha",
		SenderName: "Alpha-Station",
		Body:       "All systems nominal in Sector 1",
		Timestamp:  time.Now().UTC().Add(-2 * time.Minute),
		Emergency:  false,
		Severity:   "info",
	}
	if err := store.SaveMessage(ctx, m1); err != nil {
		t.Fatalf("failed to save chat message: %v", err)
	}

	// 3. Save an emergency message
	m2 := &broadcast.ChatMessage{
		ID:         "msg-2",
		SenderID:   "peer-bravo",
		SenderName: "Bravo-Station",
		Body:       "URGENT: Flood waters breached barrier",
		Timestamp:  time.Now().UTC().Add(-1 * time.Minute),
		Emergency:  true,
		Severity:   "critical",
	}
	if err := store.SaveMessage(ctx, m2); err != nil {
		t.Fatalf("failed to save emergency message: %v", err)
	}

	// 4. Duplicate ID must be ignored
	if err := store.SaveMessage(ctx, m1); err != nil {
		t.Fatalf("failed on duplicate save: %v", err)
	}

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("count error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}

	// 5. Retrieve messages - should be in chronological order: m1 then m2
	retrieved, err := store.GetMessages(ctx, 50, 0)
	if err != nil {
		t.Fatalf("failed to retrieve messages: %v", err)
	}
	if len(retrieved) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(retrieved))
	}
	if retrieved[0].ID != "msg-1" || retrieved[1].ID != "msg-2" {
		t.Fatalf("expected chronological order [msg-1, msg-2], got [%s, %s]", retrieved[0].ID, retrieved[1].ID)
	}
	if !retrieved[1].Emergency || retrieved[1].Severity != "critical" {
		t.Fatalf("expected emergency flag and critical severity on msg-2")
	}
}

func TestSQLiteStore_Pagination(t *testing.T) {
	ctx := context.Background()
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory sqlite: %v", err)
	}
	defer store.Close()

	// Insert 75 messages with sequential timestamps
	baseTime := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)
	for i := 1; i <= 75; i++ {
		msg := &broadcast.ChatMessage{
			ID:         fmt.Sprintf("msg-%03d", i),
			SenderID:   "peer-node",
			SenderName: "TestNode",
			Body:       fmt.Sprintf("Message payload %d", i),
			Timestamp:  baseTime.Add(time.Duration(i) * time.Second),
			Emergency:  false,
		}
		if err := store.SaveMessage(ctx, msg); err != nil {
			t.Fatalf("failed to save message %d: %v", i, err)
		}
	}

	// Fetch latest 50 (limit=50, offset=0) -> should be messages 26..75
	page1, err := store.GetMessages(ctx, 50, 0)
	if err != nil {
		t.Fatalf("page 1 error: %v", err)
	}
	if len(page1) != 50 {
		t.Fatalf("expected 50 messages, got %d", len(page1))
	}
	if page1[0].ID != "msg-026" || page1[49].ID != "msg-075" {
		t.Fatalf("expected page1 [msg-026 ... msg-075], got [%s ... %s]", page1[0].ID, page1[49].ID)
	}

	// Fetch next page (limit=50, offset=50) -> should be remaining 25 messages: 1..25
	page2, err := store.GetMessages(ctx, 50, 50)
	if err != nil {
		t.Fatalf("page 2 error: %v", err)
	}
	if len(page2) != 25 {
		t.Fatalf("expected 25 messages on page 2, got %d", len(page2))
	}
	if page2[0].ID != "msg-001" || page2[24].ID != "msg-025" {
		t.Fatalf("expected page2 [msg-001 ... msg-025], got [%s ... %s]", page2[0].ID, page2[24].ID)
	}
}

func TestSQLiteStore_PersistenceAcrossRestart(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "aapdasetu-persist-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbFile := filepath.Join(tempDir, "node_persistence.db")

	// Phase 1: Open node store, write messages, then simulate node shutdown
	store1, err := NewSQLiteStore(dbFile)
	if err != nil {
		t.Fatalf("failed to create store1: %v", err)
	}

	msg := &broadcast.ChatMessage{
		ID:         "persist-001",
		SenderID:   "station-alpha",
		SenderName: "Alpha-Station",
		Body:       "This message MUST survive node restart",
		Timestamp:  time.Now().UTC(),
		Emergency:  true,
		Severity:   "critical",
	}
	if err := store1.SaveMessage(ctx, msg); err != nil {
		t.Fatalf("failed to save message: %v", err)
	}

	// Simulate node shutdown
	if err := store1.Close(); err != nil {
		t.Fatalf("failed to close store1: %v", err)
	}

	// Phase 2: Simulate node restart by reopening the same database file
	store2, err := NewSQLiteStore(dbFile)
	if err != nil {
		t.Fatalf("failed to open store2 after restart: %v", err)
	}
	defer store2.Close()

	msgs, err := store2.GetMessages(ctx, 50, 0)
	if err != nil {
		t.Fatalf("failed to get messages after restart: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 persisted message after restart, got %d", len(msgs))
	}

	restored := msgs[0]
	if restored.ID != "persist-001" || restored.Body != "This message MUST survive node restart" {
		t.Fatalf("restored message content mismatch: %+v", restored)
	}
	if !restored.Emergency || restored.Severity != "critical" {
		t.Fatalf("restored message type/severity mismatch: %+v", restored)
	}
}
