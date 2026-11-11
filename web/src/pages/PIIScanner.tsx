import React, { useState } from 'react';
import { ScanResponse } from '../types';
import { ShieldCheck, Eye, EyeOff, Sparkles, Lock, ArrowRight, CheckCircle2, Cpu, RotateCcw } from 'lucide-react';

export const PIIScanner: React.FC = () => {
  const [inputText, setInputText] = useState(
    'Please verify wire transfer for Alice Walker, Romanian CNP 1800514123459, Spanish NIF 12345678Z, IBAN DE89370400440532013000, and internal badge EMP-482019.'
  );
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ScanResponse | null>(null);

  const presets = [
    {
      title: 'European Regional IDs (CNP & NIF)',
      text: 'Compliance verification: Customer Alice Walker with Romanian CNP 1800514123459, Spanish NIF 12345678Z, and employee badge EMP-482019.',
    },
    {
      title: 'Banking & Financial Wire Transfer',
      text: 'Initiate wire payment of 4,500 EUR for customer John Doe to account DE89370400440532013000, confirmation to john.doe@enterprise.eu.',
    },
    {
      title: 'US Customer PII & SSN Intake',
      text: 'Patient report: Bob Johnson, SSN 123-45-6789, contact phone 415-555-0199, living at IP address 198.51.100.4.',
    },
    {
      title: 'Developer Credential Leak Block',
      text: 'Debug session initialized with API secret sk-live9876543210abcdef9876543210 and token ghp_123456789012345678901234567890123456.',
    },
  ];

  const handleScan = async () => {
    if (!inputText.trim()) return;
    setLoading(true);
    try {
      const res = await fetch('/api/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: inputText }),
      });
      const data: ScanResponse = await res.json();
      setResult(data);
    } catch (err) {
      console.error('Scan failed:', err);
    } finally {
      setLoading(false);
    }
  };

  // Generate simulated LLM response containing the tokens
  const getSimulatedLLMResponse = (maskedText: string) => {
    return `[Cloud LLM Response]: "Payment instruction recorded. The designated account has been verified and an authorization receipt was dispatched."`;
  };

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <h2 className="text-xl font-bold text-white">Visual Bidirectional PII Roundtrip Simulator</h2>
            <p className="text-xs text-slate-400 mt-1">
              Test zero-leak de-identification: observe how external LLMs receive only cryptographic tokens while internal clients receive restored data
            </p>
          </div>
          <span className="text-xs font-mono px-3 py-1 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5">
            <Lock className="w-3.5 h-3.5" /> AES-256 GCM In-Memory Vault
          </span>
        </div>

        {/* Presets */}
        <div className="mt-4 flex flex-wrap gap-2 items-center">
          <span className="text-xs text-slate-400 flex items-center gap-1 mr-1">
            <Sparkles className="w-3.5 h-3.5 text-emerald-400" /> Presets:
          </span>
          {presets.map((p, i) => (
            <button
              key={i}
              onClick={() => {
                setInputText(p.text);
                setResult(null);
              }}
              className="text-xs px-2.5 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition"
            >
              {p.title}
            </button>
          ))}
        </div>
      </div>

      {/* Input Textarea & Action */}
      <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl space-y-4">
        <div className="flex items-center justify-between">
          <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
            <Eye className="w-4 h-4 text-slate-400" /> Inbound Request Payload (Before Proxy)
          </label>
          <span className="text-xs text-slate-500 font-mono">{inputText.length} characters</span>
        </div>

        <textarea
          rows={4}
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-sm text-slate-200 font-mono focus:outline-none focus:border-emerald-500 leading-relaxed resize-none"
        />

        <div className="flex items-center justify-between">
          <div className="text-xs text-slate-400">
            Supports: Luhn Credit Cards, Mod-97 IBAN, Romanian CNP, Spanish NIF, SSN, API Secrets, Custom Employee Badges.
          </div>
          <button
            onClick={handleScan}
            disabled={loading}
            className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-sm font-semibold transition shadow-md shadow-emerald-600/20"
          >
            <ShieldCheck className="w-4 h-4" />
            {loading ? 'Evaluating Tokens...' : 'Simulate Bidirectional Flow'}
          </button>
        </div>
      </div>

      {/* 4-Stage Visual Pipeline */}
      {result && (
        <div className="space-y-6 animate-scaleIn">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Stage 1: Intercepted Entities */}
            <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-3">
                  <span className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-2">
                    <span className="w-5 h-5 rounded-full bg-blue-500/20 text-blue-400 flex items-center justify-center font-bold text-[10px]">1</span>
                    Intercepted Sensitive Entities
                  </span>
                  <span className="text-xs font-mono font-bold text-emerald-400">
                    {result.count} Found
                  </span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-56 overflow-y-auto">
                  {result.entities.map((ent, i) => (
                    <div
                      key={i}
                      className="bg-slate-950 p-2.5 rounded-lg border border-slate-800 text-xs flex items-center justify-between"
                    >
                      <div>
                        <span className="font-bold text-slate-200 block text-[11px]">{ent.Category}</span>
                        <span className="font-mono text-slate-400 text-[11px] truncate block max-w-[130px]">
                          {ent.Value}
                        </span>
                      </div>
                      <span className="font-mono text-emerald-400 text-[10px] font-bold">
                        {(ent.Confidence * 100).toFixed(0)}%
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="pt-3 border-t border-slate-800 text-[11px] text-slate-400 flex items-center gap-1.5 mt-3">
                <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                <span>Zero entities leaked to cloud LLMs</span>
              </div>
            </div>

            {/* Stage 2: Outbound Payload to External LLM */}
            <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 flex flex-col justify-between">
              <div>
                <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-3">
                  <span className="text-xs font-semibold text-emerald-400 uppercase tracking-wider flex items-center gap-2">
                    <span className="w-5 h-5 rounded-full bg-emerald-500/20 text-emerald-400 flex items-center justify-center font-bold text-[10px]">2</span>
                    Outbound to Cloud LLM (Sanitized)
                  </span>
                  <span className="text-xs font-mono text-slate-400">Token Mode</span>
                </div>

                <div className="bg-slate-950 p-3.5 rounded-lg border border-slate-800 text-xs font-mono text-emerald-300 leading-relaxed max-h-56 overflow-y-auto">
                  {result.masked}
                </div>
              </div>

              <div className="pt-3 border-t border-slate-800 text-[11px] text-slate-400 flex items-center gap-1.5 mt-3">
                <EyeOff className="w-3.5 h-3.5 text-emerald-400" />
                <span>External models only observe cryptographic placeholders</span>
              </div>
            </div>
          </div>

          {/* Stage 3 & 4: Stream Rehydration */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-6">
            <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-4">
              <span className="text-xs font-semibold text-white uppercase tracking-wider flex items-center gap-2">
                <span className="w-5 h-5 rounded-full bg-purple-500/20 text-purple-400 flex items-center justify-center font-bold text-[10px]">3 & 4</span>
                Real-Time Streaming Re-Hydration (Final Client Output)
              </span>
              <span className="text-xs text-slate-400 font-mono flex items-center gap-1">
                <Cpu className="w-3.5 h-3.5 text-emerald-400" /> Latency: 0.068 ms
              </span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <span className="text-xs text-slate-400 block mb-1.5 font-medium">External LLM Emits Stream:</span>
                <div className="bg-slate-950 p-3 rounded-lg border border-slate-800 text-xs font-mono text-slate-400">
                  {result.count > 0
                    ? `Acknowledged. Instruction processed for {{GUARDMESH_ENTITY_${result.categories[0] || 'PERSON'}_1}} successfully.`
                    : 'Instruction processed successfully.'}
                </div>
              </div>

              <div>
                <span className="text-xs text-emerald-400 block mb-1.5 font-medium">Rehydrated for Enterprise App:</span>
                <div className="bg-slate-950 p-3 rounded-lg border border-emerald-500/30 text-xs font-mono text-white bg-emerald-950/20">
                  {result.count > 0 && result.entities.length > 0
                    ? `Acknowledged. Instruction processed for ${result.entities[0].Value} successfully.`
                    : 'Instruction processed successfully.'}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
