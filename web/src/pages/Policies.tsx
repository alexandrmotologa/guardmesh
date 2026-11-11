import React from 'react';
import { ShieldAlert, AlertTriangle, AlertCircle, CheckCircle2, Lock, Sliders, Bell } from 'lucide-react';
import { RiskBadge } from '../components/RiskBadge';

export const Policies: React.FC = () => {
  const piiDetectors = [
    { name: 'Credit Cards (Visa, MC, Amex, Discover)', algorithm: 'Luhn Checksum Algorithm', active: true, category: 'CREDIT_CARD' },
    { name: 'Bank Accounts (IBAN)', algorithm: 'ISO 7064 Mod-97-10 Checksum', active: true, category: 'IBAN' },
    { name: 'Romanian Personal Numeric Code (CNP)', algorithm: 'Constant Key 279146358279 Validation', active: true, category: 'ROMANIAN_CNP' },
    { name: 'Spanish Tax ID (DNI / NIF / NIE)', algorithm: 'Modulo 23 Letter Lookup Verification', active: true, category: 'SPANISH_NIF' },
    { name: 'US Social Security Numbers (SSN)', algorithm: 'Area / Group / Serial Number Sanitizer', active: true, category: 'SSN' },
    { name: 'Developer Secrets & API Keys', algorithm: 'OpenAI, GitHub, AWS, JWT, Private Keys', active: true, category: 'API_SECRET' },
    { name: 'Email Addresses', algorithm: 'RFC 5322 Compliant Regex Matcher', active: true, category: 'EMAIL' },
    { name: 'Phone Numbers', algorithm: 'International E.164 and Standard Formats', active: true, category: 'PHONE' },
    { name: 'IP Addresses', algorithm: 'Non-Loopback IPv4 and IPv6 Validator', active: true, category: 'IP_ADDRESS' },
    { name: 'Person Names', algorithm: 'Honorific Titles & Conversational Intro Parser', active: true, category: 'PERSON_NAME' },
    { name: 'Custom Enterprise Employee Badge', algorithm: 'EMP-[0-9]{6} Regex Pattern', active: true, category: 'EMPLOYEE_ID' },
  ];

  const prohibitedRules = [
    { article: 'Article 5(1)(a)', title: 'Subliminal or Manipulative Behavioral Distortion', action: 'Hard Block (HTTP 403)' },
    { article: 'Article 5(1)(c)', title: 'Social Scoring or Citizen Trustworthiness Rating', action: 'Hard Block (HTTP 403)' },
    { article: 'Article 5(1)(d)', title: 'Predictive Policing Based Solely on Personality Profiling', action: 'Hard Block (HTTP 403)' },
    { article: 'Article 5(1)(e)', title: 'Untargeted Facial Image Scraping from Web / CCTV', action: 'Hard Block (HTTP 403)' },
    { article: 'Article 5(1)(f)', title: 'Emotion Recognition in Workplaces and Educational Institutions', action: 'Hard Block (HTTP 403)' },
    { article: 'Article 5(1)(g)', title: 'Biometric Categorization Deducing Sensitive Protected Traits', action: 'Hard Block (HTTP 403)' },
  ];

  const highRiskCategories = [
    { annex: 'Annex III(4)(a)', domain: 'Recruitment & CV Screening', control: 'Article 14 Human Oversight + Article 11 Dossier' },
    { annex: 'Annex III(4)(b)', domain: 'Worker Management & Performance Evaluation', control: 'Article 9 Risk Management System' },
    { annex: 'Annex III(5)(b)', domain: 'Credit Scoring & Creditworthiness Assessment', control: 'Article 10 Data Quality Governance' },
    { annex: 'Annex III(5)(c)', domain: 'Risk Evaluation for Life & Health Insurance', control: 'Bias & Fairness Testing' },
    { annex: 'Annex III(5)(d)', domain: 'Emergency Services Dispatch Prioritization', control: 'Article 15 Accuracy & Cybersecurity' },
  ];

  return (
    <div className="space-y-8 animate-fadeIn">
      {/* Header */}
      <div className="bg-slate-900/90 border border-slate-800 p-6 rounded-xl">
        <h2 className="text-xl font-bold text-white">Governance Policies & Enterprise DLP Rules</h2>
        <p className="text-xs text-slate-400 mt-1">
          Active regulatory thresholds, statutory compliance policies, and DLP entity detectors enforced by guardmesh
        </p>
      </div>

      {/* Grid: PII Detectors & Governance Rules */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Active PII Detectors */}
        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-6 space-y-4">
          <div className="flex items-center justify-between pb-3 border-b border-slate-800">
            <div className="flex items-center gap-2">
              <Lock className="w-4 h-4 text-emerald-400" />
              <h3 className="text-sm font-bold text-white uppercase tracking-wider">Active DLP Entity Detectors</h3>
            </div>
            <span className="text-xs font-mono text-emerald-400 font-bold">{piiDetectors.length} Active</span>
          </div>

          <div className="space-y-2.5 max-h-[460px] overflow-y-auto pr-1">
            {piiDetectors.map((det, i) => (
              <div key={i} className="bg-slate-950 p-3 rounded-lg border border-slate-800 flex items-center justify-between">
                <div>
                  <div className="text-xs font-bold text-slate-200">{det.name}</div>
                  <div className="text-[11px] text-slate-400 font-mono mt-0.5">{det.algorithm}</div>
                </div>
                <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-semibold">
                  {det.category}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Article 5 Prohibited Enforcement */}
        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-6 space-y-4 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-4">
              <div className="flex items-center gap-2">
                <ShieldAlert className="w-4 h-4 text-red-400" />
                <h3 className="text-sm font-bold text-white uppercase tracking-wider">Article 5 Prohibited AI Rules</h3>
              </div>
              <span className="text-xs font-mono text-red-400 font-bold">Hard Kill: Active</span>
            </div>

            <div className="space-y-2.5">
              {prohibitedRules.map((rule, i) => (
                <div key={i} className="bg-slate-950 p-3 rounded-lg border border-red-500/30 flex items-center justify-between">
                  <div>
                    <div className="text-xs font-bold text-slate-200">{rule.title}</div>
                    <div className="text-[11px] text-red-400 font-mono mt-0.5">{rule.article}</div>
                  </div>
                  <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-red-950/80 text-red-400 border border-red-500/40">
                    {rule.action}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Article 50 Transparency Box */}
          <div className="mt-4 pt-4 border-t border-slate-800 bg-blue-950/20 p-4 rounded-xl border border-blue-500/30">
            <div className="flex items-center gap-2 mb-1">
              <AlertCircle className="w-4 h-4 text-blue-400" />
              <span className="text-xs font-bold text-blue-300">Article 50 Transparency Enforcer</span>
            </div>
            <p className="text-[11px] text-slate-300 leading-relaxed">
              When workloads match Limited Risk (conversational chatbots), guardmesh automatically injects HTTP header <code>X-AI-Generated: true</code> and statutory disclosure disclaimers into generated messages.
            </p>
          </div>
        </div>
      </div>

      {/* Annex III High Risk Domains */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-6">
        <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-4">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-amber-400" />
            <h3 className="text-sm font-bold text-white uppercase tracking-wider">Annex III High-Risk Classification Triggers</h3>
          </div>
          <span className="text-xs text-slate-400 font-mono">Article 6 Mandatory Telemetry</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {highRiskCategories.map((cat, i) => (
            <div key={i} className="bg-slate-950 p-3.5 rounded-lg border border-slate-800 flex flex-col justify-between">
              <div>
                <span className="text-[10px] font-mono font-bold text-amber-400 block mb-1">{cat.annex}</span>
                <span className="text-xs font-bold text-slate-200 block">{cat.domain}</span>
              </div>
              <div className="mt-3 pt-2 border-t border-slate-800/80 text-[11px] text-slate-400">
                {cat.control}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
