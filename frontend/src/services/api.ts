import { NodeStatus, Peer, ChatMessage } from '../types';

const API_BASE = '/api';

export const apiService = {
  async getHealth(): Promise<{ status: string; time: string }> {
    const res = await fetch(`${API_BASE}/health`);
    if (!res.ok) throw new Error('Failed to fetch health');
    return res.json();
  },

  async getStatus(): Promise<NodeStatus> {
    const res = await fetch(`${API_BASE}/status`);
    if (!res.ok) throw new Error('Failed to fetch node status');
    return res.json();
  },

  async getPeers(): Promise<Peer[]> {
    const res = await fetch(`${API_BASE}/peers`);
    if (!res.ok) return [];
    return res.json();
  },

  async getMessages(limit: number = 100): Promise<ChatMessage[]> {
    const res = await fetch(`${API_BASE}/chat/messages?limit=${limit}`);
    if (!res.ok) return [];
    return res.json();
  },

  async sendMessage(body: string): Promise<ChatMessage> {
    const res = await fetch(`${API_BASE}/chat/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ body }),
    });
    if (!res.ok) {
      const errData = await res.text();
      throw new Error(errData || 'Failed to send message');
    }
    return res.json();
  },

  async sendBroadcast(body: string, severity: 'critical' | 'warning' | 'info' = 'critical'): Promise<ChatMessage> {
    const res = await fetch(`${API_BASE}/broadcast/send`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ body, severity }),
    });
    if (!res.ok) {
      const errData = await res.text();
      throw new Error(errData || 'Failed to send broadcast');
    }
    return res.json();
  },
};
