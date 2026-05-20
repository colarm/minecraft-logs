'use client';

import { usePolling } from '@/lib/hooks';
import { api } from '@/lib/api';
import dynamic from 'next/dynamic';

const ReactECharts = dynamic(() => import('echarts-for-react'), { ssr: false });

const SERVER_ID = 'survival-01';

export default function DashboardPage() {
  const { data: stats } = usePolling(() => api.getStats(SERVER_ID), 5000);
  const { data: logs } = usePolling(() => api.getLogs(SERVER_ID, 50), 3000);

  const tpsOption = {
    tooltip: { formatter: '{b}: {c} TPS' },
    xAxis: { type: 'category', data: stats?.tps_history?.map((t: any) => new Date(t.bucket).toLocaleTimeString()) || [] },
    yAxis: { type: 'value', min: 0, max: 20 },
    series: [{
      type: 'line',
      data: stats?.tps_history?.map((t: any) => t.avg_tps) || [],
      smooth: true,
      lineStyle: { color: '#4fc3f7' },
      areaStyle: { color: 'rgba(79,195,247,0.2)' },
    }],
    grid: { top: 10, bottom: 30, left: 40, right: 20 },
  };

  return (
    <div>
      <h1 style={{ margin: '0 0 20px' }}>Server Dashboard</h1>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', marginBottom: '24px' }}>
        <StatCard label="Current TPS" value={stats?.current_tps?.toFixed(2) || '--'} color={tpsColor(stats?.current_tps)} />
        <StatCard label="Online Players" value={stats?.online_count ?? '--'} />
        <StatCard label="Peak Online" value={stats?.peak_online ?? '--'} />
        <StatCard label="Last Updated" value={stats?.last_updated ? new Date(stats.last_updated).toLocaleTimeString() : '--'} />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
        <div style={{ background: '#1a2332', borderRadius: '8px', padding: '16px' }}>
          <h3 style={{ margin: '0 0 12px' }}>TPS Trend (24h)</h3>
          <ReactECharts option={tpsOption} style={{ height: '300px' }} />
        </div>

        <div style={{ background: '#1a2332', borderRadius: '8px', padding: '16px' }}>
          <h3 style={{ margin: '0 0 12px' }}>Recent Events</h3>
          <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
            {logs?.events?.map((e: any) => (
              <div key={e.id} style={{ padding: '4px 0', borderBottom: '1px solid #2a3a4a', fontSize: '13px' }}>
                <span style={{ color: '#666' }}>{new Date(e.timestamp).toLocaleTimeString()}</span>{' '}
                <span style={{ color: eventTypeColor(e.event_type) }}>[{e.event_type}]</span>{' '}
                {e.player_name && <span style={{ color: '#aed581' }}>{e.player_name}</span>}
                {e.message && <span> {e.message}</span>}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: string | number; color?: string }) {
  return (
    <div style={{ background: '#1a2332', borderRadius: '8px', padding: '16px' }}>
      <div style={{ color: '#888', fontSize: '14px' }}>{label}</div>
      <div style={{ fontSize: '28px', fontWeight: 'bold', color: color || '#e0e0e0' }}>{value}</div>
    </div>
  );
}

function tpsColor(tps: number | undefined): string {
  if (tps == null) return '#e0e0e0';
  if (tps >= 18) return '#66bb6a';
  if (tps >= 15) return '#ffa726';
  return '#ef5350';
}

function eventTypeColor(type: string): string {
  switch (type) {
    case 'player_join': return '#66bb6a';
    case 'player_leave': return '#ef5350';
    case 'chat': return '#4fc3f7';
    case 'death': return '#ffa726';
    case 'error': return '#ef5350';
    default: return '#888';
  }
}
