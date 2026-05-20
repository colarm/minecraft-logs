'use client';

import { useState } from 'react';
import { usePolling } from '@/lib/hooks';
import { api } from '@/lib/api';
import dynamic from 'next/dynamic';

const ReactECharts = dynamic(() => import('echarts-for-react'), { ssr: false });

const SERVER_ID = 'survival-01';

export default function HistoryPage() {
  const [start, setStart] = useState(() => {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    return d.toISOString();
  });
  const [end, setEnd] = useState(() => new Date().toISOString());
  const [interval, setInterval] = useState('1h');

  const { data, loading } = usePolling(
    () => api.getHistory(SERVER_ID, start, end, interval),
    30000,
  );

  const chartOption = {
    tooltip: { trigger: 'axis' as const },
    legend: { data: ['TPS', 'Joins', 'Leaves', 'Chat', 'Deaths'] },
    xAxis: {
      type: 'category',
      data: data?.buckets?.map((b: any) => new Date(b.bucket).toLocaleString()) || [],
    },
    yAxis: [
      { type: 'value', name: 'TPS', min: 0, max: 20 },
      { type: 'value', name: 'Count' },
    ],
    series: [
      { name: 'TPS', type: 'line', yAxisIndex: 0, data: data?.buckets?.map((b: any) => b.avg_tps) || [], smooth: true },
      { name: 'Joins', type: 'bar', yAxisIndex: 1, data: data?.buckets?.map((b: any) => b.joins) || [] },
      { name: 'Leaves', type: 'bar', yAxisIndex: 1, data: data?.buckets?.map((b: any) => b.leaves) || [] },
      { name: 'Chat', type: 'bar', yAxisIndex: 1, data: data?.buckets?.map((b: any) => b.chat_messages) || [] },
      { name: 'Deaths', type: 'bar', yAxisIndex: 1, data: data?.buckets?.map((b: any) => b.deaths) || [] },
    ],
    grid: { top: 40, bottom: 40, left: 50, right: 50 },
  };

  return (
    <div>
      <h1 style={{ margin: '0 0 20px' }}>History</h1>

      <div style={{ display: 'flex', gap: '12px', marginBottom: '20px', alignItems: 'center' }}>
        <label style={{ color: '#888' }}>
          Start:
          <input type="datetime-local" value={start.slice(0, 16)} onChange={(e) => setStart(new Date(e.target.value).toISOString())}
            style={{ marginLeft: '8px', padding: '4px 8px', background: '#1a2332', border: '1px solid #2a3a4a', borderRadius: '4px', color: '#e0e0e0' }} />
        </label>
        <label style={{ color: '#888' }}>
          End:
          <input type="datetime-local" value={end.slice(0, 16)} onChange={(e) => setEnd(new Date(e.target.value).toISOString())}
            style={{ marginLeft: '8px', padding: '4px 8px', background: '#1a2332', border: '1px solid #2a3a4a', borderRadius: '4px', color: '#e0e0e0' }} />
        </label>
        <label style={{ color: '#888' }}>
          Interval:
          <select value={interval} onChange={(e) => setInterval(e.target.value)}
            style={{ marginLeft: '8px', padding: '4px 8px', background: '#1a2332', border: '1px solid #2a3a4a', borderRadius: '4px', color: '#e0e0e0' }}>
            <option value="1m">1 minute</option>
            <option value="5m">5 minutes</option>
            <option value="15m">15 minutes</option>
            <option value="1h">1 hour</option>
            <option value="6h">6 hours</option>
            <option value="1d">1 day</option>
          </select>
        </label>
      </div>

      {loading ? <p>Loading...</p> : (
        <div style={{ background: '#1a2332', borderRadius: '8px', padding: '16px' }}>
          <ReactECharts option={chartOption} style={{ height: '400px' }} />
        </div>
      )}
    </div>
  );
}
