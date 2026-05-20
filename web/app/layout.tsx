import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'MC Logs Dashboard',
  description: 'Minecraft server log analysis dashboard',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body style={{ margin: 0, fontFamily: 'system-ui, sans-serif', background: '#0f1923', color: '#e0e0e0' }}>
        <nav style={{
          display: 'flex', gap: '16px', padding: '12px 24px',
          background: '#1a2332', borderBottom: '1px solid #2a3a4a',
        }}>
          <a href="/" style={{ color: '#4fc3f7', textDecoration: 'none' }}>Dashboard</a>
          <a href="/players" style={{ color: '#4fc3f7', textDecoration: 'none' }}>Players</a>
          <a href="/chat" style={{ color: '#4fc3f7', textDecoration: 'none' }}>Chat</a>
          <a href="/history" style={{ color: '#4fc3f7', textDecoration: 'none' }}>History</a>
        </nav>
        <main style={{ padding: '24px' }}>
          {children}
        </main>
      </body>
    </html>
  );
}
