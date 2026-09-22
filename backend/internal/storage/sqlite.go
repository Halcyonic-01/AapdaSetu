package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/internal/broadcast"
	_ "modernc.org/sqlite"
)

// Store defines the interface for local message persistence.
type Store interface {
	SaveMessage(ctx context.Context, msg *broadcast.ChatMessage) error
	GetMessages(ctx context.Context, limit, offset int) ([]*broadcast.ChatMessage, error)
	Count(ctx context.Context) (int, error)
	Close() error
}

// SQLiteStore implements Store using an embedded SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens or creates an SQLite database file and initializes the schema.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	}

	// modernc.org/sqlite driver name is "sqlite"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", dbPath, err)
	}

	// Configure connection pool for SQLite concurrency
	db.SetMaxOpenConns(1) // SQLite works best with 1 writer connection

	store := &SQLiteStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates the messages table and performance index if they don't exist.
func (s *SQLiteStore) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		sender_id TEXT NOT NULL,
		sender_name TEXT,
		content TEXT NOT NULL,
		message_type TEXT NOT NULL,
		severity TEXT,
		timestamp DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC);
	`
	_, err := s.db.Exec(query)
	return err
}

// SaveMessage persists a chat or emergency message. Duplicates by ID are ignored.
func (s *SQLiteStore) SaveMessage(ctx context.Context, msg *broadcast.ChatMessage) error {
	if msg == nil {
		return nil
	}

	msgType := "chat"
	if msg.Emergency {
		msgType = "emergency"
	}

	severity := msg.Severity
	if severity == "" && msg.Emergency {
		severity = "critical"
	}

	query := `
	INSERT OR IGNORE INTO messages (id, sender_id, sender_name, content, message_type, severity, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?);
	`

	// Store timestamp in UTC format
	ts := msg.Timestamp.UTC()
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	_, err := s.db.ExecContext(ctx, query,
		msg.ID,
		msg.SenderID,
		msg.SenderName,
		msg.Body,
		msgType,
		severity,
		ts.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("failed to insert message %s: %w", msg.ID, err)
	}

	return nil
}

// GetMessages retrieves recent messages with pagination, returned in chronological order.
func (s *SQLiteStore) GetMessages(ctx context.Context, limit, offset int) ([]*broadcast.ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	// Use idx_messages_timestamp to fetch the latest messages
	query := `
	SELECT id, sender_id, sender_name, content, message_type, severity, timestamp
	FROM messages
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?;
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var descMessages []*broadcast.ChatMessage
	for rows.Next() {
		var (
			id, senderID, senderName, content, msgType, severity, tsStr string
		)

		if err := rows.Scan(&id, &senderID, &senderName, &content, &msgType, &severity, &tsStr); err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}

		t, err := time.Parse(time.RFC3339Nano, tsStr)
		if err != nil {
			// Fallback parse formats
			t, _ = time.Parse(time.RFC3339, tsStr)
		}

		descMessages = append(descMessages, &broadcast.ChatMessage{
			ID:         id,
			SenderID:   senderID,
			SenderName: senderName,
			Body:       content,
			Emergency:  msgType == "emergency",
			Severity:   severity,
			Timestamp:  t,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Reverse to return in chronological order (oldest to newest) for UI display
	n := len(descMessages)
	messages := make([]*broadcast.ChatMessage, n)
	for i, m := range descMessages {
		messages[n-1-i] = m
	}

	return messages, nil
}

// Count returns the total number of messages stored in the database.
func (s *SQLiteStore) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages;").Scan(&count)
	return count, err
}

// Close closes the underlying SQLite database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
