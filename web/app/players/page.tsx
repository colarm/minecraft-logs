'use client';

import { useState } from 'react';
import { usePolling } from '@/lib/hooks';
import { api } from '@/lib/api';

const SERVER_ID = 'survival-01';

export default function PlayersPage() {
  const [search, setSearch] = useState('');
  const { data, loading } = usePolling(
    () => api.getPlayers(SERVER_ID, 100, 0, search || undefined),
    5000,
  );

  return (
    <div>
      <h1 style={{ margin: '0 0 20px' }}>Players</h1>
      <input
        type="text"
        placeholder="Search player..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        style={{
          padding: '8px 12px', background: '#1a2332', border: '1px solid #2a3a4a',
          borderRadius: '4px', color: '#e0e0e0', marginBottom: '16px', width: '250px',
        }}
      />
      {loading ? <p>Loading...</p> : (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid #2a3a4a', textAlign: 'left' }}>
              <th style={{ padding: '8px' }}>Name</th>
              <th style={{ padding: '8px' }}>Status</th>
              <th style={{ padding: '8px' }}>Join Count</th>
              <th style={{ padding: '8px' }}>Play Time</th>
              <th style={{ padding: '8px' }}>Last Seen</th>
            </tr>
          </thead>
          <tbody>
            {data?.players?.map((p: any) => (
              <tr key={p.id} style={{ borderBottom: '1px solid #1a2332' }}>
                <td style={{ padding: '8px' }}>{p.name}</td>
                <td style={{ padding: '8px' }}>
                  <span style={{
                    padding: '2px 8px', borderRadius: '4px', fontSize: '12px',
                    background: p.is_online ? '#1b5e20' : '#37474f',
                    color: p.is_online ? '#a5d6a7' : '#90a4ae',
                  }}>
                    {p.is_online ? 'Online' : 'Offline'}
                  </span>
                </td>
                <td style={{ padding: '8px' }}>{p.join_count}</td>
                <td style={{ padding: '8px' }}>{formatPlayTime(p.total_play_time)}</td>
                <td style={{ padding: '8px' }}>{new Date(p.last_seen).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

function formatPlayTime(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return `${h}h ${m}m`;
}
