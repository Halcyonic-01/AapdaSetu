package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds runtime configuration for an AapdaSetu node.
type Config struct {
	HTTPPort    int    // Port for the local REST / WebSocket API
	P2PPort     int    // Port for libp2p listen address
	Rendezvous  string // mDNS rendezvous string for peer discovery
	ChatTopic   string // PubSub topic name for standard chat
	AlertTopic  string // PubSub topic name for emergency broadcast
	DataDir       string // Directory for storing node state / logs
	NodeName      string // Human-readable friendly node name
	BootstrapPeer string // Optional initial peer multiaddr to connect to on start
}

// LoadConfig parses command-line flags and environment variables.
func LoadConfig() *Config {
	cfg := &Config{}

	flag.IntVar(&cfg.HTTPPort, "http-port", getEnvAsInt("HTTP_PORT", 8080), "HTTP API port for frontend")
	flag.IntVar(&cfg.P2PPort, "p2p-port", getEnvAsInt("P2P_PORT", 9000), "libp2p listen port")
	flag.StringVar(&cfg.Rendezvous, "rendezvous", getEnv("MDNS_RENDEZVOUS", "aapdasetu-emergency-net"), "mDNS rendezvous discovery string")
	flag.StringVar(&cfg.ChatTopic, "chat-topic", getEnv("CHAT_TOPIC", "aapdasetu-chat"), "PubSub topic for chat messages")
	flag.StringVar(&cfg.AlertTopic, "alert-topic", getEnv("ALERT_TOPIC", "aapdasetu-alert"), "PubSub topic for emergency broadcasts")
	flag.StringVar(&cfg.DataDir, "data-dir", getEnv("DATA_DIR", "./data"), "Path for local node storage")
	flag.StringVar(&cfg.NodeName, "node-name", getEnv("NODE_NAME", "AapdaSetu-Node"), "Friendly display name for this node")
	flag.StringVar(&cfg.BootstrapPeer, "peer", getEnv("BOOTSTRAP_PEER", ""), "Optional initial peer multiaddr to connect to on start")

	flag.Parse()
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
