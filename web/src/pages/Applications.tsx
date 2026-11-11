import React from 'react';
import { Application } from '../types';
import { RiskBadge } from '../components/RiskBadge';
import { FileDown, FileText, CheckCircle2, AlertTriangle, XCircle, Clock } from 'lucide-react';

interface Props {
  apps: Application[];
  onRefresh: () => void;
}

export const Applications: React.FC<Props> = ({ apps, onRefresh }) => {
  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 bg-slate-900/90 border border-slate-800 p-6 rounded-xl">
        <div>
          <h2 className="text-xl font-bold text-white">Discovered Applications & Shadow AI Map</h2>
          <p className="text-xs text-slate-400 mt-1">
            Internal services and client apps routing generative AI traffic through the governance gateway
          </p>
        </div>
      </div>

      {/* Grid of Applications */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {apps.length === 0 ? (
          <div className="col-span-full bg-slate-900/60 border border-slate-800 rounded-xl p-12 text-center text-slate-500">
            No applications discovered yet. Send requests to the proxy to auto-register services.
          </div>
        ) : (
          apps.map((app) => {
            return (
              <div
                key={app.id}
                className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 shadow-sm flex flex-col justify-between hover:border-slate-700 transition"
              >
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <RiskBadge tier={app.highest_risk_tier} size="sm" />
                    {app.compliance_status === 'COMPLIANT' && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-medium text-emerald-400 bg-emerald-950/60 border border-emerald-500/30 px-2 py-0.5 rounded-full">
                        <CheckCircle2 className="w-3 h-3" /> Compliant
                      </span>
                    )}
                    {app.compliance_status === 'NEEDS_REVIEW' && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-medium text-amber-400 bg-amber-950/60 border border-amber-500/30 px-2 py-0.5 rounded-full">
                        <AlertTriangle className="w-3 h-3" /> Needs Review
                      </span>
                    )}
                    {app.compliance_status === 'VIOLATION' && (
                      <span className="inline-flex items-center gap-1 text-[11px] font-medium text-red-400 bg-red-950/60 border border-red-500/30 px-2 py-0.5 rounded-full">
                        <XCircle className="w-3 h-3" /> Violation
                      </span>
                    )}
                  </div>

                  <h3 className="text-base font-bold text-white font-mono">{app.name}</h3>
                  <p className="text-xs text-slate-400 mt-1 font-mono">App ID: {app.id}</p>

                  <div className="grid grid-cols-2 gap-2 my-4 bg-slate-950 p-3 rounded-lg border border-slate-800/80 text-xs">
                    <div>
                      <span className="text-slate-500 block">Total Requests:</span>
                      <span className="font-mono font-bold text-slate-200">{app.total_requests.toLocaleString()}</span>
                    </div>
                    <div>
                      <span className="text-slate-500 block">PII Masked:</span>
                      <span className="font-mono font-bold text-emerald-400">{app.total_pii_masked.toLocaleString()}</span>
                    </div>
                  </div>

                  <div className="flex items-center gap-1.5 text-[11px] text-slate-400 mb-4">
                    <Clock className="w-3.5 h-3.5 text-slate-500" />
                    <span>Last active: {new Date(app.last_seen).toLocaleDateString()} {new Date(app.last_seen).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                  </div>
                </div>

                {/* Dossier Download Action Buttons */}
                <div className="pt-3 border-t border-slate-800 flex items-center gap-2">
                  <a
                    href={`/api/apps/${app.id}/dossier/pdf`}
                    download
                    className="flex-1 inline-flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition shadow-sm"
                  >
                    <FileDown className="w-3.5 h-3.5" />
                    PDF Dossier
                  </a>
                  <a
                    href={`/api/apps/${app.id}/dossier/md`}
                    download
                    className="inline-flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium border border-slate-700 transition"
                  >
                    <FileText className="w-3.5 h-3.5" />
                    Markdown
                  </a>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
