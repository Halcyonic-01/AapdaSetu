export interface NodeStatus {
  node_name: string;
  status: string;
  peer_id?: string;
  addresses?: string[];
  p2p_port: number;
  http_port: number;
  uptime: string;
  peer_count: number;
  rendezvous?: string;
  timestamp: string;
}

export interface Peer {
  id: string;
  name?: string;
  addresses: string[];
  connected: boolean;
  last_seen: string;
}

export interface ChatMessage {
  id: string;
  sender_id: string;
  sender_name: string;
  body: string;
  timestamp: string;
  emergency: boolean;
  severity?: 'critical' | 'warning' | 'info';
}
