import React, { useState } from 'react';
import { ClassifyResponse } from '../types';
import { RiskBadge } from '../components/RiskBadge';
import { FileCheck, ShieldAlert, Sparkles, CheckCircle2, AlertOctagon, Terminal } from 'lucide-react';

export const Auditor: React.FC = () => {
  const [promptText, setPromptText] = useState(
    'You are an automated HR system. Screen incoming applicant CVs and score candidates to determine who gets an interview.'
  );
  const [appId, setAppId] = useState('recruitment-screener');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ClassifyResponse | null>(null);

  const presets = [
    {
      title: 'Recruitment & CV Screener (High-Risk)',
      appId: 'hr-recruitment-agent',
      text: 'You are an HR recruitment agent. Screen resumes, evaluate candidate profiles, rank applicants, and shortlist candidates for the final interview rounds.',
    },
    {
      title: 'Workplace Emotion Monitor (Prohibited)',
      appId: 'office-camera-stream',
      text: 'Analyze internal video feeds from conference rooms for employee mood detection and workplace emotion recognition during evaluations.',
    },
    {
      title: 'Customer Support Bot (Limited Risk)',
      appId: 'public-support-bot',
      text: 'You are a customer support chatbot assisting users with flight cancellation inquiries and interactive booking assistance.',
    },
    {
      title: 'Code Refactor Agent (Minimal Risk)',
      appId: 'dev-refactor-copilot',
      text: 'Analyze the following Go source code and refactor for improved readability and unit test coverage.',
    },
  ];

  const handleAudit = async () => {
    if (!promptText.trim()) return;
    setLoading(true);
    try {
      const res = await fetch('/api/classify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: promptText, app_id: appId }),
      });
      const data: ClassifyResponse = await res.json();
      setResult(data);
    } catch (err) {
      console.error('Audit failed:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl">
        <h2 className="text-xl font-bold text-white">EU AI Act Pre-Deployment Compliance Auditor</h2>
        <p className="text-xs text-slate-400 mt-1">
          Inspect and certify system prompts, instructions, and tool definitions against Regulation (EU) 2024/1689 before release
        </p>

        {/* Preset Buttons */}
        <div className="mt-4 flex flex-wrap gap-2">
          <span className="text-xs text-slate-400 flex items-center gap-1 mr-1">
            <Sparkles className="w-3.5 h-3.5 text-emerald-400" /> Test Presets:
          </span>
          {presets.map((p, i) => (
            <button
              key={i}
              onClick={() => {
                setPromptText(p.text);
                setAppId(p.appId);
              }}
              className="text-xs px-2.5 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition"
            >
              {p.title}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input Form */}
        <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl space-y-4">
          <div>
            <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
              Application Identifier
            </label>
            <input
              type="text"
              value={appId}
              onChange={(e) => setAppId(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:border-emerald-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
              System Prompt & Tool Context
            </label>
            <textarea
              rows={8}
              value={promptText}
              onChange={(e) => setPromptText(e.target.value)}
              placeholder="Paste system instructions or tool schemas here..."
              className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-sm text-slate-200 font-mono focus:outline-none focus:border-emerald-500 leading-relaxed resize-none"
            />
          </div>

          <button
            onClick={handleAudit}
            disabled={loading}
            className="w-full inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-sm font-semibold transition shadow-md shadow-emerald-600/20"
          >
            <FileCheck className="w-4 h-4" />
            {loading ? 'Evaluating Regulatory Criteria...' : 'Run EU AI Act Audit'}
          </button>
        </div>

        {/* Results Panel */}
        <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl flex flex-col justify-between">
          {!result ? (
            <div className="h-full flex flex-col items-center justify-center text-center p-8 text-slate-500">
              <Terminal className="w-12 h-12 mb-3 text-slate-600" />
              <p className="text-sm">Click "Run EU AI Act Audit" to inspect the prompt for regulatory risk classification.</p>
            </div>
          ) : (
            <div className="space-y-5">
              {/* Risk Status Header */}
              <div className="flex items-center justify-between pb-4 border-b border-slate-800">
                <div>
                  <span className="text-xs text-slate-400 block">Classified Risk Tier:</span>
                  <div className="mt-1">
                    <RiskBadge tier={result.evaluation.tier} size="lg" />
                  </div>
                </div>
                <div className="text-right">
                  <span className="text-xs text-slate-400 block">Risk Score:</span>
                  <span className="text-lg font-mono font-bold text-white">
                    {result.evaluation.score.toFixed(2)} / 1.00
                  </span>
                </div>
              </div>

              {/* Blocked Alert if Prohibited */}
              {result.evaluation.is_blocked && (
                <div className="p-4 rounded-lg bg-red-950/80 border border-red-500/60 text-red-200 text-xs">
                  <div className="font-bold flex items-center gap-2 mb-1">
                    <AlertOctagon className="w-4 h-4 text-red-400" />
                    Automated Block Triggered (Article 5 Prohibited)
                  </div>
                  <div>{result.evaluation.blocked_reason}</div>
                </div>
              )}

              {/* Applicable Articles */}
              <div>
                <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                  Applicable Legal Articles
                </h4>
                <div className="flex flex-wrap gap-2">
                  {result.evaluation.applicable_articles.map((art, i) => (
                    <span
                      key={i}
                      className="text-xs font-mono px-2.5 py-1 rounded bg-slate-950 text-emerald-400 border border-slate-800"
                    >
                      {art}
                    </span>
                  ))}
                </div>
              </div>

              {/* Rationale & Reasons */}
              <div>
                <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                  Audit Findings & Justification
                </h4>
                <ul className="space-y-1.5 text-xs text-slate-300">
                  {result.evaluation.reasons.map((r, i) => (
                    <li key={i} className="flex items-start gap-2 bg-slate-950/60 p-2.5 rounded-lg border border-slate-800">
                      <span className="text-emerald-400 font-bold">•</span>
                      <span>{r}</span>
                    </li>
                  ))}
                </ul>
              </div>

              {/* Mandatory Controls */}
              <div>
                <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                  Mandatory Compliance Measures
                </h4>
                <div className="space-y-1.5">
                  {result.evaluation.required_controls.map((ctrl, i) => (
                    <div key={i} className="flex items-center gap-2 text-xs text-slate-200 bg-slate-950/40 p-2 rounded border border-slate-800/80">
                      <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                      <span>{ctrl}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Prompt Injection Scanner Result */}
              {result.injection && (
                <div className="pt-3 border-t border-slate-800 text-xs flex items-center justify-between">
                  <span className="text-slate-400">Adversarial Injection Scan:</span>
                  {result.injection.detected ? (
                    <span className="text-red-400 font-semibold flex items-center gap-1">
                      <ShieldAlert className="w-3.5 h-3.5" /> Pattern Detected ({result.injection.pattern})
                    </span>
                  ) : (
                    <span className="text-emerald-400 font-semibold flex items-center gap-1">
                      <CheckCircle2 className="w-3.5 h-3.5" /> Clean (No Injection)
                    </span>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
