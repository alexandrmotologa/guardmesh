import React, { useState, useEffect } from 'react';
import { Header } from './components/Header';
import { Overview } from './pages/Overview';
import { LiveEvents } from './pages/LiveEvents';
import { Applications } from './pages/Applications';
import { Auditor } from './pages/Auditor';
import { PIIScanner } from './pages/PIIScanner';
import { GlobalStats, AuditEvent, Application } from './types';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState('overview');
  const [stats, setStats] = useState<GlobalStats | null>(null);
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [apps, setApps] = useState<Application[]>([]);

  const fetchData = async () => {
    try {
      const [statsRes, eventsRes, appsRes] = await Promise.all([
        fetch('/api/stats').then((r) => (r.ok ? r.json() : null)),
        fetch('/api/events?limit=50').then((r) => (r.ok ? r.json() : [])),
        fetch('/api/apps').then((r) => (r.ok ? r.json() : [])),
      ]);
      if (statsRes) setStats(statsRes);
      if (eventsRes) setEvents(eventsRes);
      if (appsRes) setApps(appsRes);
    } catch (err) {
      console.error('Failed to fetch telemetry data:', err);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col font-sans">
      <Header
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        avgLatency={stats?.avg_latency_ms || 3.2}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {activeTab === 'overview' && (
          <Overview stats={stats} onNavigate={(tab) => setActiveTab(tab)} />
        )}
        {activeTab === 'events' && (
          <LiveEvents events={events} onRefresh={fetchData} />
        )}
        {activeTab === 'apps' && (
          <Applications apps={apps} onRefresh={fetchData} />
        )}
        {activeTab === 'auditor' && <Auditor />}
        {activeTab === 'scanner' && <PIIScanner />}
      </main>

      <footer className="border-t border-slate-900 bg-slate-950/80 py-4 text-center text-xs text-slate-500 font-mono">
        guardmesh v1.0.0 &bull; Regulation (EU) 2024/1689 &bull; NIST AI RMF 1.0 &bull; Sub-5ms Proxy Daemon
      </footer>
    </div>
  );
};

export default App;
