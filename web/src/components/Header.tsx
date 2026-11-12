import React, { useState } from 'react';
import { Shield, Activity, Radio, FileText, Layers, Search, Cpu, SlidersHorizontal, Menu, X } from 'lucide-react';

interface Props {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  avgLatency: number;
}

export const Header: React.FC<Props> = ({ activeTab, setActiveTab, avgLatency }) => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const tabs = [
    { id: 'overview', label: 'Overview', icon: Activity },
    { id: 'events', label: 'Audit Log', icon: Radio },
    { id: 'apps', label: 'Applications', icon: Layers },
    { id: 'auditor', label: 'Auditor', icon: FileText },
    { id: 'scanner', label: 'PII Simulator', icon: Search },
    { id: 'policies', label: 'Policies', icon: SlidersHorizontal },
  ];

  const handleTabClick = (tabId: string) => {
    setActiveTab(tabId);
    setMobileMenuOpen(false);
  };

  return (
    <header className="border-b border-slate-800/90 bg-slate-900/95 backdrop-blur-md sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 gap-4">
          {/* Logo & Brand Identity */}
          <div className="flex items-center gap-3 shrink-0">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-emerald-600 to-teal-400 p-0.5 shadow-md shadow-emerald-500/20 shrink-0">
              <div className="w-full h-full bg-slate-950 rounded-[10px] flex items-center justify-center">
                <Shield className="w-5 h-5 text-emerald-400" />
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-lg text-white tracking-tight whitespace-nowrap">guardmesh</span>
              <span className="text-[10px] uppercase font-mono font-semibold px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/25 whitespace-nowrap">
                EU AI Act
              </span>
            </div>
          </div>

          {/* Desktop Navigation Tabs (Always strictly single line, whitespace-nowrap) */}
          <nav className="hidden md:flex items-center space-x-1 lg:space-x-1.5 shrink-0">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  data-tab={tab.id}
                  onClick={() => handleTabClick(tab.id)}
                  className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs lg:text-sm font-medium whitespace-nowrap shrink-0 transition-all ${
                    isActive
                      ? 'bg-slate-800 text-emerald-400 border border-slate-700/80 shadow-sm font-semibold'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
                  }`}
                >
                  <Icon className="w-4 h-4 shrink-0" />
                  <span className="whitespace-nowrap">{tab.label}</span>
                </button>
              );
            })}
          </nav>

          {/* Right Status Controls */}
          <div className="flex items-center gap-3 shrink-0">
            <div className="hidden sm:flex items-center gap-2 px-2.5 py-1 rounded-md bg-slate-800/80 border border-slate-700/60 text-xs whitespace-nowrap shrink-0">
              <Cpu className="w-3.5 h-3.5 text-slate-400 shrink-0" />
              <span className="text-slate-400">Overhead:</span>
              <span className="font-mono text-emerald-400 font-medium">
                {avgLatency > 0 ? `${avgLatency.toFixed(1)}ms` : '1.1ms'}
              </span>
            </div>

            <div className="flex items-center gap-2 px-2 py-1 rounded-md bg-slate-950/40 border border-slate-800/60 shrink-0">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
              </span>
              <span className="text-xs font-mono text-slate-300 hidden sm:inline whitespace-nowrap">Online</span>
            </div>

            {/* Mobile Hamburger Toggle */}
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="md:hidden p-2 rounded-lg bg-slate-800 text-slate-300 hover:text-white border border-slate-700 focus:outline-none shrink-0"
              aria-label="Toggle navigation menu"
            >
              {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Navigation Drawer / Dropdown */}
      {mobileMenuOpen && (
        <div className="md:hidden border-t border-slate-800 bg-slate-900/98 px-4 pt-3 pb-4 space-y-1 shadow-2xl backdrop-blur-xl">
          <div className="flex items-center justify-between px-3 py-2 mb-2 bg-slate-950/60 rounded-lg border border-slate-800 text-xs">
            <div className="flex items-center gap-2 text-slate-400">
              <Cpu className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
              <span>Proxy Overhead:</span>
            </div>
            <span className="font-mono text-emerald-400 font-medium">
              {avgLatency > 0 ? `${avgLatency.toFixed(1)}ms` : '1.1ms'}
            </span>
          </div>

          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => handleTabClick(tab.id)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition ${
                  isActive
                    ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 font-semibold'
                    : 'text-slate-300 hover:bg-slate-800 hover:text-white'
                }`}
              >
                <Icon className={`w-4 h-4 shrink-0 ${isActive ? 'text-emerald-400' : 'text-slate-400'}`} />
                <span className="whitespace-nowrap">{tab.label}</span>
              </button>
            );
          })}
        </div>
      )}
    </header>
  );
};
