const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001';

async function fetchApi<T>(path: string, params?: Record<string, string>): Promise<T> {
  const url = new URL(`/api${path}`, API_URL);
  if (params) {
    Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));
  }
  const res = await fetch(url.toString());
  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json();
}

export function usePolling<T>(
  fetchFn: () => Promise<T>,
  intervalMs: number,
): { data: T | null; loading: boolean } {
  // This is a simple description - actual hook implementation is in the React component files
  // Using a custom hook pattern defined in the component files
  return { data: null, loading: false };
}

export const api = {
  getStats: (serverId: string) =>
    fetchApi('/stats', { server_id: serverId }),

  getPlayers: (serverId: string, limit = 50, offset = 0, search?: string) =>
    fetchApi('/players', {
      server_id: serverId,
      limit: String(limit),
      offset: String(offset),
      ...(search ? { search } : {}),
    }),

  getLogs: (serverId: string, limit = 100, offset = 0, eventType?: string, playerName?: string) =>
    fetchApi('/logs', {
      server_id: serverId,
      limit: String(limit),
      offset: String(offset),
      ...(eventType ? { event_type: eventType } : {}),
      ...(playerName ? { player_name: playerName } : {}),
    }),

  getHistory: (serverId: string, start: string, end: string, interval = '1h') =>
    fetchApi('/history', {
      server_id: serverId,
      start,
      end,
      interval,
    }),
};
