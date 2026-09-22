package broadcast

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// ChatMessage represents a peer-to-peer communication payload.
type ChatMessage struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Body       string    `json:"body"`
	Timestamp  time.Time `json:"timestamp"`
	Emergency  bool      `json:"emergency"`
	Severity   string    `json:"severity,omitempty"` // "critical", "warning", "info"
}

// NewMessage creates a new formatted ChatMessage.
func NewMessage(senderID, senderName, body string, emergency bool, severity string) *ChatMessage {
	if severity == "" && emergency {
		severity = "critical"
	}

	return &ChatMessage{
		ID:         generateRandomID(),
		SenderID:   senderID,
		SenderName: senderName,
		Body:       body,
		Timestamp:  time.Now().UTC(),
		Emergency:  emergency,
		Severity:   severity,
	}
}

// MessageStore maintains an in-memory chronological history of mesh messages.
type MessageStore struct {
	mu       sync.RWMutex
	messages []*ChatMessage
	maxSize  int
}

// NewMessageStore creates a new thread-safe message buffer.
func NewMessageStore(maxSize int) *MessageStore {
	if maxSize <= 0 {
		maxSize = 250
	}
	return &MessageStore{
		messages: make([]*ChatMessage, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add appends a message to the store, pruning older entries if capacity is reached.
func (s *MessageStore) Add(msg *ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prevent duplicate messages by ID
	for _, m := range s.messages {
		if m.ID == msg.ID {
			return
		}
	}

	if len(s.messages) >= s.maxSize {
		// Drop oldest 10% to prevent excessive re-allocations
		dropCount := s.maxSize / 10
		if dropCount < 1 {
			dropCount = 1
		}
		s.messages = s.messages[dropCount:]
	}

	s.messages = append(s.messages, msg)
}

// List returns all messages currently stored.
func (s *MessageStore) List() []*ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ChatMessage, len(s.messages))
	copy(result, s.messages)
	return result
}

// ListRecent returns up to `limit` most recent messages.
func (s *MessageStore) ListRecent(limit int) []*ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit >= len(s.messages) {
		result := make([]*ChatMessage, len(s.messages))
		copy(result, s.messages)
		return result
	}

	start := len(s.messages) - limit
	result := make([]*ChatMessage, limit)
	copy(result, s.messages[start:])
	return result
}

func generateRandomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(b)
}
