import React, { useState } from 'react';
import { ScanResponse } from '../types';
import { ShieldCheck, Eye, EyeOff, Sparkles, Lock, Key, Mail, CreditCard, Phone } from 'lucide-react';

export const PIIScanner: React.FC = () => {
  const [inputText, setInputText] = useState(
    'Please transfer $1,250 to customer John Doe at IBAN DE89370400440532013000, email john.doe@acme.com, phone +1-555-234-5678, SSN 123-45-6789. Also note developer key sk-abcdef1234567890abcdef1234567890.'
  );
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ScanResponse | null>(null);

  const presets = [
    {
      title: 'Banking & Financial Transfer',
      text: 'Initiate wire payment of 4,500 EUR for customer Alice Walker to account DE89370400440532013000, confirmation to alice.w@banking.eu.',
    },
    {
      title: 'Customer PII & SSN Intake',
      text: 'Patient report: Bob Johnson, SSN 123-45-6789, contact phone 415-555-0199, living at IP address 198.51.100.4.',
    },
    {
      title: 'Developer Credential Leak',
      text: 'Debugging log: client initialized with API key sk-live9876543210abcdef9876543210 and token ghp_123456789012345678901234567890123456.',
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

  return (
    <div className="space-y-6 animate-fadeIn">
      {/* Header */}
      <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl">
        <h2 className="text-xl font-bold text-white">Bidirectional Zero-Leak PII De-Identifier</h2>
        <p className="text-xs text-slate-400 mt-1">
          Verify real-time cryptographic token replacement before payload egress to public LLMs (OpenAI, Anthropic)
        </p>

        {/* Presets */}
        <div className="mt-4 flex flex-wrap gap-2">
          <span className="text-xs text-slate-400 flex items-center gap-1 mr-1">
            <Sparkles className="w-3.5 h-3.5 text-emerald-400" /> Presets:
          </span>
          {presets.map((p, i) => (
            <button
              key={i}
              onClick={() => setInputText(p.text)}
              className="text-xs px-2.5 py-1 rounded-md bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition"
            >
              {p.title}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Input */}
        <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl space-y-4">
          <div className="flex items-center justify-between">
            <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
              <Eye className="w-4 h-4 text-slate-400" /> Raw Inbound Payload
            </label>
            <span className="text-xs text-slate-500 font-mono">{inputText.length} characters</span>
          </div>

          <textarea
            rows={8}
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-sm text-slate-200 font-mono focus:outline-none focus:border-emerald-500 leading-relaxed resize-none"
          />

          <button
            onClick={handleScan}
            disabled={loading}
            className="w-full inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-sm font-semibold transition shadow-md shadow-emerald-600/20"
          >
            <ShieldCheck className="w-4 h-4" />
            {loading ? 'Analyzing Entities...' : 'De-Identify & Generate Tokens'}
          </button>
        </div>

        {/* Sanitized Output */}
        <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-4">
              <label className="text-xs font-semibold text-emerald-400 uppercase tracking-wider flex items-center gap-1.5">
                <EyeOff className="w-4 h-4 text-emerald-400" /> Sanitized Outbound (Sent to Cloud LLM)
              </label>
              {result && (
                <span className="text-xs font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {result.count} Entities Masked
                </span>
              )}
            </div>

            {!result ? (
              <div className="py-12 text-center text-slate-500 text-sm">
                Click "De-Identify & Generate Tokens" to preview sanitized payload.
              </div>
            ) : (
              <div className="space-y-4">
                <div className="bg-slate-950 p-4 rounded-xl border border-slate-800 text-sm font-mono text-emerald-300 leading-relaxed overflow-x-auto whitespace-pre-wrap">
                  {result.masked}
                </div>

                <div>
                  <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
                    Intercepted Sensitive Entities
                  </h4>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 max-h-48 overflow-y-auto">
                    {result.entities.map((ent, i) => (
                      <div
                        key={i}
                        className="bg-slate-950/80 p-2.5 rounded-lg border border-slate-800/80 text-xs flex items-center justify-between"
                      >
                        <div>
                          <span className="font-bold text-slate-200 block">{ent.Category}</span>
                          <span className="font-mono text-slate-400 text-[11px] truncate block max-w-[140px]">
                            {ent.Value}
                          </span>
                        </div>
                        <span className="font-mono text-emerald-400 text-[11px]">
                          {(ent.Confidence * 100).toFixed(0)}%
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
          </div>

          <div className="pt-4 border-t border-slate-800 mt-4 text-[11px] text-slate-400 flex items-center justify-between">
            <span className="flex items-center gap-1">
              <Lock className="w-3.5 h-3.5 text-emerald-400" /> Vault Key: In-Memory AES-256 GCM
            </span>
            <span>TTL: 300s Auto-Expiry</span>
          </div>
        </div>
      </div>
    </div>
  );
};
