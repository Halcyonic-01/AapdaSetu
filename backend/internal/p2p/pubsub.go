package p2p

import (
	"context"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// PubSubManager wraps GossipSub topics and subscriptions for mesh communication.
type PubSubManager struct {
	ps     *pubsub.PubSub
	mu     sync.RWMutex
	topics map[string]*pubsub.Topic
	subs   map[string]*pubsub.Subscription
}

// NewPubSubManager initializes a GossipSub pubsub router on the given host.
func NewPubSubManager(ctx context.Context, h host.Host) (*PubSubManager, error) {
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GossipSub: %w", err)
	}

	return &PubSubManager{
		ps:     ps,
		topics: make(map[string]*pubsub.Topic),
		subs:   make(map[string]*pubsub.Subscription),
	}, nil
}

// JoinTopic joins a PubSub topic if not already joined.
func (m *PubSubManager) JoinTopic(topicName string) (*pubsub.Topic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t, exists := m.topics[topicName]; exists {
		return t, nil
	}

	topic, err := m.ps.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic %s: %w", topicName, err)
	}

	m.topics[topicName] = topic
	return topic, nil
}

// Subscribe subscribes to a topic and returns the subscription handle.
func (m *PubSubManager) Subscribe(topicName string) (*pubsub.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub, exists := m.subs[topicName]; exists {
		return sub, nil
	}

	topic, exists := m.topics[topicName]
	if !exists {
		var err error
		topic, err = m.ps.Join(topicName)
		if err != nil {
			return nil, fmt.Errorf("failed to join topic %s before subscribing: %w", topicName, err)
		}
		m.topics[topicName] = topic
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topicName, err)
	}

	m.subs[topicName] = sub
	return sub, nil
}

// Publish broadcasts raw byte payload to all peers subscribed to the topic.
func (m *PubSubManager) Publish(ctx context.Context, topicName string, data []byte) error {
	m.mu.RLock()
	topic, exists := m.topics[topicName]
	m.mu.RUnlock()

	if !exists {
		var err error
		topic, err = m.JoinTopic(topicName)
		if err != nil {
			return err
		}
	}

	return topic.Publish(ctx, data)
}

// ListPeers returns the IDs of all peers known to be subscribed to the topic.
func (m *PubSubManager) ListPeers(topicName string) []peer.ID {
	m.mu.RLock()
	topic, exists := m.topics[topicName]
	m.mu.RUnlock()

	if !exists {
		return nil
	}
	return topic.ListPeers()
}

// Close closes all topics and subscriptions.
func (m *PubSubManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sub := range m.subs {
		sub.Cancel()
	}
	for _, topic := range m.topics {
		topic.Close()
	}
}
