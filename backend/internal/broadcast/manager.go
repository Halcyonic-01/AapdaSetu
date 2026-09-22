package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/p2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Persister is an optional storage backend for persistent message retention (e.g., SQLite).
type Persister interface {
	SaveMessage(ctx context.Context, msg *ChatMessage) error
	GetMessages(ctx context.Context, limit, offset int) ([]*ChatMessage, error)
}

// Manager coordinates chat and emergency broadcast messaging across GossipSub topics.
type Manager struct {
	node      *p2p.Node
	cfg       *config.Config
	store     *MessageStore
	persister Persister
	chatSub   *pubsub.Subscription
	alertSub  *pubsub.Subscription
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewTestManager creates a Manager without pubsub subscriptions for unit testing purposes.
func NewTestManager(persister Persister) *Manager {
	return &Manager{
		persister: persister,
		store:     NewMessageStore(100),
	}
}

// NewManager creates and starts a new GossipSub broadcast manager.
func NewManager(parentCtx context.Context, node *p2p.Node, cfg *config.Config, store *MessageStore) (*Manager, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	if store == nil {
		store = NewMessageStore(250)
	}

	// Subscribe to standard chat topic
	chatSub, err := node.PubSub().Subscribe(cfg.ChatTopic)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to subscribe to chat topic '%s': %w", cfg.ChatTopic, err)
	}

	// Subscribe to emergency alert topic
	alertSub, err := node.PubSub().Subscribe(cfg.AlertTopic)
	if err != nil {
		chatSub.Cancel()
		cancel()
		return nil, fmt.Errorf("failed to subscribe to alert topic '%s': %w", cfg.AlertTopic, err)
	}

	mgr := &Manager{
		node:     node,
		cfg:      cfg,
		store:    store,
		chatSub:  chatSub,
		alertSub: alertSub,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start background listeners
	go mgr.listen(chatSub, "Chat")
	go mgr.listen(alertSub, "SOS-Alert")

	log.Printf("[Broadcast] Manager initialized (Topics: '%s', '%s')", cfg.ChatTopic, cfg.AlertTopic)
	return mgr, nil
}

// listen reads inbound messages from a subscription channel until cancelled.
func (m *Manager) listen(sub *pubsub.Subscription, label string) {
	for {
		msg, err := sub.Next(m.ctx)
		if err != nil {
			if m.ctx.Err() != nil {
				return // Context cancelled, shutdown cleanly
			}
			log.Printf("[Broadcast] Subscription error on %s: %v", label, err)
			return
		}

		// Don't duplicate messages originated from self
		if msg.ReceivedFrom == m.node.ID() {
			continue
		}

		var chatMsg ChatMessage
		if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
			log.Printf("[Broadcast] Failed to unmarshal message on %s: %v", label, err)
			continue
		}

		log.Printf("[Broadcast][%s] From %s (%s): %s", label, chatMsg.SenderName, chatMsg.SenderID, chatMsg.Body)
		m.store.Add(&chatMsg)
		if m.persister != nil {
			if err := m.persister.SaveMessage(m.ctx, &chatMsg); err != nil {
				log.Printf("[Broadcast] Failed to persist inbound message %s: %v", chatMsg.ID, err)
			}
		}
	}
}

// SendChat dispatches a standard chat message across the mesh.
func (m *Manager) SendChat(ctx context.Context, body string) (*ChatMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("message body cannot be empty")
	}

	msg := NewMessage(m.node.ID().String(), m.cfg.NodeName, body, false, "info")

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize chat message: %w", err)
	}

	if err := m.node.PubSub().Publish(ctx, m.cfg.ChatTopic, data); err != nil {
		return nil, fmt.Errorf("failed to publish to chat topic: %w", err)
	}

	// Record in memory store and persistent storage
	m.store.Add(msg)
	if m.persister != nil {
		if err := m.persister.SaveMessage(ctx, msg); err != nil {
			log.Printf("[Broadcast] Failed to persist sent chat message %s: %v", msg.ID, err)
		}
	}
	return msg, nil
}

// SendBroadcast dispatches a high-priority emergency broadcast alert across the mesh.
func (m *Manager) SendBroadcast(ctx context.Context, body string, severity string) (*ChatMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("broadcast alert body cannot be empty")
	}

	if severity == "" {
		severity = "critical"
	}

	msg := NewMessage(m.node.ID().String(), m.cfg.NodeName, body, true, severity)

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize broadcast message: %w", err)
	}

	// Publish to alert topic
	if err := m.node.PubSub().Publish(ctx, m.cfg.AlertTopic, data); err != nil {
		return nil, fmt.Errorf("failed to publish to alert topic: %w", err)
	}

	// Also publish to chat topic so it appears in standard conversational timeline
	_ = m.node.PubSub().Publish(ctx, m.cfg.ChatTopic, data)

	// Record in memory store and persistent storage
	m.store.Add(msg)
	if m.persister != nil {
		if err := m.persister.SaveMessage(ctx, msg); err != nil {
			log.Printf("[Broadcast] Failed to persist sent emergency message %s: %v", msg.ID, err)
		}
	}
	log.Printf("[Broadcast] *** EMERGENCY ALERT DISPATCHED: %s (Severity: %s) ***", body, severity)
	return msg, nil
}

// SetPersister sets or updates the persistent storage provider.
func (m *Manager) SetPersister(p Persister) {
	m.persister = p
}

// GetMessages retrieves recent messages, prioritizing persistent storage if configured.
func (m *Manager) GetMessages(ctx context.Context, limit, offset int) ([]*ChatMessage, error) {
	if m.persister != nil {
		return m.persister.GetMessages(ctx, limit, offset)
	}
	return m.store.ListRecent(limit), nil
}

// Store returns the in-memory message store reference.
func (m *Manager) Store() *MessageStore {
	return m.store
}

// Close unsubscribes and terminates listeners.
func (m *Manager) Close() {
	m.cancel()
	if m.chatSub != nil {
		m.chatSub.Cancel()
	}
	if m.alertSub != nil {
		m.alertSub.Cancel()
	}
}
