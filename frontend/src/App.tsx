import { useState, useEffect } from 'react';
import { Radio, AlertTriangle, Users, MessageSquare, ShieldAlert, Wifi, RefreshCw } from 'lucide-react';
import { apiService } from './services/api';
import { NodeStatus } from './types';

export default function App() {
  const [nodeStatus, setNodeStatus] = useState<NodeStatus | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = async () => {
    try {
      setLoading(true);
      const status = await apiService.getStatus();
      setNodeStatus(status);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Unable to connect to local AapdaSetu node');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 10000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="flex h-screen bg-slate-950 text-slate-100 antialiased font-sans">
      {/* Sidebar */}
      <aside className="w-80 border-r border-slate-800 bg-slate-900/60 flex flex-col backdrop-blur">
        {/* Brand Header */}
        <div className="p-4 border-b border-slate-800 flex items-center gap-3">
          <div className="p-2.5 rounded-xl bg-red-600/20 text-red-500 border border-red-500/30">
            <Radio className="w-6 h-6 animate-pulse" />
          </div>
          <div>
            <h1 className="font-bold text-lg text-slate-50 tracking-tight">AapdaSetu</h1>
            <p className="text-xs text-slate-400 font-medium">Offline Emergency Mesh</p>
          </div>
        </div>

        {/* Node Status Card */}
        <div className="p-4 m-3 rounded-xl bg-slate-800/50 border border-slate-700/60 text-xs space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-slate-400 font-semibold uppercase tracking-wider text-[10px]">Local Node</span>
            <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping" />
              {nodeStatus?.status || 'Connecting...'}
            </span>
          </div>
          <div className="text-sm font-semibold text-slate-200 truncate">
            {nodeStatus?.node_name || 'AapdaSetu Node'}
          </div>
          <div className="pt-2 border-t border-slate-700/40 grid grid-cols-2 gap-2 text-slate-400">
            <div>
              <span className="block text-[10px] text-slate-500">P2P Port</span>
              <span className="text-slate-300 font-mono">{nodeStatus?.p2p_port || 9000}</span>
            </div>
            <div>
              <span className="block text-[10px] text-slate-500">Uptime</span>
              <span className="text-slate-300 font-mono">{nodeStatus?.uptime || '0s'}</span>
            </div>
          </div>
        </div>

        {/* Navigation Tabs */}
        <div className="px-3 space-y-1">
          <button className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg bg-red-500/10 text-red-400 border border-red-500/20 text-sm font-medium">
            <MessageSquare className="w-4 h-4" />
            <span>Emergency Mesh Chat</span>
          </button>
          <button className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-slate-400 hover:bg-slate-800/60 hover:text-slate-200 text-sm font-medium transition">
            <Users className="w-4 h-4" />
            <span>Connected Peers ({nodeStatus?.peer_count ?? 0})</span>
          </button>
        </div>

        {/* Peer List Preview */}
        <div className="flex-1 overflow-y-auto px-4 py-3">
          <div className="flex items-center justify-between text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
            <span>Nearby Peers</span>
            <button onClick={fetchStatus} title="Refresh" className="text-slate-500 hover:text-slate-300">
              <RefreshCw className={`w-3 h-3 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
          <div className="text-center py-8 text-slate-500 text-xs">
            <Wifi className="w-6 h-6 mx-auto mb-2 opacity-40" />
            <p>Scanning local network via mDNS...</p>
            <p className="text-[11px] text-slate-600 mt-1">Peers on the same WiFi will appear automatically</p>
          </div>
        </div>

        {/* Footer info */}
        <div className="p-3 border-t border-slate-800 text-[11px] text-slate-500 flex items-center justify-between">
          <span>AapdaSetu v0.1 (Offline-first)</span>
          <span className="text-emerald-500">P2P Mesh</span>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col bg-slate-950 overflow-hidden">
        {/* Top Notification / Emergency Banner Bar */}
        <header className="h-16 border-b border-slate-800 px-6 flex items-center justify-between bg-slate-900/30">
          <div className="flex items-center gap-3">
            <span className="inline-flex items-center gap-2 px-3 py-1 rounded-md bg-amber-500/10 border border-amber-500/20 text-amber-400 text-xs font-medium">
              <AlertTriangle className="w-3.5 h-3.5" />
              Internet-Free Mesh Active
            </span>
            {error && (
              <span className="text-xs text-red-400 bg-red-950/40 px-2 py-0.5 rounded border border-red-800/40">
                Backend connection: {error}
              </span>
            )}
          </div>

          <div className="flex items-center gap-3">
            <button className="flex items-center gap-2 px-3.5 py-1.5 rounded-lg bg-red-600 hover:bg-red-500 text-white text-xs font-semibold shadow-lg shadow-red-600/20 transition cursor-pointer">
              <ShieldAlert className="w-4 h-4" />
              Broadcast SOS
            </button>
          </div>
        </header>

        {/* Chat / Content View */}
        <div className="flex-1 flex flex-col justify-between p-6">
          <div className="flex-1 flex flex-col items-center justify-center text-center max-w-md mx-auto">
            <div className="p-4 rounded-2xl bg-slate-900 border border-slate-800 mb-4 text-slate-400 shadow-inner">
              <Radio className="w-10 h-10 text-red-500 mx-auto animate-pulse" />
            </div>
            <h2 className="text-xl font-bold text-slate-100 mb-1">AapdaSetu Emergency Mesh Network</h2>
            <p className="text-sm text-slate-400 mb-6 leading-relaxed">
              Decentralized peer-to-peer communication platform. Broadcast SOS alerts and coordinate rescue operations with nearby devices without internet access.
            </p>
            <div className="grid grid-cols-2 gap-3 w-full text-left text-xs">
              <div className="p-3 rounded-lg bg-slate-900/80 border border-slate-800">
                <span className="font-semibold text-slate-300 block mb-1">mDNS Auto-Discovery</span>
                <span className="text-slate-500">Detects nearby nodes on the local WiFi automatically.</span>
              </div>
              <div className="p-3 rounded-lg bg-slate-900/80 border border-slate-800">
                <span className="font-semibold text-slate-300 block mb-1">libp2p GossipSub</span>
                <span className="text-slate-500">Reliable multi-hop mesh broadcast & chat.</span>
              </div>
            </div>
          </div>

          {/* Chat Input Placeholder */}
          <div className="mt-4 pt-4 border-t border-slate-800">
            <div className="flex gap-2">
              <input
                type="text"
                placeholder="Type an emergency message to nearby peers..."
                className="flex-1 bg-slate-900 border border-slate-800 rounded-xl px-4 py-2.5 text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-red-500/50"
                disabled
              />
              <button
                disabled
                className="px-5 py-2.5 rounded-xl bg-slate-800 text-slate-500 text-sm font-semibold cursor-not-allowed"
              >
                Send
              </button>
            </div>
            <p className="text-[11px] text-slate-500 mt-2 text-center">
              Phase 1 App Shell loaded. Full GossipSub messaging will activate in Phase 2 & 3.
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}
