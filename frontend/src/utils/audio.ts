// Offline Web Audio API Engine for Emergency Sirens and Dispatch Alerts
// 100% offline, zero external audio asset dependencies.

let sharedAudioCtx: AudioContext | null = null;
let isUnlocked = false;

export function isAudioUnlocked(): boolean {
  return isUnlocked;
}

/**
 * Returns a singleton AudioContext instance, creating it if needed.
 */
function getAudioContext(): AudioContext | null {
  try {
    if (!sharedAudioCtx) {
      const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
      if (!AudioCtx) {
        console.warn('[Audio] Web Audio API is not supported in this browser.');
        return null;
      }
      sharedAudioCtx = new AudioCtx();
    }

    if (sharedAudioCtx.state === 'suspended') {
      sharedAudioCtx.resume().catch((err) => {
        console.warn('[Audio] Context resume deferred until user interaction:', err);
      });
    }

    return sharedAudioCtx;
  } catch (err) {
    console.error('[Audio] Error initializing AudioContext:', err);
    return null;
  }
}

/**
 * Auto-unlocks the AudioContext upon first user gesture on the page.
 * Required by modern browser autoplay policies.
 */
export function initAudioUnlocker(): () => void {
  if (typeof window === 'undefined') return () => {};

  const unlock = () => {
    const ctx = getAudioContext();
    if (ctx && ctx.state === 'suspended') {
      ctx.resume().then(() => {
        isUnlocked = true;
        console.log('[Audio] AudioContext unlocked by user gesture.');
      }).catch(() => {});
    } else if (ctx && ctx.state === 'running') {
      isUnlocked = true;
    }
  };

  const events = ['click', 'touchstart', 'keydown'];
  events.forEach((evt) => window.addEventListener(evt, unlock, { once: true, passive: true }));

  return () => {
    events.forEach((evt) => window.removeEventListener(evt, unlock));
  };
}

/**
 * Plays an emergency siren tone based on severity.
 * @param severity 'critical' | 'warning' | 'info' | 'test'
 */
export async function playAlertSound(severity: 'critical' | 'warning' | 'info' | 'test' = 'critical'): Promise<boolean> {
  try {
    const ctx = getAudioContext();
    if (!ctx) return false;

    if (ctx.state === 'suspended') {
      await ctx.resume();
    }

    const now = ctx.currentTime;

    if (severity === 'critical') {
      // High-priority alternating dual-tone emergency siren (880Hz <-> 660Hz warble)
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = 'sawtooth';

      // Alternating warble tone
      osc.frequency.setValueAtTime(880, now);
      osc.frequency.setValueAtTime(660, now + 0.15);
      osc.frequency.setValueAtTime(880, now + 0.30);
      osc.frequency.setValueAtTime(660, now + 0.45);
      osc.frequency.setValueAtTime(880, now + 0.60);

      // Smooth envelope to prevent audio clicking
      gain.gain.setValueAtTime(0.01, now);
      gain.gain.linearRampToValueAtTime(0.28, now + 0.05);
      gain.gain.setValueAtTime(0.28, now + 0.65);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.85);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now);
      osc.stop(now + 0.85);
      return true;
    }

    if (severity === 'warning') {
      // Two-step caution beep (620Hz -> 780Hz)
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = 'triangle';
      osc.frequency.setValueAtTime(620, now);
      osc.frequency.setValueAtTime(780, now + 0.18);

      gain.gain.setValueAtTime(0.01, now);
      gain.gain.linearRampToValueAtTime(0.25, now + 0.04);
      gain.gain.setValueAtTime(0.25, now + 0.35);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.50);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now);
      osc.stop(now + 0.50);
      return true;
    }

    // Info / Advisory / Test chime
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = 'sine';
    osc.frequency.setValueAtTime(587.33, now); // D5
    osc.frequency.setValueAtTime(880.00, now + 0.15); // A5

    gain.gain.setValueAtTime(0.01, now);
    gain.gain.linearRampToValueAtTime(0.22, now + 0.03);
    gain.gain.exponentialRampToValueAtTime(0.001, now + 0.45);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start(now);
    osc.stop(now + 0.45);
    return true;
  } catch (err) {
    console.warn('[Audio] Failed to playback alert sound:', err);
    return false;
  }
}
