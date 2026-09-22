package broadcast

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/p2p"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
)

func TestMessageStore(t *testing.T) {
	store := NewMessageStore(5)

	m1 := NewMessage("peer1", "Node1", "Message 1", false, "info")
	m2 := NewMessage("peer2", "Node2", "Message 2", true, "critical")

	store.Add(m1)
	store.Add(m2)
	// Duplicate ID shouldn't add twice
	store.Add(m1)

	if len(store.List()) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(store.List()))
	}

	recent := store.ListRecent(1)
	if len(recent) != 1 || recent[0].ID != m2.ID {
		t.Fatalf("expected recent message to be m2")
	}
}

func TestBroadcastMultiNodeChatAndAlert(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cfg1 := &config.Config{
		P2PPort:    9881,
		HTTPPort:   8881,
		Rendezvous: "aapdasetu-broadcast-test",
		ChatTopic:  "test-chat-topic",
		AlertTopic: "test-alert-topic",
		NodeName:   "Alpha",
	}
	cfg2 := &config.Config{
		P2PPort:    9882,
		HTTPPort:   8882,
		Rendezvous: "aapdasetu-broadcast-test",
		ChatTopic:  "test-chat-topic",
		AlertTopic: "test-alert-topic",
		NodeName:   "Bravo",
	}

	store1 := peers.NewStore()
	store2 := peers.NewStore()

	node1, err := p2p.NewNode(ctx, cfg1, store1)
	if err != nil {
		t.Fatalf("failed to create node1: %v", err)
	}
	defer node1.Close()

	node2, err := p2p.NewNode(ctx, cfg2, store2)
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}
	defer node2.Close()

	// Connect nodes
	addr1 := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", cfg1.P2PPort, node1.ID())
	if err := node2.ConnectDirect(ctx, addr1); err != nil {
		t.Fatalf("failed to connect node2 to node1: %v", err)
	}

	msgStore1 := NewMessageStore(100)
	msgStore2 := NewMessageStore(100)

	mgr1, err := NewManager(ctx, node1, cfg1, msgStore1)
	if err != nil {
		t.Fatalf("failed to create mgr1: %v", err)
	}
	defer mgr1.Close()

	mgr2, err := NewManager(ctx, node2, cfg2, msgStore2)
	if err != nil {
		t.Fatalf("failed to create mgr2: %v", err)
	}
	defer mgr2.Close()

	// Wait for GossipSub heartbeat
	time.Sleep(1 * time.Second)

	// Node 1 sends a chat message
	chatPayload := "Need status report from Sector B"
	if _, err := mgr1.SendChat(ctx, chatPayload); err != nil {
		t.Fatalf("failed to send chat: %v", err)
	}

	// Verify Node 2 received it
	var receivedChat *ChatMessage
	for i := 0; i < 20; i++ {
		list := msgStore2.List()
		for _, m := range list {
			if m.Body == chatPayload {
				receivedChat = m
				break
			}
		}
		if receivedChat != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if receivedChat == nil {
		t.Fatalf("node 2 failed to receive chat message from node 1")
	}

	// Node 2 sends an Emergency Broadcast
	alertPayload := "FLASH FLOOD WARNING - MOVE TO HIGH GROUND"
	if _, err := mgr2.SendBroadcast(ctx, alertPayload, "critical"); err != nil {
		t.Fatalf("failed to send emergency broadcast: %v", err)
	}

	// Verify Node 1 received it
	var receivedAlert *ChatMessage
	for i := 0; i < 20; i++ {
		list := msgStore1.List()
		for _, m := range list {
			if m.Emergency && m.Body == alertPayload {
				receivedAlert = m
				break
			}
		}
		if receivedAlert != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if receivedAlert == nil {
		t.Fatalf("node 1 failed to receive emergency broadcast from node 2")
	}
}

type mockPersister struct {
	sync.Mutex
	saved []*ChatMessage
}

func (m *mockPersister) SaveMessage(ctx context.Context, msg *ChatMessage) error {
	m.Lock()
	defer m.Unlock()
	m.saved = append(m.saved, msg)
	return nil
}

func (m *mockPersister) GetMessages(ctx context.Context, limit, offset int) ([]*ChatMessage, error) {
	m.Lock()
	defer m.Unlock()
	if offset >= len(m.saved) {
		return []*ChatMessage{}, nil
	}
	end := offset + limit
	if end > len(m.saved) {
		end = len(m.saved)
	}
	return m.saved[offset:end], nil
}

func TestBroadcastWithPersister(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cfg1 := &config.Config{
		P2PPort:    9891,
		HTTPPort:   8891,
		Rendezvous: "aapdasetu-persister-test",
		ChatTopic:  "test-chat-persister",
		AlertTopic: "test-alert-persister",
		NodeName:   "Persist-Node-1",
	}
	cfg2 := &config.Config{
		P2PPort:    9892,
		HTTPPort:   8892,
		Rendezvous: "aapdasetu-persister-test",
		ChatTopic:  "test-chat-persister",
		AlertTopic: "test-alert-persister",
		NodeName:   "Persist-Node-2",
	}

	store1 := peers.NewStore()
	store2 := peers.NewStore()

	node1, err := p2p.NewNode(ctx, cfg1, store1)
	if err != nil {
		t.Fatalf("failed to create node1: %v", err)
	}
	defer node1.Close()

	node2, err := p2p.NewNode(ctx, cfg2, store2)
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}
	defer node2.Close()

	// Connect nodes directly
	addr1 := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", cfg1.P2PPort, node1.ID())
	if err := node2.ConnectDirect(ctx, addr1); err != nil {
		t.Fatalf("failed to connect node2 to node1: %v", err)
	}

	persister1 := &mockPersister{}
	persister2 := &mockPersister{}

	mgr1, err := NewManager(ctx, node1, cfg1, nil)
	if err != nil {
		t.Fatalf("failed to create mgr1: %v", err)
	}
	defer mgr1.Close()
	mgr1.SetPersister(persister1)

	mgr2, err := NewManager(ctx, node2, cfg2, nil)
	if err != nil {
		t.Fatalf("failed to create mgr2: %v", err)
	}
	defer mgr2.Close()
	mgr2.SetPersister(persister2)

	// Wait for pubsub mesh to connect
	time.Sleep(1 * time.Second)

	// Node 1 sends chat message -> should be persisted outbound on Node 1, and inbound on Node 2
	chatMsg, err := mgr1.SendChat(ctx, "Test persist message")
	if err != nil {
		t.Fatalf("failed to send chat: %v", err)
	}

	// Verify mgr1 saved outbound
	msgs1, err := mgr1.GetMessages(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to get messages from mgr1: %v", err)
	}
	if len(msgs1) != 1 || msgs1[0].ID != chatMsg.ID {
		t.Fatalf("expected outbound message persisted on mgr1, got %v", msgs1)
	}

	// Verify mgr2 saved inbound
	var foundOnNode2 bool
	for i := 0; i < 25; i++ {
		msgs2, _ := mgr2.GetMessages(ctx, 10, 0)
		for _, m := range msgs2 {
			if m.ID == chatMsg.ID {
				foundOnNode2 = true
				break
			}
		}
		if foundOnNode2 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !foundOnNode2 {
		t.Fatalf("expected inbound message to be saved by persister on mgr2")
	}
}

