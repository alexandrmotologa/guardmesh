import React from 'react';
import { Shield, Activity, Radio, FileText, Layers, Search, Cpu } from 'lucide-react';

interface Props {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  avgLatency: number;
}

export const Header: React.FC<Props> = ({ activeTab, setActiveTab, avgLatency }) => {
  const tabs = [
    { id: 'overview', label: 'Overview', icon: Activity },
    { id: 'events', label: 'Live Audit Log', icon: Radio },
    { id: 'apps', label: 'Applications & Dossiers', icon: Layers },
    { id: 'auditor', label: 'EU AI Act Auditor', icon: FileText },
    { id: 'scanner', label: 'PII De-Identifier', icon: Search },
  ];

  return (
    <header className="border-b border-slate-800 bg-slate-900/80 backdrop-blur-md sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo & Brand */}
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-emerald-600 to-teal-400 p-0.5 shadow-lg shadow-emerald-500/20">
              <div className="w-full h-full bg-slate-950 rounded-[10px] flex items-center justify-center">
                <Shield className="w-5 h-5 text-emerald-400" />
              </div>
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-bold text-lg text-white tracking-tight">guardmesh</span>
                <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  EU AI Act Gateway
                </span>
              </div>
              <p className="text-xs text-slate-400 hidden sm:block">Real-time PII Anonymizer & Governance Proxy</p>
            </div>
          </div>

          {/* Navigation Tabs */}
          <nav className="flex space-x-1 sm:space-x-2">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all ${
                    isActive
                      ? 'bg-slate-800 text-emerald-400 border border-slate-700 shadow-sm'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  <span className="hidden md:inline">{tab.label}</span>
                </button>
              );
            })}
          </nav>

          {/* Live Status Indicator */}
          <div className="flex items-center gap-4">
            <div className="hidden lg:flex items-center gap-2 px-2.5 py-1 rounded-md bg-slate-800/80 border border-slate-700/60 text-xs">
              <Cpu className="w-3.5 h-3.5 text-slate-400" />
              <span className="text-slate-400">Proxy Overhead:</span>
              <span className="font-mono text-emerald-400 font-medium">
                {avgLatency > 0 ? `${avgLatency.toFixed(1)}ms` : '3.2ms'}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="relative flex h-2.5 w-2.5">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
              </span>
              <span className="text-xs font-mono text-slate-300 hidden sm:inline">Active</span>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};
