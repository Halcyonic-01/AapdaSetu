import React, { useState, useEffect, useRef } from 'react';
import {
  Radio,
  Send,
  AlertTriangle,
  Settings,
  X,
  Copy,
  Check,
  Volume2,
  VolumeX,
  Server,
  RefreshCw,
  Clock,
  Shield,
  FileText,
  Activity,
  Layers,
} from 'lucide-react';
import { apiService } from './services/api';
import { NodeStatus, Peer, ChatMessage } from './types';

// Web Audio API beep for emergency bulletins (100% offline)
function playAlertBeep() {
  try {
    const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
    if (!AudioCtx) return;
    const ctx = new AudioCtx();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = 'sawtooth';
    osc.frequency.setValueAtTime(800, ctx.currentTime);
    osc.frequency.setValueAtTime(600, ctx.currentTime + 0.15);

    gain.gain.setValueAtTime(0.18, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.3);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start();
    osc.stop(ctx.currentTime + 0.3);
  } catch {
    // Audio autoplay restrictions
  }
}

export default function App() {
  const [nodeStatus, setNodeStatus] = useState<NodeStatus | null>(null);
  const [peers, setPeers] = useState<Peer[]>([]);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputMessage, setInputMessage] = useState<string>('');
  const [soundEnabled, setSoundEnabled] = useState<boolean>(true);
  const [sending, setSending] = useState<boolean>(false);
  const [copiedId, setCopiedId] = useState<boolean>(false);
  const [refreshing, setRefreshing] = useState<boolean>(false);

  // Emergency Broadcast Modal
  const [broadcastModalOpen, setBroadcastModalOpen] = useState<boolean>(false);
  const [broadcastText, setBroadcastText] = useState<string>('');
  const [broadcastSeverity, setBroadcastSeverity] = useState<'critical' | 'warning' | 'info'>('critical');
  const [broadcastSending, setBroadcastSending] = useState<boolean>(false);

  // Settings Modal
  const [settingsOpen, setSettingsOpen] = useState<boolean>(false);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const prevMsgCount = useRef<number>(0);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const fetchData = async () => {
    try {
      setRefreshing(true);
      const [status, peerList, msgList] = await Promise.all([
        apiService.getStatus(),
        apiService.getPeers(),
        apiService.getMessages(200),
      ]);

      setNodeStatus(status);
      setPeers(peerList || []);

      if (msgList) {
        setMessages(msgList);

        // Sound alert for incoming emergency broadcast from other nodes
        const emergencyMsgs = msgList.filter((m) => m.emergency);
        if (emergencyMsgs.length > 0) {
          const newest = emergencyMsgs[emergencyMsgs.length - 1];
          if (msgList.length > prevMsgCount.current && newest.sender_id !== status.peer_id) {
            if (soundEnabled) playAlertBeep();
          }
        }
        prevMsgCount.current = msgList.length;
      }
    } catch {
      // Quiet background polling
    } finally {
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 2000);
    return () => clearInterval(interval);
  }, [soundEnabled]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const copyMyId = () => {
    if (nodeStatus?.peer_id) {
      navigator.clipboard.writeText(nodeStatus.peer_id);
      setCopiedId(true);
      setTimeout(() => setCopiedId(false), 2000);
    }
  };

  const handleSendMessage = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const trimmed = inputMessage.trim();
    if (!trimmed || sending) return;

    try {
      setSending(true);
      const newMsg = await apiService.sendMessage(trimmed);
      setMessages((prev) => [...prev, newMsg]);
      setInputMessage('');
      setTimeout(scrollToBottom, 40);
    } catch (err: any) {
      alert(`Dispatch error: ${err.message}`);
    } finally {
      setSending(false);
    }
  };

  const handleSendBroadcast = async () => {
    const trimmed = broadcastText.trim();
    if (!trimmed || broadcastSending) return;

    try {
      setBroadcastSending(true);
      const newAlert = await apiService.sendBroadcast(trimmed, broadcastSeverity);
      setMessages((prev) => [...prev, newAlert]);
      setBroadcastModalOpen(false);
      setBroadcastText('');
      setTimeout(scrollToBottom, 40);
    } catch (err: any) {
      alert(`Emergency transmission error: ${err.message}`);
    } finally {
      setBroadcastSending(false);
    }
  };

  const connectedPeers = peers.filter((p) => p.connected);

  const emergencyTemplates = [
    { label: 'Medical Assistance', text: 'URGENT: Immediate medical assistance required at this location. Paramedics / first-aid supplies needed.', severity: 'critical' as const },
    { label: 'Search & Rescue', text: 'CRITICAL: People trapped under debris / collapsed structure. Search and rescue team required.', severity: 'critical' as const },
    { label: 'Evacuation Notice', text: 'EVACUATION ADVISORY: Hazard / water level rising. Move to designated high ground immediately.', severity: 'critical' as const },
    { label: 'Critical Supplies Depleted', text: 'LOGISTICS REQUEST: Clean drinking water and emergency food rations depleted.', severity: 'warning' as const },
  ];

  return (
    <div className="flex flex-col h-screen w-screen bg-[#f0f2f5] text-[#1c2d42] font-sans overflow-hidden">
      {/* GOVERNMENT OFFICIAL TOP BANNER */}
      <div className="bg-[#0b2545] text-white border-b-2 border-[#d97706] flex-shrink-0">
        {/* Tricolor Institutional Accent Line */}
        <div className="h-1 w-full flex">
          <div className="h-full flex-1 bg-[#ff9933]" />
          <div className="h-full flex-1 bg-[#ffffff]" />
          <div className="h-full flex-1 bg-[#138808]" />
        </div>

        <div className="px-6 py-2.5 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded bg-[#133e6f] border border-[#235896] flex items-center justify-center font-bold text-white shadow-sm">
              <Shield className="w-5 h-5 text-[#d97706]" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-extrabold text-sm tracking-wide uppercase text-white">
                  AapdaSetu | आपदा सेतु
                </span>
                <span className="text-[10px] bg-[#d97706] text-slate-900 font-bold px-1.5 py-0.2 rounded uppercase">
                  Emergency Mesh
                </span>
              </div>
              <p className="text-[11px] text-[#93c5fd]">
                National Disaster Emergency Communication Portal • Offline P2P Mesh Protocol
              </p>
            </div>
          </div>

          <div className="flex items-center gap-4 text-xs">
            <div className="hidden sm:flex items-center gap-2 bg-[#133e6f]/80 px-3 py-1 rounded border border-[#235896] text-[11px]">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
              <span className="text-emerald-300 font-semibold uppercase">Network: Active (LAN)</span>
              <span className="text-slate-400">|</span>
              <span className="text-slate-300">Zero Internet Required</span>
            </div>

            <button
              onClick={() => setSoundEnabled(!soundEnabled)}
              className="p-1.5 rounded hover:bg-[#133e6f] text-slate-300 hover:text-white transition"
              title={soundEnabled ? 'Alert audio enabled' : 'Alert audio muted'}
            >
              {soundEnabled ? <Volume2 className="w-4 h-4 text-emerald-400" /> : <VolumeX className="w-4 h-4 text-slate-400" />}
            </button>

            <button
              onClick={() => setSettingsOpen(true)}
              className="p-1.5 rounded hover:bg-[#133e6f] text-slate-300 hover:text-white transition"
              title="Terminal Information"
            >
              <Settings className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* WORKSPACE AREA */}
      <div className="flex-1 flex overflow-hidden">
        {/* LEFT CONTROL & DISPATCH PANEL */}
        <aside className="w-72 md:w-80 bg-white border-r border-[#cbd5e1] flex flex-col flex-shrink-0 select-none">
          {/* Emergency Broadcast Dispatch Button */}
          <div className="p-3.5 border-b border-[#cbd5e1] bg-[#fff5f5]">
            <button
              onClick={() => setBroadcastModalOpen(true)}
              className="w-full py-2.5 px-3 bg-[#b91c1c] hover:bg-[#991b1b] text-white text-xs font-bold uppercase tracking-wider rounded border border-[#7f1d1d] shadow-sm flex items-center justify-center gap-2 transition cursor-pointer"
            >
              <AlertTriangle className="w-4 h-4 text-amber-300 flex-shrink-0" />
              <span>Emergency Broadcast</span>
            </button>
            <p className="text-[10px] text-[#991b1b] font-medium text-center mt-1.5">
              Transmits high-priority bulletin to all local nodes
            </p>
          </div>

          {/* Station Status Block */}
          <div className="p-3.5 border-b border-[#cbd5e1] bg-[#f8fafc] text-xs space-y-2">
            <div className="flex items-center justify-between font-bold text-[#0b2545] uppercase text-[11px]">
              <span className="flex items-center gap-1.5">
                <Activity className="w-3.5 h-3.5 text-[#0b2545]" />
                Station Status
              </span>
              <span className="text-[10px] bg-emerald-100 text-emerald-800 font-bold px-1.5 py-0.5 rounded border border-emerald-300">
                OPERATIONAL
              </span>
            </div>

            <div className="bg-white p-2.5 rounded border border-[#cbd5e1] space-y-1 text-[11px]">
              <div className="flex justify-between">
                <span className="text-slate-500">Terminal Name:</span>
                <span className="font-semibold text-slate-800 truncate max-w-[130px]">
                  {nodeStatus?.node_name || 'AapdaSetu-Node'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Listen Port:</span>
                <span className="font-mono text-slate-800">{nodeStatus?.p2p_port || 9000}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Uptime:</span>
                <span className="font-mono text-slate-800">{nodeStatus?.uptime || '0s'}</span>
              </div>
              <div className="pt-1 border-t border-slate-100 flex items-center justify-between">
                <span className="text-slate-500">Node ID:</span>
                <button
                  onClick={copyMyId}
                  className="font-mono text-[10px] text-blue-700 hover:underline flex items-center gap-1"
                  title="Copy full Node ID"
                >
                  <span>{nodeStatus?.peer_id ? `${nodeStatus.peer_id.slice(0, 6)}...` : 'Generating'}</span>
                  {copiedId ? <Check className="w-3 h-3 text-emerald-600" /> : <Copy className="w-3 h-3 text-slate-400" />}
                </button>
              </div>
            </div>
          </div>

          {/* Connected Terminals (Nearby Devices) */}
          <div className="flex-1 flex flex-col min-h-0">
            <div className="px-3.5 py-2.5 bg-[#f1f5f9] border-b border-[#cbd5e1] flex items-center justify-between text-xs">
              <span className="font-bold text-[#0b2545] uppercase text-[11px] flex items-center gap-1.5">
                <Server className="w-3.5 h-3.5 text-slate-600" />
                Connected Terminals
              </span>
              <span className="font-mono text-[11px] font-bold text-slate-600 bg-white px-1.5 py-0.5 rounded border border-[#cbd5e1]">
                {connectedPeers.length} Active
              </span>
            </div>

            <div className="flex-1 overflow-y-auto p-2.5 space-y-1.5">
              {peers.length === 0 ? (
                <div className="p-6 text-center text-xs text-slate-500 space-y-2">
                  <Radio className="w-6 h-6 mx-auto text-slate-400 animate-pulse" />
                  <p className="font-semibold text-slate-700">Scanning Local Mesh</p>
                  <p className="text-[11px] text-slate-500">
                    Listening for AapdaSetu stations on local Wi-Fi / hotspot...
                  </p>
                </div>
              ) : (
                peers.map((peer) => (
                  <div
                    key={peer.id}
                    className="p-2.5 rounded bg-white border border-[#cbd5e1] hover:border-slate-400 text-xs transition"
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-slate-800 font-mono text-[11px] truncate max-w-[130px]" title={peer.id}>
                        {peer.name || `STATION-${peer.id.slice(0, 6).toUpperCase()}`}
                      </span>
                      <span
                        className={`text-[9px] font-bold px-1.5 py-0.2 rounded uppercase ${
                          peer.connected
                            ? 'bg-emerald-100 text-emerald-800 border border-emerald-300'
                            : 'bg-slate-100 text-slate-600 border border-slate-300'
                        }`}
                      >
                        {peer.connected ? 'ONLINE' : 'OFFLINE'}
                      </span>
                    </div>
                    {peer.addresses && peer.addresses.length > 0 && (
                      <div className="text-[10px] text-slate-500 font-mono mt-1 truncate">
                        {peer.addresses[0]}
                      </div>
                    )}
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Footer refresh action */}
          <div className="p-2.5 bg-[#f8fafc] border-t border-[#cbd5e1] flex items-center justify-between text-[11px] text-slate-600">
            <span>Protocol: mDNS / GossipSub</span>
            <button
              onClick={fetchData}
              className="text-slate-600 hover:text-slate-900 font-medium flex items-center gap-1 cursor-pointer"
            >
              <RefreshCw className={`w-3 h-3 ${refreshing ? 'animate-spin' : ''}`} />
              <span>Refresh</span>
            </button>
          </div>
        </aside>

        {/* CENTER DISPATCH LOG & MESSAGING LOG */}
        <main className="flex-1 flex flex-col bg-white overflow-hidden">
          {/* Section Header */}
          <div className="px-6 py-2.5 bg-[#f8fafc] border-b border-[#cbd5e1] flex items-center justify-between flex-shrink-0">
            <div>
              <div className="flex items-center gap-2">
                <FileText className="w-4 h-4 text-[#0b2545]" />
                <h2 className="text-xs font-bold uppercase tracking-wider text-[#0b2545]">
                  Incident Communication Log — General Frequency
                </h2>
              </div>
              <p className="text-[11px] text-slate-500 mt-0.5">
                Local Mesh Channel • {connectedPeers.length} active terminal{connectedPeers.length === 1 ? '' : 's'} linked
              </p>
            </div>

            <div className="text-[11px] text-slate-500 font-mono flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5 text-slate-400" />
              <span>LOGGED DISPATCHES: {messages.length}</span>
            </div>
          </div>

          {/* Dispatches Feed */}
          <div className="flex-1 overflow-y-auto p-6 space-y-3 bg-[#f8fafc]">
            {messages.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-center text-slate-500 p-8">
                <div className="w-12 h-12 rounded bg-slate-200 border border-slate-300 flex items-center justify-center text-slate-600 mb-2.5">
                  <Layers className="w-6 h-6" />
                </div>
                <h3 className="font-bold text-sm text-slate-800 uppercase tracking-wide">
                  Communication Log Initialized
                </h3>
                <p className="text-xs text-slate-500 max-w-md mt-1 leading-relaxed">
                  No transmissions recorded in this operational session. All messages dispatched here are relayed instantaneously to all reachable terminals over the local mesh.
                </p>
              </div>
            ) : (
              messages.map((msg, index) => {
                const isSelf = msg.sender_id === nodeStatus?.peer_id;
                const isEmergency = msg.emergency;

                if (isEmergency) {
                  return (
                    <div key={msg.id || index} className="my-2">
                      <div className="rounded border-2 border-[#b91c1c] bg-[#fef2f2] p-4 text-slate-900 shadow-xs space-y-2">
                        <div className="flex items-center justify-between border-b border-[#fca5a5] pb-1.5 text-xs">
                          <div className="flex items-center gap-2 text-[#991b1b] font-extrabold uppercase tracking-wide">
                            <AlertTriangle className="w-4 h-4 text-[#b91c1c]" />
                            <span>PRIORITY BULLETIN • EMERGENCY BROADCAST [{msg.severity?.toUpperCase() || 'CRITICAL'}]</span>
                          </div>
                          <span className="font-mono text-[11px] text-slate-600">
                            {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                          </span>
                        </div>
                        <p className="text-sm font-bold text-[#7f1d1d] leading-relaxed">
                          {msg.body}
                        </p>
                        <div className="flex items-center justify-between text-[11px] text-slate-600 pt-1 border-t border-[#fca5a5]">
                          <span>Originating Station: <strong className="text-slate-800">{msg.sender_name}</strong> {isSelf ? '(This Station)' : ''}</span>
                          <span className="font-mono text-[10px] text-slate-500">ID: {msg.sender_id.slice(0, 12)}</span>
                        </div>
                      </div>
                    </div>
                  );
                }

                return (
                  <div
                    key={msg.id || index}
                    className={`flex flex-col ${isSelf ? 'items-end' : 'items-start'}`}
                  >
                    <div className="flex items-center gap-1.5 text-[10px] font-mono text-slate-500 mb-0.5 px-1 uppercase">
                      <span className="font-bold text-slate-700">
                        {isSelf ? 'This Station' : msg.sender_name}
                      </span>
                      <span>•</span>
                      <span>
                        {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                      </span>
                    </div>
                    <div
                      className={`max-w-xl px-4 py-2.5 rounded border text-xs leading-relaxed shadow-xs ${
                        isSelf
                          ? 'bg-[#0b2545] text-white border-[#0b2545]'
                          : 'bg-white text-slate-900 border-[#cbd5e1]'
                      }`}
                    >
                      {msg.body}
                    </div>
                  </div>
                );
              })
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Bottom Dispatch Input Bar */}
          <div className="p-3 bg-white border-t border-[#cbd5e1] flex-shrink-0">
            <form onSubmit={handleSendMessage} className="flex items-center gap-2">
              <input
                type="text"
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                placeholder="Type dispatch message to broadcast to nearby mesh terminals..."
                disabled={sending}
                className="flex-1 bg-[#f8fafc] border border-[#cbd5e1] rounded px-3.5 py-2 text-xs text-slate-900 placeholder-slate-400 focus:outline-none focus:border-[#0b2545] focus:bg-white"
              />
              <button
                type="submit"
                disabled={!inputMessage.trim() || sending}
                className="px-4 py-2 rounded bg-[#0b2545] hover:bg-[#133e6f] disabled:bg-slate-300 text-white disabled:text-slate-500 text-xs font-bold uppercase tracking-wider transition cursor-pointer disabled:cursor-not-allowed flex items-center gap-1.5"
              >
                <Send className="w-3.5 h-3.5" />
                <span>Transmit</span>
              </button>
            </form>
            <div className="flex items-center justify-between text-[10px] text-slate-500 mt-1 px-1">
              <span>Transmitting on GossipSub channel: <code>aapdasetu-chat</code></span>
              <span>Press [Enter] to send immediately</span>
            </div>
          </div>
        </main>
      </div>

      {/* EMERGENCY BROADCAST CONFIRMATION MODAL */}
      {broadcastModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 animate-in fade-in duration-100">
          <div className="w-full max-w-lg bg-white rounded border-2 border-[#b91c1c] shadow-2xl p-5 space-y-4">
            <div className="flex items-start justify-between border-b border-slate-200 pb-3">
              <div className="flex items-center gap-2.5 text-[#b91c1c]">
                <AlertTriangle className="w-5 h-5 flex-shrink-0" />
                <div>
                  <h3 className="font-extrabold text-sm uppercase tracking-wide text-slate-900">
                    Send Emergency Message to Nearby Devices?
                  </h3>
                  <p className="text-xs text-slate-600 mt-0.5">
                    Warning: This message will be transmitted with priority alert status to all connected stations.
                  </p>
                </div>
              </div>
              <button
                onClick={() => setBroadcastModalOpen(false)}
                className="text-slate-400 hover:text-slate-700 p-1"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Quick Emergency Templates */}
            <div>
              <label className="block text-[11px] font-bold uppercase tracking-wider text-slate-700 mb-1.5">
                Standard Incident Templates
              </label>
              <div className="grid grid-cols-2 gap-1.5">
                {emergencyTemplates.map((template, idx) => (
                  <button
                    key={idx}
                    type="button"
                    onClick={() => {
                      setBroadcastText(template.text);
                      setBroadcastSeverity(template.severity);
                    }}
                    className="p-2 text-left rounded bg-[#f8fafc] hover:bg-[#fee2e2] border border-[#cbd5e1] hover:border-[#ef4444] text-[11px] font-semibold text-slate-800 transition cursor-pointer"
                  >
                    {template.label}
                  </button>
                ))}
              </div>
            </div>

            {/* Emergency message body */}
            <div>
              <label className="block text-[11px] font-bold uppercase tracking-wider text-slate-700 mb-1">
                Emergency Message Description
              </label>
              <textarea
                rows={3}
                value={broadcastText}
                onChange={(e) => setBroadcastText(e.target.value)}
                placeholder="State the nature of the emergency, exact coordinates/location, and immediate required response..."
                className="w-full bg-[#f8fafc] border border-[#cbd5e1] rounded p-2.5 text-xs text-slate-900 placeholder-slate-400 focus:outline-none focus:border-[#b91c1c] focus:bg-white"
              />
            </div>

            {/* Priority selection */}
            <div>
              <label className="block text-[11px] font-bold uppercase tracking-wider text-slate-700 mb-1">
                Priority Classification
              </label>
              <div className="grid grid-cols-3 gap-2">
                {(['critical', 'warning', 'info'] as const).map((sev) => (
                  <button
                    key={sev}
                    type="button"
                    onClick={() => setBroadcastSeverity(sev)}
                    className={`py-1.5 px-2 rounded text-[11px] font-bold uppercase tracking-wider border transition cursor-pointer ${
                      broadcastSeverity === sev
                        ? sev === 'critical'
                          ? 'bg-[#fee2e2] text-[#991b1b] border-[#ef4444]'
                          : sev === 'warning'
                          ? 'bg-[#fef3c7] text-[#92400e] border-[#f59e0b]'
                          : 'bg-[#e0f2fe] text-[#075985] border-[#38bdf8]'
                        : 'bg-white text-slate-600 border-[#cbd5e1] hover:bg-slate-50'
                    }`}
                  >
                    {sev === 'critical' ? 'Level 1 (Critical)' : sev === 'warning' ? 'Level 2 (Warning)' : 'Level 3 (Advisory)'}
                  </button>
                ))}
              </div>
            </div>

            {/* Modal Actions */}
            <div className="flex items-center justify-end gap-2 pt-2 border-t border-slate-200">
              <button
                type="button"
                onClick={() => setBroadcastModalOpen(false)}
                className="px-3.5 py-1.5 rounded border border-[#cbd5e1] text-xs font-semibold text-slate-700 hover:bg-slate-100 transition cursor-pointer"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleSendBroadcast}
                disabled={!broadcastText.trim() || broadcastSending}
                className="px-4 py-1.5 rounded bg-[#b91c1c] hover:bg-[#991b1b] disabled:bg-slate-300 text-white text-xs font-bold uppercase tracking-wider transition cursor-pointer disabled:cursor-not-allowed"
              >
                {broadcastSending ? 'Transmitting...' : 'Dispatch Broadcast'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* SETTINGS / TERMINAL MODAL */}
      {settingsOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 animate-in fade-in duration-100">
          <div className="w-full max-w-sm bg-white rounded border border-[#cbd5e1] shadow-xl p-5 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-200 pb-2.5">
              <h3 className="font-bold text-xs uppercase tracking-wider text-slate-900">
                Terminal Configuration
              </h3>
              <button onClick={() => setSettingsOpen(false)} className="text-slate-400 hover:text-slate-700 p-1">
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-2.5 text-xs">
              <div>
                <label className="block text-slate-500 font-semibold text-[11px] mb-0.5">Terminal Name</label>
                <div className="p-2 rounded bg-slate-50 border border-slate-200 font-medium text-slate-800">
                  {nodeStatus?.node_name || 'AapdaSetu-Node'}
                </div>
              </div>

              <div>
                <label className="block text-slate-500 font-semibold text-[11px] mb-0.5">P2P Network Port</label>
                <div className="p-2 rounded bg-slate-50 border border-slate-200 font-mono text-slate-800">
                  {nodeStatus?.p2p_port || 9000}
                </div>
              </div>

              <div>
                <label className="block text-slate-500 font-semibold text-[11px] mb-0.5">Cryptographic Node ID</label>
                <div className="p-2 rounded bg-slate-50 border border-slate-200 font-mono text-slate-700 text-[10px] flex items-center justify-between">
                  <span className="truncate max-w-[210px]">{nodeStatus?.peer_id || 'Generating...'}</span>
                  <button onClick={copyMyId} className="text-blue-700 hover:text-blue-900 ml-2" title="Copy">
                    {copiedId ? <Check className="w-3.5 h-3.5 text-emerald-600" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>

              <div className="pt-2 border-t border-slate-100 flex items-center justify-between">
                <span className="font-semibold text-slate-700">Audio Distress Beep</span>
                <button
                  type="button"
                  onClick={() => setSoundEnabled(!soundEnabled)}
                  className={`px-2.5 py-1 rounded text-xs font-bold uppercase border ${
                    soundEnabled
                      ? 'bg-emerald-50 border-emerald-300 text-emerald-800'
                      : 'bg-slate-100 border-slate-300 text-slate-500'
                  }`}
                >
                  {soundEnabled ? 'Enabled' : 'Muted'}
                </button>
              </div>
            </div>

            <div className="pt-2 border-t border-slate-200 flex justify-end">
              <button
                onClick={() => setSettingsOpen(false)}
                className="px-4 py-1.5 rounded bg-[#0b2545] text-white text-xs font-semibold"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
