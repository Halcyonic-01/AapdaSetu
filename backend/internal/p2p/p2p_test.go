package p2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Halcyonic-01/AapdaSetu/backend/config"
	"github.com/Halcyonic-01/AapdaSetu/backend/internal/peers"
)

func TestP2PNodeCreationAndPubSub(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	store1 := peers.NewStore()
	store2 := peers.NewStore()

	cfg1 := &config.Config{
		P2PPort:    9871,
		HTTPPort:   8871,
		Rendezvous: "aapdasetu-test-net",
		NodeName:   "TestNode-1",
	}

	cfg2 := &config.Config{
		P2PPort:    9872,
		HTTPPort:   8872,
		Rendezvous: "aapdasetu-test-net",
		NodeName:   "TestNode-2",
	}

	node1, err := NewNode(ctx, cfg1, store1)
	if err != nil {
		t.Fatalf("failed to create node1: %v", err)
	}
	defer node1.Close()

	node2, err := NewNode(ctx, cfg2, store2)
	if err != nil {
		t.Fatalf("failed to create node2: %v", err)
	}
	defer node2.Close()

	// Connect node2 to node1 directly using multiaddr
	addr1 := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", cfg1.P2PPort, node1.ID())
	if err := node2.ConnectDirect(ctx, addr1); err != nil {
		t.Fatalf("failed to connect node2 to node1: %v", err)
	}

	// Give a moment for connection events to register in both stores
	time.Sleep(500 * time.Millisecond)

	if store1.Count() != 1 {
		t.Errorf("expected node1 to have 1 connected peer, got %d", store1.Count())
	}
	if store2.Count() != 1 {
		t.Errorf("expected node2 to have 1 connected peer, got %d", store2.Count())
	}

	// Test PubSub message dissemination between the two nodes
	topicName := "emergency-test-channel"
	sub1, err := node1.PubSub().Subscribe(topicName)
	if err != nil {
		t.Fatalf("node1 failed to subscribe: %v", err)
	}

	sub2, err := node2.PubSub().Subscribe(topicName)
	if err != nil {
		t.Fatalf("node2 failed to subscribe: %v", err)
	}
	_ = sub2

	// Wait briefly for GossipSub mesh link formation
	time.Sleep(1 * time.Second)

	testMessage := []byte("EVACUATION NOTICE: Sector 4")
	if err := node2.PubSub().Publish(ctx, topicName, testMessage); err != nil {
		t.Fatalf("failed to publish message: %v", err)
	}

	// Read message on node1
	msgChan := make(chan []byte, 1)
	go func() {
		msg, err := sub1.Next(ctx)
		if err == nil {
			msgChan <- msg.Data
		}
	}()

	select {
	case received := <-msgChan:
		if string(received) != string(testMessage) {
			t.Errorf("expected %q, got %q", string(testMessage), string(received))
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for pubsub message on node1")
	}
}
