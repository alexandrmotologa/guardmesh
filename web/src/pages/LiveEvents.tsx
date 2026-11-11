import React, { useState } from 'react';
import { AuditEvent, RiskTier } from '../types';
import { RiskBadge } from '../components/RiskBadge';
import { Search, Filter, Ban, CheckCircle, ExternalLink, X, RefreshCw, Download, FileSpreadsheet } from 'lucide-react';

interface Props {
  events: AuditEvent[];
  onRefresh: () => void;
}

export const LiveEvents: React.FC<Props> = ({ events, onRefresh }) => {
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null);
  const [search, setSearch] = useState('');
  const [selectedTier, setSelectedTier] = useState<string>('ALL');

  React.useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setSelectedEvent(null);
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const filtered = events.filter((e) => {
    const matchesSearch =
      e.app_id.toLowerCase().includes(search.toLowerCase()) ||
      e.model.toLowerCase().includes(search.toLowerCase()) ||
      e.sample_snippet.toLowerCase().includes(search.toLowerCase());
    const matchesTier = selectedTier === 'ALL' || e.risk_tier === selectedTier;
    return matchesSearch && matchesTier;
  });

  const getExportUrl = (format: 'csv' | 'jsonl') => {
    let url = `/api/events/export?format=${format}`;
    if (selectedTier !== 'ALL') {
      url += `&risk_tier=${selectedTier}`;
    }
    return url;
  };

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Controls Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-4 bg-slate-900/90 border border-slate-800 p-4 rounded-xl">
        <div className="relative w-full sm:w-80">
          <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search by app, model, or snippet..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 focus:outline-none focus:border-emerald-500 transition"
          />
        </div>

        <div className="flex items-center gap-3 w-full sm:w-auto">
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-slate-400" />
            <select
              value={selectedTier}
              onChange={(e) => setSelectedTier(e.target.value)}
              className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-slate-200 focus:outline-none focus:border-emerald-500"
            >
              <option value="ALL">All Risk Tiers</option>
              <option value="PROHIBITED">Prohibited (Art. 5)</option>
              <option value="HIGH_RISK">High-Risk (Annex III)</option>
              <option value="LIMITED_RISK">Limited Risk (Art. 50)</option>
              <option value="MINIMAL_RISK">Minimal Risk</option>
            </select>
          </div>

          <div className="flex items-center gap-2">
            <a
              href={getExportUrl('csv')}
              download="guardmesh_audit.csv"
              className="flex items-center gap-1.5 px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium border border-slate-700 transition"
              title="Export filtered events as CSV"
            >
              <FileSpreadsheet className="w-3.5 h-3.5 text-emerald-400" />
              <span>CSV</span>
            </a>
            <a
              href={getExportUrl('jsonl')}
              download="guardmesh_audit.jsonl"
              className="flex items-center gap-1.5 px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium border border-slate-700 transition"
              title="Export filtered events as JSONL"
            >
              <Download className="w-3.5 h-3.5 text-blue-400" />
              <span>JSONL</span>
            </a>
            <button
              onClick={onRefresh}
              className="flex items-center gap-2 px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium border border-slate-700 transition"
            >
              <RefreshCw className="w-3.5 h-3.5" />
              <span>Refresh</span>
            </button>
          </div>
        </div>
      </div>

      {/* Events Table */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-xl overflow-hidden shadow-lg">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-950/80 text-xs font-semibold text-slate-400 uppercase tracking-wider border-b border-slate-800">
              <tr>
                <th className="px-4 py-3">Timestamp</th>
                <th className="px-4 py-3">Application</th>
                <th className="px-4 py-3">Model</th>
                <th className="px-4 py-3">Risk Tier</th>
                <th className="px-4 py-3">PII Masked</th>
                <th className="px-4 py-3">Latency</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3 text-right">Details</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 font-mono text-xs">
              {filtered.length === 0 ? (
                <tr>
                  <td colSpan={8} className="px-6 py-12 text-center text-slate-500 font-sans">
                    No proxy audit events found matching filters.
                  </td>
                </tr>
              ) : (
                filtered.map((e) => (
                  <tr
                    key={e.id}
                    onClick={() => setSelectedEvent(e)}
                    className="hover:bg-slate-800/40 cursor-pointer transition"
                  >
                    <td className="px-4 py-3 text-slate-400 whitespace-nowrap">
                      {new Date(e.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                    </td>
                    <td className="px-4 py-3 font-medium text-slate-200 font-sans">{e.app_id}</td>
                    <td className="px-4 py-3 text-slate-300">{e.model}</td>
                    <td className="px-4 py-3">
                      <RiskBadge tier={e.risk_tier} size="sm" />
                    </td>
                    <td className="px-4 py-3">
                      {e.pii_detected_count > 0 ? (
                        <span className="text-emerald-400 font-bold">
                          {e.pii_detected_count} ({e.pii_categories.join(', ')})
                        </span>
                      ) : (
                        <span className="text-slate-500">0</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-slate-300">{e.latency_ms.toFixed(1)} ms</td>
                    <td className="px-4 py-3 font-sans">
                      {e.is_blocked ? (
                        <span className="inline-flex items-center gap-1 text-red-400 font-medium">
                          <Ban className="w-3.5 h-3.5" /> Blocked
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1 text-emerald-400 font-medium">
                          <CheckCircle className="w-3.5 h-3.5" /> Passed
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <ExternalLink className="w-4 h-4 text-slate-500 hover:text-slate-300 inline" />
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Modal Inspector */}
      {selectedEvent && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-2xl w-full p-6 shadow-2xl relative animate-scaleIn max-h-[90vh] overflow-y-auto">
            <button
              onClick={() => setSelectedEvent(null)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white p-1 rounded-lg bg-slate-800 hover:bg-slate-700 transition"
            >
              <X className="w-5 h-5" />
            </button>

            <div className="flex items-center gap-3 mb-4">
              <RiskBadge tier={selectedEvent.risk_tier} size="lg" />
              <div className="text-xs font-mono text-slate-400">ID: {selectedEvent.id}</div>
            </div>

            <h3 className="text-lg font-bold text-white mb-1">
              Audit Event Telemetry & Legal Assessment
            </h3>
            <p className="text-xs text-slate-400 mb-6 font-mono">
              Fingerprint (SHA-256): {selectedEvent.prompt_sha256}
            </p>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6 bg-slate-950 p-4 rounded-xl border border-slate-800 text-xs">
              <div>
                <span className="text-slate-500 block">Application:</span>
                <span className="font-semibold text-slate-200">{selectedEvent.app_id}</span>
              </div>
              <div>
                <span className="text-slate-500 block">Model:</span>
                <span className="font-semibold text-slate-200">{selectedEvent.model}</span>
              </div>
              <div>
                <span className="text-slate-500 block">Latency Overhead:</span>
                <span className="font-semibold text-emerald-400 font-mono">{selectedEvent.latency_ms.toFixed(1)} ms</span>
              </div>
              <div>
                <span className="text-slate-500 block">PII Masked:</span>
                <span className="font-semibold text-slate-200 font-mono">{selectedEvent.pii_detected_count} entities</span>
              </div>
            </div>

            {selectedEvent.is_blocked && (
              <div className="mb-6 p-4 rounded-xl bg-red-950/60 border border-red-500/50 text-red-300 text-xs">
                <div className="font-bold flex items-center gap-2 mb-1">
                  <Ban className="w-4 h-4 text-red-400" />
                  Request Hard-Blocked (HTTP 403 Forbidden)
                </div>
                <div>{selectedEvent.blocked_reason}</div>
              </div>
            )}

            <div className="mb-6">
              <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                EU AI Act Classification Reasons
              </h4>
              <ul className="space-y-1.5 text-xs text-slate-300">
                {selectedEvent.risk_reasons.map((r, i) => (
                  <li key={i} className="flex items-start gap-2 bg-slate-950/60 p-2 rounded-lg border border-slate-800">
                    <span className="text-emerald-400 font-bold">•</span>
                    <span>{r}</span>
                  </li>
                ))}
              </ul>
            </div>

            <div>
              <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                Sanitized Request Sample
              </h4>
              <div className="bg-slate-950 p-3 rounded-lg border border-slate-800 text-xs font-mono text-slate-300 overflow-x-auto">
                {selectedEvent.sample_snippet}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
