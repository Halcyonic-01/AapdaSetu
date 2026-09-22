export interface NodeStatus {
  node_name: string;
  status: string;
  p2p_port: number;
  http_port: number;
  uptime: string;
  peer_count: number;
  timestamp: string;
}

export interface Peer {
  id: string;
  name?: string;
  address: string;
  connected_at: string;
  status: 'online' | 'offline';
}

export interface ChatMessage {
  id: string;
  from: string;
  from_name?: string;
  body: string;
  timestamp: string;
  is_self?: boolean;
}

export interface BroadcastMessage extends ChatMessage {
  emergency: boolean;
  severity?: 'critical' | 'warning' | 'info';
}
