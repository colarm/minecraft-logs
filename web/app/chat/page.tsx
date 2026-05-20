'use client';

import { useRef, useEffect } from 'react';
import { usePolling } from '@/lib/hooks';
import { api } from '@/lib/api';

const SERVER_ID = 'survival-01';

export default function ChatPage() {
  const { data, loading } = usePolling(() => api.getLogs(SERVER_ID, 200, 0, 'chat'), 3000);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [data?.events?.length]);

  const events = data?.events ? [...data.events].reverse() : [];

  return (
    <div>
      <h1 style={{ margin: '0 0 20px' }}>Chat Log</h1>
      <div
        ref={containerRef}
        style={{
          background: '#1a2332', borderRadius: '8px', padding: '16px',
          maxHeight: '70vh', overflowY: 'auto', fontFamily: 'monospace',
        }}
      >
        {loading ? <p>Loading...</p> : events.map((e: any) => (
          <div key={e.id} style={{ padding: '2px 0', borderBottom: '1px solid #2a3a4a' }}>
            <span style={{ color: '#666' }}>{new Date(e.timestamp).toLocaleTimeString()}</span>{' '}
            <span style={{ color: '#aed581', fontWeight: 'bold' }}>&lt;{e.player_name}&gt;</span>{' '}
            <span style={{ color: '#e0e0e0' }}>{e.message}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
