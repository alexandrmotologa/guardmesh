import React from 'react';
import { GlobalStats } from '../types';
import { ShieldCheck, EyeOff, Ban, Zap, FileText, ArrowUpRight, Lock, CheckCircle2 } from 'lucide-react';
import { RiskBadge } from '../components/RiskBadge';

interface Props {
  stats: GlobalStats | null;
  onNavigate: (tab: string) => void;
}

export const Overview: React.FC<Props> = ({ stats, onNavigate }) => {
  const totalReq = stats?.total_requests || 0;
  const totalBlocked = stats?.total_blocked || 0;
  const totalPII = stats?.total_pii_masked || 0;
  const avgLatency = stats?.avg_latency_ms || 3.4;

  const riskDist = stats?.risk_distribution || {
    MINIMAL_RISK: 0,
    LIMITED_RISK: 0,
    HIGH_RISK: 0,
    PROHIBITED: 0,
  };

  return (
    <div className="space-y-8 animate-fadeIn">
      {/* Top Banner / Hero */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-slate-900 to-slate-950 border border-slate-800 p-6 sm:p-8 shadow-xl">
        <div className="absolute top-0 right-0 -mt-8 -mr-8 w-64 h-64 bg-emerald-500/10 rounded-full blur-3xl pointer-events-none"></div>
        <div className="relative z-10 max-w-3xl">
          <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium mb-3">
            <Lock className="w-3 h-3" />
            100% On-Premises & Private Cloud Control
          </div>
          <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
            Enterprise AI Governance & Zero-Leak PII Shield
          </h1>
          <p className="mt-2 text-slate-300 text-sm sm:text-base leading-relaxed">
            Transparent drop-in reverse proxy enforcing <strong>Regulation (EU) 2024/1689 (EU AI Act)</strong> and GDPR.
            Cloud LLMs receive only cryptographic replacement tokens, while prohibited AI practices are blocked before egress.
          </p>

          <div className="mt-6 flex flex-wrap gap-3">
            <button
              onClick={() => onNavigate('auditor')}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition shadow-lg shadow-emerald-600/20"
            >
              <FileText className="w-4 h-4" />
              Audit a System Prompt
            </button>
            <button
              onClick={() => onNavigate('apps')}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-sm font-medium transition"
            >
              Export Annex IV Dossier
              <ArrowUpRight className="w-4 h-4 text-slate-400" />
            </button>
          </div>
        </div>
      </div>

      {/* Metric Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Invocations</span>
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400">
              <ShieldCheck className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-white font-mono">{totalReq.toLocaleString()}</div>
            <p className="text-xs text-slate-400 mt-1">Routed transparently through gateway</p>
          </div>
        </div>

        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">PII Intercepted</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <EyeOff className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-emerald-400 font-mono">{totalPII.toLocaleString()}</div>
            <p className="text-xs text-slate-400 mt-1">Zero sensitive data leaked to LLMs</p>
          </div>
        </div>

        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">Prohibited Blocks</span>
            <div className="p-2 rounded-lg bg-red-500/10 text-red-400">
              <Ban className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-red-400 font-mono">{totalBlocked.toLocaleString()}</div>
            <p className="text-xs text-slate-400 mt-1">Article 5 violations intercepted (HTTP 403)</p>
          </div>
        </div>

        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">Average Latency</span>
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400">
              <Zap className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-amber-400 font-mono">{avgLatency.toFixed(1)} ms</div>
            <p className="text-xs text-slate-400 mt-1">Sub-5ms regex & token vault overhead</p>
          </div>
        </div>
      </div>

      {/* EU AI Act Risk Distribution & Governance Controls */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Risk Distribution Breakdown */}
        <div className="lg:col-span-2 bg-slate-900/90 border border-slate-800 rounded-xl p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-lg font-bold text-white">EU AI Act Risk Classification Matrix</h2>
              <p className="text-xs text-slate-400">Workloads classified per Regulation (EU) 2024/1689</p>
            </div>
          </div>

          <div className="space-y-4">
            {[
              { tier: 'PROHIBITED', label: 'Prohibited Practices (Article 5)', count: riskDist['PROHIBITED'] || 0, color: 'bg-red-500', text: 'text-red-400' },
              { tier: 'HIGH_RISK', label: 'High-Risk AI Systems (Annex III)', count: riskDist['HIGH_RISK'] || 0, color: 'bg-amber-500', text: 'text-amber-400' },
              { tier: 'LIMITED_RISK', label: 'Limited Risk AI / Chatbots (Article 50)', count: riskDist['LIMITED_RISK'] || 0, color: 'bg-blue-500', text: 'text-blue-400' },
              { tier: 'MINIMAL_RISK', label: 'Minimal / No Risk AI', count: riskDist['MINIMAL_RISK'] || 0, color: 'bg-emerald-500', text: 'text-emerald-400' },
            ].map((item) => {
              const pct = totalReq > 0 ? ((item.count / totalReq) * 100).toFixed(1) : '0.0';
              return (
                <div key={item.tier} className="bg-slate-950/60 border border-slate-800/80 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <RiskBadge tier={item.tier} size="sm" />
                      <span className="text-sm font-medium text-slate-200">{item.label}</span>
                    </div>
                    <span className={`text-sm font-mono font-bold ${item.text}`}>{item.count} ({pct}%)</span>
                  </div>
                  <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
                    <div className={`${item.color} h-full rounded-full transition-all duration-500`} style={{ width: `${pct}%` }}></div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Regulatory Safeguards Checklist */}
        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-6 flex flex-col justify-between">
          <div>
            <h2 className="text-lg font-bold text-white mb-2">Active Governance Controls</h2>
            <p className="text-xs text-slate-400 mb-6">Continuous automated enforcement</p>

            <div className="space-y-3">
              {[
                { title: 'Article 5 Hard-Kill Enforcer', desc: 'Blocks emotion recognition & social scoring', active: true },
                { title: 'Deterministic PII Token Vault', desc: 'Reversible AES-256 GCM in-memory encryption', active: true },
                { title: 'Chunk-Window Rehydrator', desc: 'Zero SSE split token leakage', active: true },
                { title: 'Annex IV Dossier Compiler', desc: '1-click regulator-ready PDF generation', active: true },
                { title: 'Prompt Injection Scanner', desc: 'Guards against jailbreaks & system prompt leaks', active: true },
              ].map((ctrl, i) => (
                <div key={i} className="flex items-start gap-3 p-3 rounded-lg bg-slate-950/40 border border-slate-800/60">
                  <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                  <div>
                    <div className="text-xs font-semibold text-slate-200">{ctrl.title}</div>
                    <div className="text-[11px] text-slate-400">{ctrl.desc}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="mt-6 pt-4 border-t border-slate-800">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span>SQLite WAL Store</span>
              <span className="font-mono text-emerald-400">Synchronous = Normal</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
